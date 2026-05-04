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
	if image.UserID == nil || image.StorageID == nil {
		return nil, fmt.Errorf("user_id and storage_id are required")
	}

	userID, err := uuid.Parse(*image.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}
	storageID, err := uuid.Parse(*image.StorageID)
	if err != nil {
		return nil, fmt.Errorf("invalid storage_id: %w", err)
	}

	isPrivate := false
	if image.IsPrivate != nil {
		isPrivate = *image.IsPrivate
	}

	now := time.Now()
	data := &ent.ImageEntity{
		UserID:    userID,
		StorageID: storageID,
		IsPrivate: isPrivate,
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
