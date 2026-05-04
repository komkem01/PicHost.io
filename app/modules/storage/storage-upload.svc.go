package storage

import (
	"context"
	"fmt"
	"strings"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
)

type UploadRequestService struct {
	Provider string
	Path     *string
	URL      *string
	FileSize int64
	MIMEType *string
}

func (s *Service) Upload(ctx context.Context, req UploadRequestService) (*ent.StorageEntity, error) {
	provider := strings.TrimSpace(req.Provider)
	if provider == "" {
		provider = "Railway"
	}

	src, err := s.openSource(ctx, req)
	if err != nil {
		return nil, err
	}
	defer src.Reader.Close()

	objectPath, objectURL, uploadedSize, uploadedMIME, err := s.uploadToS3(ctx, src)
	if err != nil {
		return nil, err
	}

	finalSize := uploadedSize
	if finalSize <= 0 && req.FileSize > 0 {
		finalSize = req.FileSize
	}

	finalMIME := uploadedMIME
	if req.MIMEType != nil && strings.TrimSpace(*req.MIMEType) != "" {
		finalMIME = strings.TrimSpace(*req.MIMEType)
	}

	path := objectPath
	url := objectURL

	var lastErr error
	for i := 0; i < 5; i++ {
		shortCode, genErr := s.generateShortCode(storageShortCodeLength)
		if genErr != nil {
			return nil, genErr
		}

		created, createErr := s.store.CreateStorage(ctx, entitiesdto.CreateStorage{
			ShortCode: shortCode,
			Provider:  provider,
			Path:      &path,
			URL:       &url,
			FileSize:  finalSize,
			MIMEType:  &finalMIME,
		})
		if createErr == nil {
			return created, nil
		}

		lastErr = createErr
		if !s.isShortCodeUniqueConflict(createErr) {
			return nil, createErr
		}
	}

	return nil, fmt.Errorf("failed to generate unique short code after retries: %w", lastErr)
}
