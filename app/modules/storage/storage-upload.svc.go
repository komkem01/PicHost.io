package storage

import (
	"context"
	"fmt"
	"mime/multipart"
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

	return s.saveSource(ctx, provider, src, req.FileSize, req.MIMEType)
}

// UploadFromSource uploads a pre-built uploadSource (e.g. from a multipart file) to S3
// and persists the storage record.
func (s *Service) UploadFromSource(ctx context.Context, provider string, src *uploadSource) (*ent.StorageEntity, error) {
	if strings.TrimSpace(provider) == "" {
		provider = "Railway"
	}
	return s.saveSource(ctx, provider, src, src.Size, nil)
}

func (s *Service) saveSource(ctx context.Context, provider string, src *uploadSource, hintSize int64, hintMIME *string) (*ent.StorageEntity, error) {
	objectPath, objectURL, uploadedSize, uploadedMIME, err := s.uploadToS3(ctx, src)
	if err != nil {
		return nil, err
	}

	finalSize := uploadedSize
	if finalSize <= 0 && hintSize > 0 {
		finalSize = hintSize
	}

	finalMIME := uploadedMIME
	if hintMIME != nil && strings.TrimSpace(*hintMIME) != "" {
		finalMIME = strings.TrimSpace(*hintMIME)
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

// UploadMultipartFile uploads a multipart form file to S3 and persists a storage record.
// Unlike UploadFile, it does NOT create an image record – safe to use for internal uploads
// such as payment slips that should not count towards user image quota.
func (s *Service) UploadMultipartFile(ctx context.Context, provider string, fh *multipart.FileHeader) (*ent.StorageEntity, error) {
	if strings.TrimSpace(provider) == "" {
		provider = "Railway"
	}
	src, err := formFileToSource(fh)
	if err != nil {
		return nil, err
	}
	defer src.Reader.Close()
	return s.saveSource(ctx, provider, src, src.Size, nil)
}
