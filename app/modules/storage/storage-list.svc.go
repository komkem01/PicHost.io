package storage

import (
	"context"
	"database/sql"
	"errors"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

func (s *Service) ListFiles(ctx context.Context, userID uuid.UUID) ([]*ent.StorageEntity, error) {
	images, err := s.imageEnt.GetImagesByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*ent.StorageEntity{}, nil
		}
		return nil, err
	}

	result := make([]*ent.StorageEntity, 0, len(images))
	for _, img := range images {
		storage, err := s.store.GetStorageByID(ctx, img.StorageID)
		if err != nil {
			continue
		}
		result = append(result, storage)
	}
	return result, nil
}
