package storage

import (
	"context"

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
	return s.store.CreateStorage(ctx, entitiesdto.CreateStorage{
		Provider: req.Provider,
		Path:     req.Path,
		URL:      req.URL,
		FileSize: req.FileSize,
		MIMEType: req.MIMEType,
	})
}
