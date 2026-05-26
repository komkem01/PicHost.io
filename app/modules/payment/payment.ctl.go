package payment

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"pichost.io/app/modules/entities/ent"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	tracer trace.Tracer
	svc    *Service
}

func newController(tracer trace.Tracer, svc *Service) *Controller {
	return &Controller{tracer: tracer, svc: svc}
}

type checkoutRequest struct {
	PlanKey string `json:"plan_key" binding:"required"`
}

type submitSlipRequest struct {
	StorageID string `json:"storage_id" binding:"required"`
}

type cancelSubscriptionRequest struct {
	UseUntilMonth *string `json:"use_until_month"`
}

func (c *Controller) CreateCheckout(ctx *gin.Context) {
	userID, ok := ctx.Get("auth_user_id")
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	authUserID, ok := userID.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	var req checkoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	result, err := c.svc.CreateCheckout(ctx.Request.Context(), CreateCheckoutInput{
		UserID:  authUserID,
		PlanKey: req.PlanKey,
	})
	if err != nil {
		if errors.Is(err, ErrPaymentPlanUnavailable) {
			base.BadRequest(ctx, err.Error(), nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, result)
}

func (c *Controller) GetMyPayment(ctx *gin.Context) {
	userID, ok := ctx.Get("auth_user_id")
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	authUserID, ok := userID.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	paymentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	result, err := c.svc.GetMyPayment(ctx.Request.Context(), authUserID, paymentID)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			base.BadRequest(ctx, err.Error(), nil)
		case errors.Is(err, ErrPaymentForbidden):
			base.Forbidden(ctx, i18n.Forbidden, nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	base.Success(ctx, result)
}

func (c *Controller) ListMyPayments(ctx *gin.Context) {
	userID, ok := ctx.Get("auth_user_id")
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	authUserID, ok := userID.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	limit := 20
	if raw := strings.TrimSpace(ctx.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
			return
		}
		limit = parsed
	}

	rows, err := c.svc.ListMyPayments(ctx.Request.Context(), authUserID, limit)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, rows)
}

type webhookConfirmRequest struct {
	PaymentID         *uuid.UUID     `json:"payment_id"`
	CheckoutReference *string        `json:"checkout_reference"`
	ProviderReference *string        `json:"provider_reference"`
	Status            string         `json:"status" binding:"required"`
	PaidAmountTHB     *int           `json:"paid_amount_thb"`
	Metadata          map[string]any `json:"metadata"`
}

func (c *Controller) ConfirmPaymentWebhook(ctx *gin.Context) {
	if strings.TrimSpace(c.svc.Val.WebhookSecret) != "" {
		if !strings.EqualFold(strings.TrimSpace(ctx.GetHeader("X-Payment-Webhook-Token")), strings.TrimSpace(c.svc.Val.WebhookSecret)) {
			base.Unauthorized(ctx, i18n.Unauthorized, gin.H{"error": ErrPaymentWebhookDenied.Error()})
			return
		}
	}

	var req webhookConfirmRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	status := ent.PaymentStatus(strings.ToLower(strings.TrimSpace(req.Status)))

	updated, upgraded, err := c.svc.ConfirmPayment(ctx.Request.Context(), ConfirmPaymentInput{
		PaymentID:         req.PaymentID,
		CheckoutReference: req.CheckoutReference,
		ProviderReference: req.ProviderReference,
		Status:            status,
		PaidAmountTHB:     req.PaidAmountTHB,
		Metadata:          req.Metadata,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound), errors.Is(err, ErrPaymentInvalidStatus), errors.Is(err, ErrPaymentInvalidAmount):
			base.BadRequest(ctx, err.Error(), nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	base.Success(ctx, gin.H{
		"payment":          updated,
		"is_plan_upgraded": upgraded,
	})
}

// SubmitSlip handles POST /billing/payments/:id/slip
// Accepts a JSON body with the storage_id returned by the external storages service.
func (c *Controller) SubmitSlip(ctx *gin.Context) {
	userID, ok := ctx.Get("auth_user_id")
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	authUserID, ok := userID.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	paymentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	var req submitSlipRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	result, err := c.svc.SubmitSlip(ctx.Request.Context(), SubmitSlipInput{
		UserID:    authUserID,
		PaymentID: paymentID,
		StorageID: req.StorageID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			base.BadRequest(ctx, err.Error(), nil)
		case errors.Is(err, ErrPaymentForbidden):
			base.Forbidden(ctx, i18n.Forbidden, nil)
		case errors.Is(err, ErrPaymentSlipAlreadySubmitted):
			base.BadRequest(ctx, err.Error(), nil)
		case errors.Is(err, ErrPaymentNotPending):
			base.BadRequest(ctx, err.Error(), nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	base.Success(ctx, result)
}

// GetPaymentMethods handles GET /billing/payment-methods
// Returns the bank account / PromptPay details configured via environment variables.
func (c *Controller) GetPaymentMethods(ctx *gin.Context) {
	base.Success(ctx, gin.H{
		"bank_name":           c.svc.Val.BankName,
		"bank_account_name":   c.svc.Val.BankAccountName,
		"bank_account_number": c.svc.Val.BankAccountNumber,
		"bank_account_type":   c.svc.Val.BankAccountType,
		"prompt_pay_id":       c.svc.Val.PromptPayID,
		"bank_logo_url":       c.svc.Val.BankLogoURL,
	})
}

// AdminListPayments handles GET /admin/payments
// Returns all payment transactions, newest first, with pagination.
func (c *Controller) AdminListPayments(ctx *gin.Context) {
	limit := 50
	offset := 0
	if raw := strings.TrimSpace(ctx.Query("limit")); raw != "" {
		if v, e := strconv.Atoi(raw); e == nil && v > 0 {
			limit = v
		}
	}
	if raw := strings.TrimSpace(ctx.Query("offset")); raw != "" {
		if v, e := strconv.Atoi(raw); e == nil && v >= 0 {
			offset = v
		}
	}

	rows, total, err := c.svc.AdminListPayments(ctx.Request.Context(), limit, offset)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	page := int64(offset/limit) + 1
	_ = base.Paginate(ctx, rows, &base.ResponsePaginate{
		Page:  page,
		Size:  int64(limit),
		Total: int64(total),
	})
}

type adminConfirmRequest struct {
	ProviderReference *string        `json:"provider_reference"`
	Status            string         `json:"status" binding:"required"`
	PaidAmountTHB     *int           `json:"paid_amount_thb"`
	ReviewReason      *string        `json:"review_reason"`
	Metadata          map[string]any `json:"metadata"`
}

// AdminConfirmPayment handles PATCH /admin/payments/:id/confirm
// Allows an admin to manually confirm or reject a payment transaction.
func (c *Controller) AdminConfirmPayment(ctx *gin.Context) {
	rawUserID, ok := ctx.Get("auth_user_id")
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	authUserID, ok := rawUserID.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	paymentID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	var req adminConfirmRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	status := ent.PaymentStatus(strings.ToLower(strings.TrimSpace(req.Status)))

	updated, upgraded, err := c.svc.ConfirmPayment(ctx.Request.Context(), ConfirmPaymentInput{
		PaymentID:         &paymentID,
		ProviderReference: req.ProviderReference,
		Status:            status,
		PaidAmountTHB:     req.PaidAmountTHB,
		ReviewReason:      req.ReviewReason,
		ReviewedBy:        &authUserID,
		Metadata:          req.Metadata,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound), errors.Is(err, ErrPaymentInvalidStatus), errors.Is(err, ErrPaymentInvalidAmount), errors.Is(err, ErrPaymentReviewReasonRequired):
			base.BadRequest(ctx, err.Error(), nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	base.Success(ctx, gin.H{
		"payment":          updated,
		"is_plan_upgraded": upgraded,
	})
}

// CancelSubscription handles POST /billing/cancel — marks the subscription as cancelled.
// The plan remains active until plan_expires_at; it downgrades to Free automatically after expiry.
func (c *Controller) CancelSubscription(ctx *gin.Context) {
	var req cancelSubscriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	userIDValue, exists := ctx.Get("auth_user_id")
	if !exists {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	user, err := c.svc.CancelSubscription(ctx.Request.Context(), CancelSubscriptionInput{
		UserID:        userID,
		UseUntilMonth: req.UseUntilMonth,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrSubscriptionNotActive):
			base.BadRequest(ctx, "No active paid subscription to cancel.", nil)
		case errors.Is(err, ErrSubscriptionAlreadyCancelled):
			base.BadRequest(ctx, "Subscription is already cancelled.", nil)
		case errors.Is(err, ErrSubscriptionUseUntilRequired):
			base.BadRequest(ctx, "Please select the month you want to use the plan until.", nil)
		case errors.Is(err, ErrSubscriptionInvalidUseUntil):
			base.BadRequest(ctx, "Invalid month format. Please use YYYY-MM.", nil)
		case errors.Is(err, ErrSubscriptionUseUntilInPast):
			base.BadRequest(ctx, "Selected month cannot be in the past.", nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}
	_ = user // updated successfully; client should re-fetch /auth/me for latest state
	base.Success(ctx, nil, "Subscription cancelled successfully.")
}
