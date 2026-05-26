package entities

import (
	"context"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
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

func (s *Service) CreatePaymentTransaction(ctx context.Context, in entitiesdto.CreatePaymentTransaction) (*ent.PaymentTransactionEntity, error) {
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

	_, err := s.db.NewInsert().
		Model(row).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return row, nil
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
