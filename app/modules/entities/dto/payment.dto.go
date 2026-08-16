package entitiesdto

import (
	"time"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

type CreatePaymentTransaction struct {
	UserID            uuid.UUID
	PlanKey           string
	AmountTHB         int
	Currency          string
	Status            ent.PaymentStatus
	Provider          string
	CheckoutReference string
	ProviderReference *string
	PaymentURL        *string
	ExpiresAt         time.Time
	PaidAt            *time.Time
	Metadata          map[string]any
}

type UpdatePaymentTransactionStatus struct {
	Status            ent.PaymentStatus
	ProviderReference *string
	PaidAt            *time.Time
	ReviewReason      *string
	ReviewedBy        *uuid.UUID
	ReviewedAt        *time.Time
	Metadata          map[string]any
}

type UpdatePaymentSlipStorageID struct {
	SlipStorageID string
}

// ErrOpenPaymentTransactionExists is returned by CreatePaymentTransactionIfNoOpen
// when the user already has an open (pending, or paid-but-not-yet-activated)
// checkout. PaymentID identifies that existing row so callers (payment.Service)
// can surface it to the client without a second query. Defined here (rather
// than in the payment package) so the entities layer never has to import
// payment, which would create an import cycle.
type ErrOpenPaymentTransactionExists struct {
	PaymentID uuid.UUID
}

func (e *ErrOpenPaymentTransactionExists) Error() string {
	return "entities: an open payment transaction already exists for this user"
}

// ConfirmPaymentTransaction is the input for the transactional confirm+activate
// operation. It locks the payment row (and, when transitioning to paid, the user
// row) so concurrent/duplicate confirmations can only apply the transition once.
type ConfirmPaymentTransaction struct {
	PaymentID            uuid.UUID
	NextStatus           ent.PaymentStatus
	ProviderReference    *string
	ReviewReason         *string
	ReviewedBy           *uuid.UUID
	ReviewedAt           *time.Time
	Metadata             map[string]any
	SubscriptionDuration time.Duration
}

// ConfirmPaymentTransactionResult reports what the transactional confirm+activate
// operation actually did, so callers can send accurate notifications and report
// truthful "upgraded" status instead of assuming paid == activated.
type ConfirmPaymentTransactionResult struct {
	// Payment is the payment row after the transition (or the unchanged row when
	// AlreadyTerminal is true).
	Payment *ent.PaymentTransactionEntity
	// User is populated only when NextStatus is paid; it reflects the user row
	// as read (and possibly updated) inside the transaction.
	User *ent.UserEntity
	// Activated is true only when this call actually applied the plan/expiry
	// change and set activated_at.
	Activated bool
	// AlreadyTerminal is true when the payment row was no longer pending by the
	// time it was locked (already handled by another caller), so no transition
	// or activation was applied by this call.
	AlreadyTerminal bool
}

// ActivateEntitlementsResult reports the outcome of activating every
// paid-but-unactivated transaction for a user in one transaction.
type ActivateEntitlementsResult struct {
	User              *ent.UserEntity
	ActivatedPayments []*ent.PaymentTransactionEntity
}
