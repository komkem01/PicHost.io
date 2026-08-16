package entities

import (
	"testing"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

// TestValidatePaymentOwnerLocked_MatchingUser_NoError exercises the
// fail-closed guard ConfirmPaymentTransaction relies on after switching its
// lock order to user→payment (matching CreatePaymentTransactionIfNoOpen and
// ActivatePendingEntitlements, avoiding a lock-order inversion deadlock).
// Since the payment ID is the only input, the owning user is discovered via
// an unlocked probe read *before* the user row is locked; this guard is what
// makes it safe to trust that probe once the payment row is actually locked.
func TestValidatePaymentOwnerLocked_MatchingUser_NoError(t *testing.T) {
	userID := uuid.New()
	row := &ent.PaymentTransactionEntity{
		ID:     uuid.New(),
		UserID: userID,
	}

	if err := validatePaymentOwnerLocked(row, userID); err != nil {
		t.Fatalf("expected no error when the locked payment's user_id matches the locked user, got %v", err)
	}
}

// TestValidatePaymentOwnerLocked_MismatchedUser_FailsClosed simulates the
// (should-never-happen, since user_id is immutable) case where the payment
// row's user_id no longer matches the user that was locked based on the
// earlier unlocked probe read. The guard must fail closed rather than let
// ConfirmPaymentTransaction activate an entitlement against the wrong user.
func TestValidatePaymentOwnerLocked_MismatchedUser_FailsClosed(t *testing.T) {
	expectedUserID := uuid.New()
	row := &ent.PaymentTransactionEntity{
		ID:     uuid.New(),
		UserID: uuid.New(), // a different user than the one we locked
	}

	if err := validatePaymentOwnerLocked(row, expectedUserID); err == nil {
		t.Fatalf("expected an error when the payment's user_id does not match the locked user")
	}
}
