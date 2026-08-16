package auth

import (
	"context"

	"github.com/google/uuid"
)

// EntitlementActivator activates any plan entitlement that was approved
// (paid) while the user's email was still unverified. It is defined here,
// in auth, rather than imported from payment: payment does not import auth,
// so injecting an implementation (payment.Service) through this interface
// after both modules are constructed avoids a package import cycle while
// keeping all payment-specific duration/notification logic in the payment
// module.
type EntitlementActivator interface {
	// ActivatePendingEntitlements activates every paid-but-unactivated
	// transaction for userID. It returns true if at least one transaction was
	// activated. It must be safe to call repeatedly (idempotent) so a replayed
	// verification link can retry safely after a transient failure.
	ActivatePendingEntitlements(ctx context.Context, userID uuid.UUID) (bool, error)
}
