package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

func (s *Service) GetFile(ctx context.Context, id uuid.UUID) (*ent.StorageEntity, error) {
	data, err := s.store.GetStorageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}
	return data, nil
}

func (s *Service) GetPresignURL(ctx context.Context, rawURL string) (*ent.StorageEntity, error) {
	data, err := s.store.GetStorageByURL(ctx, rawURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	presignedURL, err := s.PresignStorage(ctx, data)
	if err != nil {
		return nil, err
	}
	data.URL = &presignedURL

	return data, nil
}

func (s *Service) GetPresignURLByID(ctx context.Context, id uuid.UUID) (*ent.StorageEntity, error) {
	data, err := s.store.GetStorageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	presignedURL, err := s.PresignStorage(ctx, data)
	if err != nil {
		return nil, err
	}
	data.URL = &presignedURL

	return data, nil
}

func (s *Service) GetPresignURLByShortCode(ctx context.Context, shortCode string) (*ent.StorageEntity, error) {
	code := strings.TrimSpace(shortCode)
	if code == "" {
		return nil, ErrStorageNotFound
	}

	data, err := s.store.GetStorageByShortCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	presignedURL, err := s.PresignStorage(ctx, data)
	if err != nil {
		return nil, err
	}
	data.URL = &presignedURL

	return data, nil
}
