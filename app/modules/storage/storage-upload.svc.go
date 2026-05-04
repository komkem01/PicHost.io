package storage

import (
	"context"
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

	created, err := s.store.CreateStorage(ctx, entitiesdto.CreateStorage{
		Provider: provider,
		Path:     &path,
		URL:      &url,
		FileSize: finalSize,
		MIMEType: &finalMIME,
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}
