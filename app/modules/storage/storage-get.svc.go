package storage

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

func (s *Service) checkAccess(ctx context.Context, storageID uuid.UUID, callerID *uuid.UUID) error {
	img, err := s.imageEnt.GetImageByStorageID(ctx, storageID)
	if err == nil && img != nil && img.IsPrivate {
		if callerID == nil || img.UserID == nil || *img.UserID != *callerID {
			return ErrStorageNotFound
		}
	}
	return nil
}

func (s *Service) GetFile(ctx context.Context, id uuid.UUID, callerID ...*uuid.UUID) (*ent.StorageEntity, error) {
	data, err := s.store.GetStorageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	var cID *uuid.UUID
	if len(callerID) > 0 {
		cID = callerID[0]
	}

	if err := s.checkAccess(ctx, data.ID, cID); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Service) GetPresignURL(ctx context.Context, rawURL string, callerID ...*uuid.UUID) (*ent.StorageEntity, error) {
	data, err := s.store.GetStorageByURL(ctx, rawURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	var cID *uuid.UUID
	if len(callerID) > 0 {
		cID = callerID[0]
	}

	if err := s.checkAccess(ctx, data.ID, cID); err != nil {
		return nil, err
	}

	presignedURL, err := s.PresignStorage(ctx, data)
	if err != nil {
		return nil, err
	}
	data.URL = &presignedURL

	return data, nil
}

func (s *Service) GetPresignURLByID(ctx context.Context, id uuid.UUID, callerID ...*uuid.UUID) (*ent.StorageEntity, error) {
	data, err := s.store.GetStorageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStorageNotFound
		}
		return nil, err
	}

	var cID *uuid.UUID
	if len(callerID) > 0 {
		cID = callerID[0]
	}

	if err := s.checkAccess(ctx, data.ID, cID); err != nil {
		return nil, err
	}

	presignedURL, err := s.PresignStorage(ctx, data)
	if err != nil {
		return nil, err
	}
	data.URL = &presignedURL

	return data, nil
}

func (s *Service) GetPresignURLByShortCode(ctx context.Context, shortCode string, callerID ...*uuid.UUID) (*ent.StorageEntity, error) {
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

	var cID *uuid.UUID
	if len(callerID) > 0 {
		cID = callerID[0]
	}

	if err := s.checkAccess(ctx, data.ID, cID); err != nil {
		return nil, err
	}

	presignedURL, err := s.PresignStorage(ctx, data)
	if err != nil {
		return nil, err
	}
	data.URL = &presignedURL

	return data, nil
}
