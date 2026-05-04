package image

import (
	"context"
	"database/sql"
	"errors"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	quotamod "pichost.io/app/modules/quota"
	internalotel "pichost.io/internal/otel"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type CreateImageSvcRequest struct {
	UserID    uuid.UUID // uuid.Nil for guest
	StorageID uuid.UUID
	IsPrivate bool
	IsGuest   bool
}

// CreateImage checks quota, creates the ImageEntity, then consumes quota.
func (s *Service) CreateImage(ctx context.Context, req CreateImageSvcRequest) (*ent.ImageEntity, error) {
	ctx, span, log := internalotel.NewLogSpan(ctx, s.tracer, "CreateImage")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", req.UserID.String()),
		attribute.Bool("user.is_guest", req.IsGuest),
		attribute.String("storage.id", req.StorageID.String()),
	)

	// Fetch the storage record to know file size and MIME type.
	storage, err := s.store.GetStorageByID(ctx, req.StorageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			span.SetStatus(codes.Error, "storage not found")
			return nil, ErrImageStorageNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	mimeType := ""
	if storage.MIMEType != nil {
		mimeType = *storage.MIMEType
	}

	// Check quota before creating the record.
	if err := s.quotaSvc.CheckUpload(ctx, req.UserID, req.IsGuest, storage.FileSize, mimeType); err != nil {
		log.Warnf("quota check failed for user %s: %v", req.UserID, err)
		span.RecordError(err)

		switch {
		case errors.Is(err, quotamod.ErrQuotaFileTooLarge):
			return nil, ErrImageFileTooLarge
		case errors.Is(err, quotamod.ErrQuotaStorageFull):
			return nil, ErrImageStorageFull
		case errors.Is(err, quotamod.ErrQuotaImageLimitReached):
			return nil, ErrImageLimitReached
		case errors.Is(err, quotamod.ErrQuotaMIMENotAllowed):
			return nil, ErrImageMIMENotAllowed
		default:
			return nil, err
		}
	}

	// Set expiry for guest images.
	var expiresAtStr *string
	if req.IsGuest {
		limits := ent.GuestPlan
		exp := time.Now().Add(time.Duration(limits.RetentionHours) * time.Hour).Format(time.RFC3339)
		expiresAtStr = &exp
	}

	var userIDStr *string
	if !req.IsGuest && req.UserID != uuid.Nil {
		s := req.UserID.String()
		userIDStr = &s
	}
	storageIDStr := req.StorageID.String()
	isPrivate := req.IsPrivate && !req.IsGuest // guests cannot create private images

	image, err := s.image.CreateImage(ctx, entitiesdto.CreateImage{
		UserID:    userIDStr,
		StorageID: &storageIDStr,
		IsPrivate: &isPrivate,
		ExpiresAt: expiresAtStr,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Consume quota for authenticated users.
	if !req.IsGuest && req.UserID != uuid.Nil {
		if err := s.quotaSvc.ConsumeUpload(ctx, req.UserID, storage.FileSize); err != nil {
			// Non-fatal: log but don't fail the request.
			log.Warnf("failed to consume quota for user %s: %v", req.UserID, err)
		}
	}

	log.Infof("image created: id=%s user=%s", image.ID, req.UserID)
	return image, nil
}
