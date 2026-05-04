package storage

import (
	"context"

	"pichost.io/app/modules/entities/ent"
)

func (s *Service) ListFiles(ctx context.Context) ([]*ent.StorageEntity, error) {
	return s.store.GetListStorage(ctx)
}
