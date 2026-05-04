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

var _ entitiesinf.StorageEntity = (*Service)(nil)

func (s *Service) CreateStorage(ctx context.Context, storage entitiesdto.CreateStorage) (*ent.StorageEntity, error) {
	now := time.Now()
	data := &ent.StorageEntity{
		Provider:  storage.Provider,
		Path:      storage.Path,
		URL:       storage.URL,
		FileSize:  storage.FileSize,
		MIMEType:  storage.MIMEType,
		CreatedAt: now,
	}

	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetStorageByID(ctx context.Context, id uuid.UUID) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) GetListStorage(ctx context.Context) ([]*ent.StorageEntity, error) {
	var storages []*ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storages).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return storages, nil
}

func (s *Service) GetStorageByURL(ctx context.Context, url string) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("url = ?", url).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) GetStorageByEmail(ctx context.Context, email string) (*ent.StorageEntity, error) {
	var storage ent.StorageEntity
	err := s.db.NewSelect().
		Model(&storage).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (s *Service) UpdateStorage(ctx context.Context, id uuid.UUID, storage entitiesdto.UpdateStorage) (*ent.StorageEntity, error) {
	query := s.db.NewUpdate().
		Model((*ent.StorageEntity)(nil)).
		Table("storages").
		Where("id = ?", id)

	updated := 0
	if storage.Provider != nil {
		query = query.Set("provider = ?", *storage.Provider)
		updated++
	}
	if storage.Path != nil {
		query = query.Set("path = ?", *storage.Path)
		updated++
	}
	if storage.URL != nil {
		query = query.Set("url = ?", *storage.URL)
		updated++
	}
	if storage.FileSize != nil {
		query = query.Set("file_size = ?", *storage.FileSize)
		updated++
	}
	if storage.MIMEType != nil {
		query = query.Set("mime_type = ?", *storage.MIMEType)
		updated++
	}

	if updated == 0 {
		return s.GetStorageByID(ctx, id)
	}

	result, err := query.Exec(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err == nil && rows == 0 {
		return nil, fmt.Errorf("storage not found")
	}

	return s.GetStorageByID(ctx, id)
}

func (s *Service) DeleteStorage(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.NewDelete().
		Model((*ent.StorageEntity)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
