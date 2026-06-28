package entities

import (
	"context"
	"fmt"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

var _ entitiesinf.ImageEntity = (*Service)(nil)

func (s *Service) CreateImage(ctx context.Context, image entitiesdto.CreateImage) (*ent.ImageEntity, error) {
	if image.StorageID == nil {
		return nil, fmt.Errorf("storage_id is required")
	}

	var userIDPtr *uuid.UUID
	if image.UserID != nil {
		parsed, err := uuid.Parse(*image.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %w", err)
		}
		userIDPtr = &parsed
	}
	storageID, err := uuid.Parse(*image.StorageID)
	if err != nil {
		return nil, fmt.Errorf("invalid storage_id: %w", err)
	}

	isPrivate := false
	if image.IsPrivate != nil {
		isPrivate = *image.IsPrivate
	}

	var expiresAt *time.Time
	if image.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *image.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at: %w", err)
		}
		expiresAt = &t
	}

	now := time.Now()
	data := &ent.ImageEntity{
		UserID:    userIDPtr,
		StorageID: storageID,
		IsPrivate: isPrivate,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	_, err = s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetImageByID(ctx context.Context, id uuid.UUID) (*ent.ImageEntity, error) {
	var image ent.ImageEntity
	err := s.db.NewSelect().
		Model(&image).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (s *Service) GetImageByStorageID(ctx context.Context, storageID uuid.UUID) (*ent.ImageEntity, error) {
	var image ent.ImageEntity
	err := s.db.NewSelect().
		Model(&image).
		Where("storage_id = ?", storageID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (s *Service) GetImagesByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.ImageEntity, error) {
	var images []*ent.ImageEntity
	err := s.db.NewSelect().
		Model(&images).
		Where("user_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (s *Service) UpdateImage(ctx context.Context, id uuid.UUID, image entitiesdto.UpdateImage) (*ent.ImageEntity, error) {
	query := s.db.NewUpdate().
		Model((*ent.ImageEntity)(nil)).
		Table("images").
		Where("id = ?", id)

	updated := 0
	if image.UserID != nil {
		userID, err := uuid.Parse(*image.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %w", err)
		}
		query = query.Set("user_id = ?", userID)
		updated++
	}
	if image.StorageID != nil {
		storageID, err := uuid.Parse(*image.StorageID)
		if err != nil {
			return nil, fmt.Errorf("invalid storage_id: %w", err)
		}
		query = query.Set("storage_id = ?", storageID)
		updated++
	}
	if image.IsPrivate != nil {
		query = query.Set("is_private = ?", *image.IsPrivate)
		updated++
	}

	if updated == 0 {
		return s.GetImageByID(ctx, id)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, fmt.Errorf("image not found")
	}

	return s.GetImageByID(ctx, id)
}

func (s *Service) DeleteImage(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.NewDelete().
		Model((*ent.ImageEntity)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// ListExpiredImages returns all images whose expires_at is set and before the given time.
func (s *Service) ListExpiredImages(ctx context.Context, before time.Time) ([]*ent.ImageEntity, error) {
	var images []*ent.ImageEntity
	err := s.db.NewSelect().
		Model(&images).
		Where("expires_at IS NOT NULL AND expires_at <= ?", before).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return images, nil
}

// GetGuestStats counts all guest images (where user_id IS NULL) and sums their storage file size.
func (s *Service) GetGuestStats(ctx context.Context) (int, int64, error) {
	var stats struct {
		Count int   `bun:"count"`
		Size  int64 `bun:"size"`
	}
	err := s.db.NewSelect().
		TableExpr("images AS i").
		Join("JOIN storages AS s ON s.id = i.storage_id").
		ColumnExpr("count(i.id) AS count, coalesce(sum(s.file_size), 0) AS size").
		Where("i.user_id IS NULL").
		Scan(ctx, &stats)
	if err != nil {
		return 0, 0, err
	}
	return stats.Count, stats.Size, nil
}

// GetUniqueGuestIPCount counts the number of unique IP addresses of successful guest uploads since the given time.
func (s *Service) GetUniqueGuestIPCount(ctx context.Context, since time.Time) (int, error) {
	var count int
	err := s.db.NewSelect().
		Model((*ent.AuditLogEntity)(nil)).
		ColumnExpr("count(distinct ip_address)").
		Where("action = ? AND status = ? AND created_at >= ?", "storage.upload_guest", "success", since).
		Scan(ctx, &count)
	return count, err
}


