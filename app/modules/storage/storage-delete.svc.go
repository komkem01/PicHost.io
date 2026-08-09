package storage

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) DeleteFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if _, err := s.GetFile(ctx, id); err != nil {
		return err
	}

	img, err := s.imageEnt.GetImageByStorageID(ctx, id)
	if err == nil && img != nil {
		if img.UserID == nil || *img.UserID != userID {
			return ErrStorageNotFound
		}
		_ = s.imageEnt.DeleteImage(ctx, img.ID)
	}

	return s.store.DeleteStorage(ctx, id)
}
