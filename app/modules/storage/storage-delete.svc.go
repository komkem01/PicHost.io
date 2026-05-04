package storage

import (
	"context"

	"github.com/google/uuid"
)

func (s *Service) DeleteFile(ctx context.Context, id uuid.UUID) error {
	if _, err := s.GetFile(ctx, id); err != nil {
		return err
	}
	return s.store.DeleteStorage(ctx, id)
}
