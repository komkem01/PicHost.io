package image

import (
	"context"
	"database/sql"
	"errors"

	"pichost.io/app/modules/entities/ent"
	quotamod "pichost.io/app/modules/quota"

	"github.com/google/uuid"
)

type ImageWithStorage struct {
	Image   *ent.ImageEntity
	Storage *ent.StorageEntity
}

func (s *Service) ListImages(ctx context.Context, userID uuid.UUID) ([]ImageWithStorage, error) {
	if err := s.quotaSvc.EnsureUsageAllowed(ctx, userID, false); err != nil {
		if errors.Is(err, quotamod.ErrQuotaAccountLocked) {
			return nil, ErrImageAccountLocked
		}
		return nil, err
	}

	images, err := s.image.GetImagesByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []ImageWithStorage{}, nil
		}
		return nil, err
	}

	result := make([]ImageWithStorage, 0, len(images))
	for _, img := range images {
		storage, err := s.store.GetStorageByID(ctx, img.StorageID)
		if err != nil {
			continue
		}
		result = append(result, ImageWithStorage{Image: img, Storage: storage})
	}
	return result, nil
}
