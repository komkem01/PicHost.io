package payment

import "errors"

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
)
