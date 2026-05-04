package image

import (
	"context"
	"database/sql"
	"errors"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

func (s *Service) GetImage(ctx context.Context, id uuid.UUID) (*ent.ImageEntity, error) {
	item, err := s.image.GetImageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return item, nil
}

func (s *Service) GetPresignURL(ctx context.Context, imageID uuid.UUID) (string, error) {
	item, err := s.GetImage(ctx, imageID)
	if err != nil {
		return "", err
	}

	storage, err := s.store.GetStorageByID(ctx, item.StorageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrImageURLNotFound
		}
		return "", err
	}
	if storage.URL == nil || *storage.URL == "" {
		return "", ErrImageURLNotFound
	}

	return *storage.URL, nil
}
