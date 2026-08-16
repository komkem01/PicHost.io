package mailerinf

import (
	"context"
)

type Mailer interface {
	SendEmail(ctx context.Context, to, subject, htmlBody, textBody string) error
	SendPasswordReset(ctx context.Context, to, resetURL string, isTH bool) error
	SendEmailVerification(ctx context.Context, to, verifyURL string, isTH bool) error
	SendSlipApproved(ctx context.Context, to, planName string, isTH bool) error
	SendSlipRejected(ctx context.Context, to, reason string, isTH bool) error
	// SendPaymentAwaitingVerification notifies the user that their payment was
	// approved but the plan will only activate once they verify their email.
	SendPaymentAwaitingVerification(ctx context.Context, to, planName string, isTH bool) error
}
