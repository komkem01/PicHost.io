package payment

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrPaymentPlanUnavailable       = errors.New("payment: selected plan is unavailable")
	ErrPaymentNotFound              = errors.New("payment: transaction not found")
	ErrPaymentForbidden             = errors.New("payment: transaction does not belong to user")
	ErrPaymentInvalidStatus         = errors.New("payment: invalid status")
	ErrPaymentInvalidAmount         = errors.New("payment: invalid paid amount")
	ErrPaymentWebhookDenied         = errors.New("payment: invalid webhook token")
	ErrPaymentNotPending            = errors.New("payment: transaction is no longer pending")
	ErrPaymentSlipAlreadySubmitted  = errors.New("payment: slip has already been submitted")
	ErrPaymentReviewReasonRequired  = errors.New("payment: review reason is required for rejection")
	ErrSubscriptionNotActive        = errors.New("payment: no active paid subscription to cancel")
	ErrSubscriptionAlreadyCancelled = errors.New("payment: subscription is already cancelled")
	ErrSubscriptionUseUntilRequired = errors.New("payment: use_until_month is required for this account")
	ErrSubscriptionInvalidUseUntil  = errors.New("payment: invalid use_until_month format, expected YYYY-MM")
	ErrSubscriptionUseUntilInPast   = errors.New("payment: use_until_month cannot be in the past")
	ErrPaymentNotPaidForRefund      = errors.New("payment: only paid transactions can be refunded")
	ErrPaymentAlreadyRefunded       = errors.New("payment: transaction is already refunded")
	ErrWebhookSecretRequired        = errors.New("payment: webhook secret is not configured on server")
	// ErrPaymentOpenExistsSentinel is wrapped by *ErrPaymentOpenExists so callers
	// can still match it with errors.Is if they don't need the payment ID.
	ErrPaymentOpenExistsSentinel = errors.New("payment: an open checkout already exists for this user")
)

// ErrPaymentOpenExists is returned by CreateCheckout when the user already has
// an open transaction: pending (including one with a submitted slip awaiting
// review), or paid but not yet activated (awaiting email verification). It
// carries the existing payment's ID so the caller/controller can point the
// user at it instead of creating a duplicate charge.
type ErrPaymentOpenExists struct {
	PaymentID uuid.UUID
}

func (e *ErrPaymentOpenExists) Error() string {
	return ErrPaymentOpenExistsSentinel.Error()
}

func (e *ErrPaymentOpenExists) Unwrap() error {
	return ErrPaymentOpenExistsSentinel
}
