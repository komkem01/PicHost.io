package storage

import (
	"context"

	entitiesdto "pichost.io/app/modules/entities/dto"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func (s *Service) DeleteFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	st, err := s.GetFile(ctx, id)
	if err != nil {
		img, imgErr := s.imageEnt.GetImageByID(ctx, id)
		if imgErr == nil && img != nil && img.StorageID != uuid.Nil {
			id = img.StorageID
			st, err = s.GetFile(ctx, id)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	img, err := s.imageEnt.GetImageByStorageID(ctx, id)
	if err == nil && img != nil {
		if img.UserID == nil || *img.UserID != userID {
			return ErrStorageNotFound
		}
		_ = s.imageEnt.DeleteImage(ctx, img.ID)
	}

	if st != nil && s.userQuota != nil {
		_, _ = s.userQuota.AddToUserQuota(ctx, userID, entitiesdto.AddToUserQuota{
			StorageDelta:    -st.FileSize,
			ImageCountDelta: -1,
		})
	}

	// Remove physical object from MinIO / S3 storage bucket
	if st != nil {
		if objectPath, err := s.objectPathFromStorage(st); err == nil && objectPath != "" {
			if conf, err := s.getS3Config(); err == nil {
				if client, err := s.newMinioClient(conf); err == nil {
					_ = client.RemoveObject(ctx, conf.BucketName, objectPath, minio.RemoveObjectOptions{})
				}
			}
		}
	}

	return s.store.DeleteStorage(ctx, id)
}
