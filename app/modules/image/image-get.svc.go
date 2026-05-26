package image

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"pichost.io/app/modules/entities/ent"
	quotamod "pichost.io/app/modules/quota"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func (s *Service) GetImage(ctx context.Context, id uuid.UUID) (*ent.ImageEntity, error) {
	item, err := s.image.GetImageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}

	// Lazy delete: if expires_at is set and in the past, delete and return 404.
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		_ = s.image.DeleteImage(ctx, id)
		return nil, ErrImageExpired
	}

	if item.UserID != nil && *item.UserID != uuid.Nil {
		if lockErr := s.quotaSvc.EnsureUsageAllowed(ctx, *item.UserID, false); lockErr != nil {
			if errors.Is(lockErr, quotamod.ErrQuotaAccountLocked) {
				return nil, ErrImageAccountLocked
			}
			return nil, lockErr
		}
	}

	return item, nil
}

func (s *Service) GetImageByStorageID(ctx context.Context, storageID uuid.UUID) (*ent.ImageEntity, error) {
	item, err := s.image.GetImageByStorageID(ctx, storageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrImageNotFound
		}
		return nil, err
	}
	return s.GetImage(ctx, item.ID)
}

func (s *Service) GetPresignURL(ctx context.Context, imageID uuid.UUID) (string, error) {
	item, err := s.GetImage(ctx, imageID)
	if err != nil {
		return "", err
	}

	storage, err := s.store.GetStorageByID(ctx, item.StorageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrImageURLNotFound
		}
		return "", err
	}
	if storage.URL == nil || *storage.URL == "" {
		return "", ErrImageURLNotFound
	}

	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY"))
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET_NAME"))
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return "", fmt.Errorf("missing S3 config: require S3_ENDPOINT, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY, S3_BUCKET_NAME")
	}

	useSSL := false
	if raw := strings.TrimSpace(os.Getenv("S3_USE_SSL")); raw != "" {
		useSSL = strings.EqualFold(raw, "true")
	}
	if raw := strings.TrimSpace(os.Getenv("S3_USE_S_S_L")); raw != "" {
		useSSL = strings.EqualFold(raw, "true")
	}

	presignSeconds := 900
	if raw := strings.TrimSpace(os.Getenv("S3_PRESIGN_EXPIRE_SECONDS")); raw != "" {
		if v, convErr := strconv.Atoi(raw); convErr == nil && v > 0 {
			presignSeconds = v
		}
	}

	objectPath := ""
	if storage.Path != nil && strings.TrimSpace(*storage.Path) != "" {
		objectPath = strings.TrimPrefix(strings.TrimSpace(*storage.Path), "/")
	} else {
		u, parseErr := url.Parse(strings.TrimSpace(*storage.URL))
		if parseErr != nil {
			return "", ErrImageURLNotFound
		}
		prefix := "/" + strings.Trim(bucket, "/") + "/"
		if !strings.HasPrefix(u.Path, prefix) {
			return "", ErrImageURLNotFound
		}
		objectPath = strings.TrimPrefix(u.Path, prefix)
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return "", err
	}

	presigned, err := client.PresignedGetObject(ctx, bucket, objectPath, time.Duration(presignSeconds)*time.Second, nil)
	if err != nil {
		return "", err
	}

	return presigned.String(), nil
}
