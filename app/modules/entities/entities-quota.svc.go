package entities

import (
	"context"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

var _ entitiesinf.UserQuotaEntity = (*Service)(nil)

// GetUserQuota retrieves the quota record for a user.
func (s *Service) GetUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error) {
	var quota ent.UserQuotaEntity
	err := s.db.NewSelect().
		Model(&quota).
		Where("uq.user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

// UpsertUserQuota ensures a quota row exists for the user (creates with zeroes if absent).
func (s *Service) UpsertUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error) {
	now := time.Now()
	data := &ent.UserQuotaEntity{
		UserID:           userID,
		UsedStorageBytes: 0,
		ImageCount:       0,
		UpdatedAt:        now,
	}
	_, err := s.db.NewInsert().
		Model(data).
		On("CONFLICT (user_id) DO NOTHING").
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	// If the row already existed, fetch and return it.
	if data.ID == uuid.Nil {
		return s.GetUserQuota(ctx, userID)
	}
	return data, nil
}

// AddToUserQuota atomically increments (or decrements) storage and image count for a user.
func (s *Service) AddToUserQuota(ctx context.Context, userID uuid.UUID, delta entitiesdto.AddToUserQuota) (*ent.UserQuotaEntity, error) {
	var quota ent.UserQuotaEntity
	_, err := s.db.NewUpdate().
		Model(&quota).
		Set("used_storage_bytes = used_storage_bytes + ?", delta.StorageDelta).
		Set("image_count = image_count + ?", delta.ImageCountDelta).
		Set("updated_at = ?", time.Now()).
		Where("user_id = ?", userID).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &quota, nil
}
