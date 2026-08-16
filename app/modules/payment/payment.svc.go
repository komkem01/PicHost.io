package payment

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
	mailerinf "pichost.io/app/modules/mailer/inf"
	"pichost.io/internal/config"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Options struct {
	*config.Config[Config]
	tracer     trace.Tracer
	userEnt    entitiesinf.UserEntity
	planEnt    entitiesinf.PlanSettingEntity
	paymentEnt entitiesinf.PaymentTransactionEntity
}

type Service struct {
	*Options
	mailer   mailerinf.Mailer
	notifEnt entitiesinf.NotificationEntity
}

func newService(opt *Options) *Service {
	if opt.Val.CheckoutTTLMinutes <= 0 {
		opt.Val.CheckoutTTLMinutes = 15
	}
	if opt.Val.SubscriptionDays <= 0 {
		opt.Val.SubscriptionDays = 30
	}
	return &Service{Options: opt}
}

func normalizePlanKey(planKey string) string {
	return strings.ToLower(strings.TrimSpace(planKey))
}

func (s *Service) mapPlanKeyToUserPlan(planKey string) (string, error) {
	switch normalizePlanKey(planKey) {
	case "free":
		return string(ent.PlanTypeFree), nil
	case "basic":
		return string(ent.PlanTypeBasic), nil
	case "pro":
		return string(ent.PlanTypePro), nil
	case "enterprise":
		return string(ent.PlanTypeEnterprise), nil
	default:
		return "", ErrPaymentPlanUnavailable
	}
}

func (s *Service) buildCheckoutURL(checkoutReference string) *string {
	baseURL := strings.TrimSpace(s.Val.CheckoutBaseURL)
	if baseURL == "" {
		return nil
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := fmt.Sprintf("%s/pay/%s", baseURL, checkoutReference)
	return &url
}

type CreateCheckoutInput struct {
	UserID   uuid.UUID
	PlanKey  string
	Provider string
}

func (s *Service) CreateCheckout(ctx context.Context, in CreateCheckoutInput) (*ent.PaymentTransactionEntity, error) {
	user, err := s.userEnt.GetUserByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if user.Email != nil && user.EmailVerifiedAt == nil {
		return nil, ErrEmailVerificationRequired
	}

	planKey := normalizePlanKey(in.PlanKey)
	if planKey == "" {
		return nil, ErrPaymentPlanUnavailable
	}

	plan, err := s.planEnt.GetPlanSettingByKey(ctx, planKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentPlanUnavailable
		}
		return nil, err
	}
	if !plan.IsEnabled {
		return nil, ErrPaymentPlanUnavailable
	}

	checkoutReference := strings.ToUpper(uuid.NewString())
	expiresAt := time.Now().Add(time.Duration(s.Val.CheckoutTTLMinutes) * time.Minute)

	provider := strings.ToLower(strings.TrimSpace(in.Provider))
	if provider == "" {
		provider = string(ProviderManual)
	}

	row, err := s.paymentEnt.CreatePaymentTransaction(ctx, entitiesdto.CreatePaymentTransaction{
		UserID:            in.UserID,
		PlanKey:           planKey,
		AmountTHB:         plan.MonthlyPriceTHB,
		Currency:          "THB",
		Status:            ent.PaymentStatusPending,
		Provider:          provider,
		CheckoutReference: checkoutReference,
		PaymentURL:        s.buildCheckoutURL(checkoutReference),
		ExpiresAt:         expiresAt,
		Metadata: map[string]any{
			"plan_display_name": plan.DisplayName,
		},
	})
	if err != nil {
		return nil, err
	}
	return row, nil
}

