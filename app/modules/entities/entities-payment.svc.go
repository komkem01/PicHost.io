package entities

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var _ entitiesinf.PaymentTransactionEntity = (*Service)(nil)

func normalizeCheckoutRef(reference string) string {
	return strings.ToUpper(strings.TrimSpace(reference))
}

func normalizeProviderRef(reference string) string {
	trimmed := strings.TrimSpace(reference)
	if trimmed == "" {
		return ""
	}
	return strings.ToUpper(trimmed)
}

// buildPaymentTransactionRow normalizes a CreatePaymentTransaction DTO into an
// (unsaved) row, shared by CreatePaymentTransaction and
// CreatePaymentTransactionIfNoOpen so both insert identically-shaped rows.
func buildPaymentTransactionRow(in entitiesdto.CreatePaymentTransaction) *ent.PaymentTransactionEntity {
	now := time.Now()
	row := &ent.PaymentTransactionEntity{
		UserID:            in.UserID,
		PlanKey:           normalizePlanKey(in.PlanKey),
		AmountTHB:         in.AmountTHB,
		Currency:          strings.ToUpper(strings.TrimSpace(in.Currency)),
		Status:            in.Status,
		Provider:          strings.ToLower(strings.TrimSpace(in.Provider)),
		CheckoutReference: normalizeCheckoutRef(in.CheckoutReference),
		ProviderReference: in.ProviderReference,
		PaymentURL:        in.PaymentURL,
		ExpiresAt:         in.ExpiresAt,
		PaidAt:            in.PaidAt,
		Metadata:          in.Metadata,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if row.Currency == "" {
		row.Currency = "THB"
	}
	if row.Provider == "" {
		row.Provider = "manual"
	}
	if row.Metadata == nil {
		row.Metadata = map[string]any{}
	}

	if row.ProviderReference != nil {
		normalized := normalizeProviderRef(*row.ProviderReference)
		if normalized == "" {
			row.ProviderReference = nil
		} else {
			row.ProviderReference = &normalized
		}
	}

	return row
}

func (s *Service) CreatePaymentTransaction(ctx context.Context, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
	row := buildPaymentTransactionRow(in)
	_, err := s.db.NewInsert().
		Model(row).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// CreatePaymentTransactionIfNoOpen enforces the one-open-checkout invariant
// atomically: within a single DB transaction it locks the user row (FOR
// UPDATE, serializing concurrent checkout attempts for the same user), then
// locks any existing open transaction (pending, or paid-but-unactivated). A
// stale, no-slip pending row past its TTL is lazily expired in place; any
// other open row blocks the insert and is reported via
// entitiesdto.ErrOpenPaymentTransactionExists so the caller can surface its ID.
// Only when no open row remains does it insert the new transaction, so two
// concurrent CreateCheckout calls for the same user cannot both succeed.
func (s *Service) CreatePaymentTransactionIfNoOpen(ctx context.Context, userID uuid.UUID, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
	var created *ent.PaymentTransactionEntity

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var user ent.UserEntity
		if err := tx.NewSelect().Model(&user).Where("id = ?", userID).For("UPDATE").Scan(ctx); err != nil {
			return err
		}

		var open ent.PaymentTransactionEntity
		err := tx.NewSelect().
			Model(&open).
			Where("user_id = ?", userID).
			Where("(status = ? OR (status = ? AND activated_at IS NULL))", ent.PaymentStatusPending, ent.PaymentStatusPaid).
			OrderExpr("created_at DESC").
			Limit(1).
			For("UPDATE").
			Scan(ctx)

		switch {
		case err == nil:
			stale := open.Status == ent.PaymentStatusPending &&
				(open.SlipStorageID == nil || strings.TrimSpace(*open.SlipStorageID) == "") &&
				time.Now().After(open.ExpiresAt)
			if !stale {
				return &entitiesdto.ErrOpenPaymentTransactionExists{PaymentID: open.ID}
			}
			// Lazily expire the overdue, no-slip checkout instead of letting it
			// block a new one; still holding its row lock while we do so.
			now := time.Now()
			if _, expErr := tx.NewUpdate().
				TableExpr("payment_transactions").
				Set("status = ?, updated_at = ?", ent.PaymentStatusExpired, now).
				Where("id = ?", open.ID).
				Exec(ctx); expErr != nil {
				return expErr
			}
		case errors.Is(err, sql.ErrNoRows):
			// No open row: nothing to expire, proceed straight to insert.
		default:
			return err
		}

		row := buildPaymentTransactionRow(in)
		if _, insErr := tx.NewInsert().Model(row).Returning("*").Exec(ctx); insErr != nil {
			return insErr
		}
		created = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) GetPaymentTransactionByID(ctx context.Context, id uuid.UUID) (*ent.PaymentTransactionEntity, error) {
	var row ent.PaymentTransactionEntity
	err := s.db.NewSelect().
		Model(&row).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) GetPaymentTransactionByCheckoutReference(ctx context.Context, checkoutReference string) (*ent.PaymentTransactionEntity, error) {
	var row ent.PaymentTransactionEntity
	err := s.db.NewSelect().
		Model(&row).
		Where("checkout_reference = ?", normalizeCheckoutRef(checkoutReference)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) ListPaymentTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*ent.PaymentTransactionEntity, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	rows := make([]*ent.PaymentTransactionEntity, 0)
	err := s.db.NewSelect().
		Model(&rows).
		Where("user_id = ?", userID).
		OrderExpr("created_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) UpdatePaymentTransactionStatus(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentTransactionStatus) (*ent.PaymentTransactionEntity, error) {
	now := time.Now()
	query := s.db.NewUpdate().
		TableExpr("payment_transactions").
		Set("status = ?", in.Status).
		Set("updated_at = ?", now).
		Where("id = ?", id)

	if in.ProviderReference != nil {
		normalized := normalizeProviderRef(*in.ProviderReference)
		if normalized == "" {
			query = query.Set("provider_reference = NULL")
		} else {
			query = query.Set("provider_reference = ?", normalized)
		}
	}
	if in.PaidAt != nil {
		query = query.Set("paid_at = ?", *in.PaidAt)
	}
	if in.ReviewReason != nil {
		reviewReason := strings.TrimSpace(*in.ReviewReason)
		if reviewReason == "" {
			query = query.Set("review_reason = NULL")
		} else {
			query = query.Set("review_reason = ?", reviewReason)
		}
	}
	if in.ReviewedBy != nil {
		query = query.Set("reviewed_by = ?", *in.ReviewedBy)
	}
	if in.ReviewedAt != nil {
		query = query.Set("reviewed_at = ?", *in.ReviewedAt)
	}
	if in.Metadata != nil {
		query = query.Set("metadata = ?", in.Metadata)
	}

	if _, err := query.Exec(ctx); err != nil {
		return nil, err
	}

	return s.GetPaymentTransactionByID(ctx, id)
}

func (s *Service) UpdatePaymentSlipStorageID(ctx context.Context, id uuid.UUID, in entitiesdto.UpdatePaymentSlipStorageID) (*ent.PaymentTransactionEntity, error) {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("payment_transactions").
		Set("slip_storage_id = ?", strings.TrimSpace(in.SlipStorageID)).
		Set("updated_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetPaymentTransactionByID(ctx, id)
}

// GetOpenPaymentTransactionByUserID returns the most recent transaction that is
// still "open" for the user: pending (including one with a submitted slip
// awaiting review), or paid but not yet activated (awaiting email
// verification). Returns sql.ErrNoRows when none exists.
func (s *Service) GetOpenPaymentTransactionByUserID(ctx context.Context, userID uuid.UUID) (*ent.PaymentTransactionEntity, error) {
	var row ent.PaymentTransactionEntity
	err := s.db.NewSelect().
		Model(&row).
		Where("user_id = ?", userID).
		Where("(status = ? OR (status = ? AND activated_at IS NULL))", ent.PaymentStatusPending, ent.PaymentStatusPaid).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// validatePaymentOwnerLocked reports an error unless the (now-locked) payment
// row's user_id still matches the user row locked just before it. Used by
// ConfirmPaymentTransaction, which must lock the user row before the payment
// row (to keep lock ordering consistent with CreatePaymentTransactionIfNoOpen
// and ActivatePendingEntitlements, and so avoid a lock-order inversion
// deadlock) but only learns which user owns the payment from an earlier,
// unlocked read. user_id is immutable in this schema, so this should never
// actually fire — it exists purely to fail closed instead of silently
// activating the wrong user's entitlement if that assumption is ever violated.
func validatePaymentOwnerLocked(row *ent.PaymentTransactionEntity, expectedUserID uuid.UUID) error {
	if row.UserID != expectedUserID {
		return fmt.Errorf("entities: payment transaction %s user_id changed unexpectedly (expected %s, got %s)", row.ID, expectedUserID, row.UserID)
	}
	return nil
}

// mapPlanKeyToUserPlanType mirrors payment.Service.mapPlanKeyToUserPlan. It is
// duplicated here (rather than imported) because the payment package imports
// entities/inf, and entities cannot import payment without creating a cycle.
func mapPlanKeyToUserPlanType(planKey string) (ent.PlanType, error) {
	switch normalizePlanKey(planKey) {
	case "free":
		return ent.PlanTypeFree, nil
	case "basic":
		return ent.PlanTypeBasic, nil
	case "pro":
		return ent.PlanTypePro, nil
	case "enterprise":
		return ent.PlanTypeEnterprise, nil
	default:
		return "", fmt.Errorf("entities: unknown plan key %q", planKey)
	}
}

// activateOneLocked applies a single paid transaction's entitlement to an
// already-locked user row within an existing transaction: it stacks the
// subscription duration onto the user's plan/expiry, clears any cancellation
// intent, and marks the payment row activated. Callers must hold row locks
// (SELECT ... FOR UPDATE) on both the user and the payment before calling this.
func activateOneLocked(ctx context.Context, tx bun.Tx, user *ent.UserEntity, payment *ent.PaymentTransactionEntity, subscriptionDuration time.Duration) (*ent.UserEntity, *ent.PaymentTransactionEntity, error) {
	planValue, err := mapPlanKeyToUserPlanType(payment.PlanKey)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	var newExpiresAt time.Time
	if user.PlanExpiresAt != nil && user.PlanExpiresAt.After(now) {
		newExpiresAt = user.PlanExpiresAt.Add(subscriptionDuration)
	} else {
		newExpiresAt = now.Add(subscriptionDuration)
	}

	if _, err := tx.NewUpdate().
		TableExpr("users").
		Set("plan = ?, plan_expires_at = ?, plan_cancelled_at = NULL, updated_at = ?", planValue, newExpiresAt, now).
		Where("id = ?", user.ID).
		Exec(ctx); err != nil {
		return nil, nil, err
	}

	if _, err := tx.NewUpdate().
		TableExpr("payment_transactions").
		Set("activated_at = ?, updated_at = ?", now, now).
		Where("id = ?", payment.ID).
		Exec(ctx); err != nil {
		return nil, nil, err
	}

	var updatedUser ent.UserEntity
	if err := tx.NewSelect().Model(&updatedUser).Where("id = ?", user.ID).Scan(ctx); err != nil {
		return nil, nil, err
	}
	var updatedPayment ent.PaymentTransactionEntity
	if err := tx.NewSelect().Model(&updatedPayment).Where("id = ?", payment.ID).Scan(ctx); err != nil {
		return nil, nil, err
	}
	return &updatedUser, &updatedPayment, nil
}

// ConfirmPaymentTransaction locks the user row and then the payment row (FOR
// UPDATE, in that order) before applying any transition, so only one caller
// can move a row from pending to a terminal state even under concurrent
// webhook/admin confirmation calls. When transitioning to paid: if the email
// is verified the entitlement is activated immediately in the same
// transaction; otherwise the row is committed as paid with activated_at left
// NULL, to be activated later after email verification.
//
// Lock ordering: the payment ID is the only input, so the owning user isn't
// known up front. To still acquire locks in a consistent user→payment order —
// matching CreatePaymentTransactionIfNoOpen and ActivatePendingEntitlements,
// and so avoiding a lock-order inversion that could deadlock under concurrent
// confirm/checkout/activation calls — this does an unlocked "probe" read of
// the payment row first purely to discover user_id, locks the user row, then
// locks the payment row and re-validates its user_id still matches the probe
// (user_id is immutable in this schema, but this guards against relying on
// that unlocked read being stale in some future change).
func (s *Service) ConfirmPaymentTransaction(ctx context.Context, in entitiesdto.ConfirmPaymentTransaction) (*entitiesdto.ConfirmPaymentTransactionResult, error) {
	result := &entitiesdto.ConfirmPaymentTransactionResult{}

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var probe ent.PaymentTransactionEntity
		if err := tx.NewSelect().Model(&probe).Where("id = ?", in.PaymentID).Scan(ctx); err != nil {
			return err
		}

		var user ent.UserEntity
		if err := tx.NewSelect().Model(&user).Where("id = ?", probe.UserID).For("UPDATE").Scan(ctx); err != nil {
			return err
		}

		var row ent.PaymentTransactionEntity
		if err := tx.NewSelect().Model(&row).Where("id = ?", in.PaymentID).For("UPDATE").Scan(ctx); err != nil {
			return err
		}
		if err := validatePaymentOwnerLocked(&row, user.ID); err != nil {
			return err
		}

		// Only a pending row may transition. Anything else was already decided
		// by another (possibly concurrent/duplicate) caller; report it as-is.
		if row.Status != ent.PaymentStatusPending {
			result.Payment = &row
			result.AlreadyTerminal = true
			return nil
		}

		now := time.Now()
		update := tx.NewUpdate().
			TableExpr("payment_transactions").
			Set("status = ?", in.NextStatus).
			Set("updated_at = ?", now).
			Where("id = ?", in.PaymentID)

		if in.NextStatus == ent.PaymentStatusPaid {
			update = update.Set("paid_at = ?", now)
		}
		if in.ProviderReference != nil {
			normalized := normalizeProviderRef(*in.ProviderReference)
			if normalized == "" {
				update = update.Set("provider_reference = NULL")
			} else {
				update = update.Set("provider_reference = ?", normalized)
			}
		}
		if in.ReviewReason != nil {
			reviewReason := strings.TrimSpace(*in.ReviewReason)
			if reviewReason == "" {
				update = update.Set("review_reason = NULL")
			} else {
				update = update.Set("review_reason = ?", reviewReason)
			}
		}
		if in.ReviewedBy != nil {
			update = update.Set("reviewed_by = ?", *in.ReviewedBy)
		}
		if in.ReviewedAt != nil {
			update = update.Set("reviewed_at = ?", *in.ReviewedAt)
		}
		if in.Metadata != nil {
			update = update.Set("metadata = ?", in.Metadata)
		}

		if _, err := update.Exec(ctx); err != nil {
			return err
		}

		if err := tx.NewSelect().Model(&row).Where("id = ?", in.PaymentID).Scan(ctx); err != nil {
			return err
		}
		result.Payment = &row

		if in.NextStatus != ent.PaymentStatusPaid {
			return nil
		}

		// user is already locked (FOR UPDATE) above, acquired before the payment
		// row per the user→payment lock ordering — no second lock needed here.
		if user.EmailVerifiedAt == nil {
			// Approved, but held until the user verifies their email.
			result.User = &user
			return nil
		}

		updatedUser, updatedPayment, err := activateOneLocked(ctx, tx, &user, &row, in.SubscriptionDuration)
		if err != nil {
			return err
		}
		result.User = updatedUser
		result.Payment = updatedPayment
		result.Activated = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ActivatePendingEntitlements locks the user row, then locks and activates
// every paid-but-unactivated transaction for that user in the same
// transaction, processing rows in chronological order (paid_at, then
// created_at) so legacy/racing rows each stack their duration exactly once.
// Safe to call repeatedly: a replay after successful activation is a no-op
// because activated rows are excluded by the WHERE clause.
func (s *Service) ActivatePendingEntitlements(ctx context.Context, userID uuid.UUID, subscriptionDuration time.Duration) (*entitiesdto.ActivateEntitlementsResult, error) {
	result := &entitiesdto.ActivateEntitlementsResult{}

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var user ent.UserEntity
		if err := tx.NewSelect().Model(&user).Where("id = ?", userID).For("UPDATE").Scan(ctx); err != nil {
			return err
		}
		result.User = &user

		if user.EmailVerifiedAt == nil {
			// Nothing to activate until the email is verified.
			return nil
		}

		var rows []*ent.PaymentTransactionEntity
		if err := tx.NewSelect().
			Model(&rows).
			Where("user_id = ?", userID).
			Where("status = ?", ent.PaymentStatusPaid).
			Where("activated_at IS NULL").
			OrderExpr("paid_at ASC NULLS LAST, created_at ASC").
			For("UPDATE").
			Scan(ctx); err != nil {
			return err
		}

		curUser := &user
		activated := make([]*ent.PaymentTransactionEntity, 0, len(rows))
		for _, row := range rows {
			updatedUser, updatedPayment, err := activateOneLocked(ctx, tx, curUser, row, subscriptionDuration)
			if err != nil {
				return err
			}
			curUser = updatedUser
			activated = append(activated, updatedPayment)
		}
		result.User = curUser
		result.ActivatedPayments = activated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ListPaymentTransactions(ctx context.Context, limit int, offset int) ([]*ent.PaymentTransactionEntity, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	rows := make([]*ent.PaymentTransactionEntity, 0)
	total, err := s.db.NewSelect().
		Model(&rows).
		OrderExpr("created_at DESC").
		Limit(limit).
		Offset(offset).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