func (s *Service) GetMyPayment(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID) (*ent.PaymentTransactionEntity, error) {
	row, err := s.paymentEnt.GetPaymentTransactionByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if row.UserID != userID {
		return nil, ErrPaymentForbidden
	}
	if row.Status == ent.PaymentStatusPending && (row.SlipStorageID == nil || strings.TrimSpace(*row.SlipStorageID) == "") && time.Now().After(row.ExpiresAt) {
		row, err = s.paymentEnt.UpdatePaymentTransactionStatus(ctx, row.ID, entitiesdto.UpdatePaymentTransactionStatus{
			Status: ent.PaymentStatusExpired,
		})
		if err != nil {
			return nil, err
		}
	}
	return row, nil
}

func (s *Service) ListMyPayments(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.PaymentTransactionEntity, error) {
	rows, err := s.paymentEnt.ListPaymentTransactionsByUserID(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type ConfirmPaymentInput struct {
	PaymentID         *uuid.UUID
	CheckoutReference *string
	ProviderReference *string
	Status            ent.PaymentStatus
	PaidAmountTHB     *int
	ReviewReason      *string
	ReviewedBy        *uuid.UUID
	Metadata          map[string]any
}

func (s *Service) ConfirmPayment(ctx context.Context, in ConfirmPaymentInput) (*ent.PaymentTransactionEntity, bool, error) {
	var row *ent.PaymentTransactionEntity
	var err error

	switch {
	case in.PaymentID != nil:
		row, err = s.paymentEnt.GetPaymentTransactionByID(ctx, *in.PaymentID)
	case in.CheckoutReference != nil:
		row, err = s.paymentEnt.GetPaymentTransactionByCheckoutReference(ctx, *in.CheckoutReference)
	default:
		return nil, false, ErrPaymentNotFound
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, ErrPaymentNotFound
		}
		return nil, false, err
	}

	if row.Status == ent.PaymentStatusPending && (row.SlipStorageID == nil || strings.TrimSpace(*row.SlipStorageID) == "") && time.Now().After(row.ExpiresAt) {
		row, err = s.paymentEnt.UpdatePaymentTransactionStatus(ctx, row.ID, entitiesdto.UpdatePaymentTransactionStatus{Status: ent.PaymentStatusExpired})
		if err != nil {
			return nil, false, err
		}
	}

	if row.Status == ent.PaymentStatusPaid {
		return row, false, nil
	}
	if row.Status == ent.PaymentStatusCancelled || row.Status == ent.PaymentStatusFailed || row.Status == ent.PaymentStatusExpired {
		return row, false, nil
	}

	nextStatus := in.Status
	switch nextStatus {
	case ent.PaymentStatusPaid, ent.PaymentStatusFailed, ent.PaymentStatusCancelled, ent.PaymentStatusExpired:
	default:
		return nil, false, ErrPaymentInvalidStatus
	}

	if in.ReviewedBy != nil && (nextStatus == ent.PaymentStatusFailed || nextStatus == ent.PaymentStatusCancelled) {
		if in.ReviewReason == nil || strings.TrimSpace(*in.ReviewReason) == "" {
			return nil, false, ErrPaymentReviewReasonRequired
		}
	}

	var paidAt *time.Time
	if nextStatus == ent.PaymentStatusPaid {
		if in.PaidAmountTHB != nil && *in.PaidAmountTHB != row.AmountTHB {
			return nil, false, ErrPaymentInvalidAmount
		}
		now := time.Now()
		paidAt = &now
	}

	var reviewReason *string
	if in.ReviewReason != nil {
		trimmed := strings.TrimSpace(*in.ReviewReason)
		if trimmed != "" {
			reviewReason = &trimmed
		}
	}

	var reviewedAt *time.Time
	if in.ReviewedBy != nil {
		now := time.Now()
		reviewedAt = &now
	}

	updated, err := s.paymentEnt.UpdatePaymentTransactionStatus(ctx, row.ID, entitiesdto.UpdatePaymentTransactionStatus{
		Status:            nextStatus,
		ProviderReference: in.ProviderReference,
		PaidAt:            paidAt,
		ReviewReason:      reviewReason,
		ReviewedBy:        in.ReviewedBy,
		ReviewedAt:        reviewedAt,
		Metadata:          in.Metadata,
	})
	if err != nil {
		return nil, false, err
	}

	targetUser, _ := s.userEnt.GetUserByID(ctx, updated.UserID)

	if nextStatus != ent.PaymentStatusPaid {
		if (nextStatus == ent.PaymentStatusFailed || nextStatus == ent.PaymentStatusCancelled) {
			reasonStr := ""
			if reviewReason != nil {
				reasonStr = *reviewReason
			}
			if s.mailer != nil && targetUser != nil && targetUser.Email != nil {
				_ = s.mailer.SendSlipRejected(ctx, *targetUser.Email, reasonStr, true)
			}
			if s.notifEnt != nil {
				link := "/billing/payments/" + updated.ID.String()
				msg := "สลิปการชำระเงินสำหรับแพ็กเกจ " + strings.ToUpper(updated.PlanKey) + " ไม่ผ่านการอนุมัติ"
				if reasonStr != "" {
					msg += " เนื่องจาก: " + reasonStr
				}
				_, _ = s.notifEnt.CreateNotification(ctx, entitiesdto.CreateNotification{
					UserID:     &updated.UserID,
					TargetRole: "user",
					Type:       "payment",
					Title:      "การชำระเงินไม่ผ่านการอนุมัติ",
					Message:    msg,
					Link:       &link,
				})
			}
		}
		return updated, false, nil
	}

	planValue, err := s.mapPlanKeyToUserPlan(updated.PlanKey)
	if err != nil {
		return nil, false, err
	}

	updatedUser, err := s.userEnt.UpdateUserPlan(ctx, updated.UserID, entitiesdto.UpdateUserPlan{Plan: &planValue})
	if err != nil {
		return nil, false, err
	}

	// Stack from existing expiry if still valid, otherwise start from now.
	subDuration := time.Duration(s.Val.SubscriptionDays) * 24 * time.Hour
	var newExpiresAt time.Time
	if updatedUser.PlanExpiresAt != nil && updatedUser.PlanExpiresAt.After(time.Now()) {
		newExpiresAt = updatedUser.PlanExpiresAt.Add(subDuration)
	} else {
		newExpiresAt = time.Now().Add(subDuration)
	}
	// Renewing clears any previous cancellation intent.
	if _, err = s.userEnt.SetUserPlanExpiry(ctx, updated.UserID, &newExpiresAt, true); err != nil {
		return nil, false, err
	}

	if s.mailer != nil && updatedUser.Email != nil {
		_ = s.mailer.SendSlipApproved(ctx, *updatedUser.Email, planValue, true)
	}

	// Dispatch In-App Notification to User
	if s.notifEnt != nil {
		link := "/billing/payments/" + updated.ID.String()
		_, _ = s.notifEnt.CreateNotification(ctx, entitiesdto.CreateNotification{
			UserID:     &updated.UserID,
			TargetRole: "user",
			Type:       "payment",
			Title:      "อนุมัติการชำระเงินเรียบร้อยแล้ว",
			Message:    "แพ็กเกจ " + strings.ToUpper(updated.PlanKey) + " ของคุณได้รับการเปิดใช้งานเรียบร้อยแล้ว ขอขอบคุณที่ใช้บริการ PicHost.io",
			Link:       &link,
		})
	}

	// Send telegram notification on plan purchase & confirmation
	msg := fmt.Sprintf("🎉 <b>Plan Purchased & Confirmed!</b>\n\n"+
		"<b>User ID :</b> %s\n"+
		"<b>Transaction ID 	:</b> %s\n"+
		"<b>Amount :</b> %d THB\n"+
		"<b>Plan Key :</b> %s\n"+
		"<b>Status :</b> PAID\n\n"+
		"<i>The user's plan has been upgraded successfully!</i>",
		updated.UserID.String(), updated.ID.String(), updated.AmountTHB, updated.PlanKey)
	sendTelegramNotification(msg)

	return updated, true, nil
}

type SubmitSlipInput struct {
	UserID    uuid.UUID
	PaymentID uuid.UUID
	StorageID string
}

func (s *Service) SubmitSlip(ctx context.Context, in SubmitSlipInput) (*ent.PaymentTransactionEntity, error) {
	row, err := s.paymentEnt.GetPaymentTransactionByID(ctx, in.PaymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if row.UserID != in.UserID {
		return nil, ErrPaymentForbidden
	}
	if row.Status != ent.PaymentStatusPending {
		return nil, ErrPaymentNotPending
	}
	if row.SlipStorageID != nil && strings.TrimSpace(*row.SlipStorageID) != "" {
		return nil, ErrPaymentSlipAlreadySubmitted
	}

	storageID := strings.TrimSpace(in.StorageID)
	updated, err := s.paymentEnt.UpdatePaymentSlipStorageID(ctx, row.ID, entitiesdto.UpdatePaymentSlipStorageID{
		SlipStorageID: storageID,
	})
	if err != nil {
		return nil, err
	}

	// Dispatch In-App Notification to Admins
	if s.notifEnt != nil {
		link := "/admin/payments/" + row.ID.String()
		_, _ = s.notifEnt.CreateNotification(ctx, entitiesdto.CreateNotification{
			TargetRole: "admin",
			Type:       "payment",
			Title:      "มีสลิปโอนเงินใหม่รอตรวจสอบ",
			Message:    "ผู้ใช้ได้ส่งหลักฐานการโอนเงินจำนวน " + strconv.Itoa(row.AmountTHB) + " บาท สำหรับแพ็กเกจ " + strings.ToUpper(row.PlanKey),
			Link:       &link,
		})
	}

	// Send telegram notification on slip submission
	msg := fmt.Sprintf("🔔 <b>New Payment Slip Submitted!</b>\n\n"+
		"<b>User ID :</b> %s\n"+
		"<b>Transaction ID :</b> %s\n"+
		"<b>Amount :</b> %d THB\n"+
		"<b>Plan Key :</b> %s\n\n"+
		"<i>Please review the bank slip in the Admin Dashboard.</i>",
		in.UserID.String(), row.ID.String(), row.AmountTHB, row.PlanKey)
	sendTelegramNotification(msg)

	return updated, nil
}

func (s *Service) AdminListPayments(ctx context.Context, limit int, offset int) ([]*ent.PaymentTransactionEntity, int, error) {
	return s.paymentEnt.ListPaymentTransactions(ctx, limit, offset)
}

type AdminPaymentDetail struct {
	*ent.PaymentTransactionEntity
	User *ent.UserEntity `json:"user,omitempty"`
}

func (s *Service) AdminGetPayment(ctx context.Context, id uuid.UUID) (*AdminPaymentDetail, error) {
	tx, err := s.paymentEnt.GetPaymentTransactionByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	detail := &AdminPaymentDetail{
		PaymentTransactionEntity: tx,
	}

	if user, uErr := s.userEnt.GetUserByID(ctx, tx.UserID); uErr == nil {
		detail.User = user
	}

	return detail, nil
}

type RefundPaymentInput struct {
	PaymentID  uuid.UUID
	Reason     *string
	ReviewedBy *uuid.UUID
}

func (s *Service) RefundPayment(ctx context.Context, in RefundPaymentInput) (*ent.PaymentTransactionEntity, error) {
	row, err := s.paymentEnt.GetPaymentTransactionByID(ctx, in.PaymentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	if row.Status == ent.PaymentStatusRefunded {
		return nil, ErrPaymentAlreadyRefunded
	}
	if row.Status != ent.PaymentStatusPaid {
		return nil, ErrPaymentNotPaidForRefund
	}

	var reason *string
	if in.Reason != nil {
		trimmed := strings.TrimSpace(*in.Reason)
		if trimmed != "" {
			reason = &trimmed
		}
	}

	now := time.Now()
	updated, err := s.paymentEnt.UpdatePaymentTransactionStatus(ctx, row.ID, entitiesdto.UpdatePaymentTransactionStatus{
		Status:       ent.PaymentStatusRefunded,
		ReviewReason: reason,
		ReviewedBy:   in.ReviewedBy,
		ReviewedAt:   &now,
	})
	if err != nil {
		return nil, err
	}

	// In-App Notification to User
	if s.notifEnt != nil {
		link := "/billing/payments/" + updated.ID.String()
		_, _ = s.notifEnt.CreateNotification(ctx, entitiesdto.CreateNotification{
			UserID:     &updated.UserID,
			TargetRole: "user",
			Type:       "payment",
			Title:      "คืนเงินการชำระเงินเรียบร้อยแล้ว",
			Message:    "รายการชำระเงินจำนวน " + strconv.Itoa(updated.AmountTHB) + " บาท สำหรับแพ็กเกจ " + strings.ToUpper(updated.PlanKey) + " ได้รับการคืนเงินแล้ว",
			Link:       &link,
		})
	}

	msg := fmt.Sprintf("💸 <b>Payment Refunded!</b>\n\n"+
		"<b>User ID :</b> %s\n"+
		"<b>Transaction ID :</b> %s\n"+
		"<b>Amount :</b> %d THB\n"+
		"<b>Plan Key :</b> %s\n",
		updated.UserID.String(), updated.ID.String(), updated.AmountTHB, updated.PlanKey)
	sendTelegramNotification(msg)

	return updated, nil
}

type CancelSubscriptionInput struct {
	UserID        uuid.UUID
	UseUntilMonth *string
}

func parseUseUntilMonth(raw string, loc *time.Location) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, ErrSubscriptionInvalidUseUntil
	}

	parsed, err := time.ParseInLocation("2006-01", raw, loc)
	if err != nil {
		return time.Time{}, ErrSubscriptionInvalidUseUntil
	}

	// Set to the end of selected month in local timezone.
	return time.Date(parsed.Year(), parsed.Month()+1, 0, 23, 59, 59, 0, loc), nil
}

// CancelSubscription marks the user's current subscription as cancelled.
// The plan remains active until plan_expires_at; on the next /auth/me call after
// expiry the plan will be lazily downgraded to Free.
func (s *Service) CancelSubscription(ctx context.Context, in CancelSubscriptionInput) (*ent.UserEntity, error) {
	user, err := s.userEnt.GetUserByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	if user.Plan == ent.PlanTypeFree {
		return nil, ErrSubscriptionNotActive
	}
	if user.PlanCancelledAt != nil {
		return nil, ErrSubscriptionAlreadyCancelled
	}

	now := time.Now()
	if in.UseUntilMonth == nil || strings.TrimSpace(*in.UseUntilMonth) == "" {
		return nil, ErrSubscriptionUseUntilRequired
	}

	t, parseErr := parseUseUntilMonth(*in.UseUntilMonth, now.Location())
	if parseErr != nil {
		return nil, parseErr
	}
	if t.Before(now) {
		return nil, ErrSubscriptionUseUntilInPast
	}

	if _, err = s.userEnt.SetUserPlanExpiry(ctx, in.UserID, &t, false); err != nil {
		return nil, err
	}

	return s.userEnt.CancelUserPlan(ctx, in.UserID)
}

func sendTelegramNotification(msg string) {
	token := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	chatID := strings.TrimSpace(os.Getenv("TELEGRAM_CHAT_ID"))
	if token == "" || chatID == "" {
		fmt.Println("[telegram] TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is not set, skipping notification")
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id":    chatID,
		"text":       msg,
		"parse_mode": "HTML",
	}
	body, _ := json.Marshal(payload)

	go func() {
		fmt.Printf("[telegram] sending to chat_id=%s via bot ...%s\n", chatID, token[len(token)-6:])
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("[telegram] request failed: %v\n", err)
			return
		}
		defer resp.Body.Close()
		respBody := make([]byte, 512)
		n, _ := resp.Body.Read(respBody)
		fmt.Printf("[telegram] response status=%d body=%s\n", resp.StatusCode, string(respBody[:n]))
	}()
}
