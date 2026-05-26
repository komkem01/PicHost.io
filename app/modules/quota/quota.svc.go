package quota

import (
	"context"
	"database/sql"
	"errors"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"
	internalotel "pichost.io/internal/otel"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Options struct {
	*config.Config[Config]
	tracer   trace.Tracer
	userEnt  entitiesinf.UserEntity
	quotaEnt entitiesinf.UserQuotaEntity
	imageEnt entitiesinf.ImageEntity
	planEnt  entitiesinf.PlanSettingEntity
}

type Service struct {
	*Options
}

func newService(opt *Options) *Service {
	return &Service{Options: opt}
}

func (s *Service) getUserPlanSnapshot(ctx context.Context, userID uuid.UUID) (*ent.UserEntity, ent.PlanLimits, int64, int, error) {
	user, err := s.userEnt.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ent.PlanLimits{}, 0, 0, err
	}

	// Enforce expiry handling here too, not only at /auth/me.
	if user.Plan != ent.PlanTypeFree && user.PlanExpiresAt != nil && time.Now().After(*user.PlanExpiresAt) {
		user, err = s.userEnt.DowngradeUserPlanToFree(ctx, userID)
		if err != nil {
			return nil, ent.PlanLimits{}, 0, 0, err
		}
	}

	limits := ent.GetPlanLimits(user.Plan)
	if setting, pErr := s.planEnt.GetPlanSettingByKey(ctx, string(user.Plan)); pErr == nil {
		limits.StorageBytes = setting.StorageLimitBytes
		limits.FileSizeBytes = int64(setting.MaxUploadMB) * 1024 * 1024
		limits.MaxImages = setting.ImageLimit
		limits.AllowPrivate = setting.AllowPrivate
	} else if !errors.Is(pErr, sql.ErrNoRows) {
		return nil, ent.PlanLimits{}, 0, 0, pErr
	}

	var usedStorage int64
	var imageCount int
	quota, qErr := s.quotaEnt.GetUserQuota(ctx, userID)
	if qErr != nil {
		if !errors.Is(qErr, sql.ErrNoRows) {
			return nil, ent.PlanLimits{}, 0, 0, qErr
		}
	} else {
		usedStorage = quota.UsedStorageBytes
		imageCount = quota.ImageCount
	}

	return user, limits, usedStorage, imageCount, nil
}

// EnsureUsageAllowed blocks account usage when current usage already exceeds current plan limits.
func (s *Service) EnsureUsageAllowed(ctx context.Context, userID uuid.UUID, isGuest bool) error {
	if isGuest || userID == uuid.Nil {
		return nil
	}

	_, limits, usedStorage, imageCount, err := s.getUserPlanSnapshot(ctx, userID)
	if err != nil {
		return err
	}

	if limits.StorageBytes > 0 && usedStorage > limits.StorageBytes {
		return ErrQuotaAccountLocked
	}
	if limits.MaxImages > 0 && imageCount > limits.MaxImages {
		return ErrQuotaAccountLocked
	}

	return nil
}

// InitUserQuota creates the quota row for a newly registered user.
// Safe to call multiple times (upsert with no-op on conflict).
func (s *Service) InitUserQuota(ctx context.Context, userID uuid.UUID) error {
	ctx, span, log := internalotel.NewLogSpan(ctx, s.tracer, "InitUserQuota")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	_, err := s.quotaEnt.UpsertUserQuota(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Errf("failed to init quota for user %s: %v", userID, err)
		return err
	}

	log.Infof("quota initialized for user %s", userID)
	return nil
}

// CheckUpload validates whether a file upload is permitted under the user's plan.
// For guest uploads pass uuid.Nil as userID and the GuestPlan limits will be applied.
func (s *Service) CheckUpload(ctx context.Context, userID uuid.UUID, isGuest bool, fileSize int64, mimeType string, isPrivate bool) error {
	ctx, span, log := internalotel.NewLogSpan(ctx, s.tracer, "CheckUpload")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.Bool("user.is_guest", isGuest),
		attribute.Int64("file.size_bytes", fileSize),
		attribute.String("file.mime_type", mimeType),
	)

	var limits ent.PlanLimits
	var usedStorage int64
	var imageCount int
	var err error

	if isGuest {
		limits = ent.GuestPlan
		// Guest quota is not tracked in DB — only per-request file limits apply.
		usedStorage = 0
		imageCount = 0
	} else {
		if err := s.EnsureUsageAllowed(ctx, userID, false); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			if errors.Is(err, ErrQuotaAccountLocked) {
				return ErrQuotaAccountLocked
			}
			return err
		}

		_, limits, usedStorage, imageCount, err = s.getUserPlanSnapshot(ctx, userID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	// Check private image permission.
	if isPrivate && !limits.AllowPrivate {
		log.Warnf("private images not allowed on plan")
		return ErrQuotaPrivateNotAllowed
	}

	// Check MIME type.
	if limits.AllowedMIMEs != nil && !mimeAllowed(mimeType, limits.AllowedMIMEs) {
		log.Warnf("mime type %q not allowed on plan", mimeType)
		return ErrQuotaMIMENotAllowed
	}

	// Check per-file size.
	if limits.FileSizeBytes > 0 && fileSize > limits.FileSizeBytes {
		log.Warnf("file size %d exceeds plan limit %d", fileSize, limits.FileSizeBytes)
		return ErrQuotaFileTooLarge
	}

	// Check total storage quota.
	if limits.StorageBytes > 0 && usedStorage+fileSize > limits.StorageBytes {
		log.Warnf("storage quota would be exceeded: used=%d + file=%d > limit=%d", usedStorage, fileSize, limits.StorageBytes)
		return ErrQuotaStorageFull
	}

	// Check image count.
	if limits.MaxImages > 0 && imageCount >= limits.MaxImages {
		log.Warnf("image count limit reached: count=%d, limit=%d", imageCount, limits.MaxImages)
		return ErrQuotaImageLimitReached
	}

	return nil
}

// ConsumeUpload increments the quota counters after a successful upload.
// Should only be called for authenticated (non-guest) users.
func (s *Service) ConsumeUpload(ctx context.Context, userID uuid.UUID, fileSize int64) error {
	ctx, span, log := internalotel.NewLogSpan(ctx, s.tracer, "ConsumeUpload")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.Int64("file.size_bytes", fileSize),
	)

	_, err := s.quotaEnt.UpsertUserQuota(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	_, err = s.quotaEnt.AddToUserQuota(ctx, userID, entitiesdto.AddToUserQuota{
		StorageDelta:    fileSize,
		ImageCountDelta: 1,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Errf("failed to consume quota for user %s: %v", userID, err)
		return err
	}

	log.Infof("quota consumed: user=%s, bytes=%d", userID, fileSize)
	return nil
}

// ReleaseStorage decrements the quota counters when an image is deleted.
func (s *Service) ReleaseStorage(ctx context.Context, userID uuid.UUID, fileSize int64) error {
	ctx, span, log := internalotel.NewLogSpan(ctx, s.tracer, "ReleaseStorage")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.Int64("file.size_bytes", fileSize),
	)

	_, err := s.quotaEnt.AddToUserQuota(ctx, userID, entitiesdto.AddToUserQuota{
		StorageDelta:    -fileSize,
		ImageCountDelta: -1,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Errf("failed to release quota for user %s: %v", userID, err)
		return err
	}

	log.Infof("quota released: user=%s, bytes=%d", userID, fileSize)
	return nil
}

// mimeAllowed checks whether a MIME type is in the allowed list.
func mimeAllowed(mime string, allowed []string) bool {
	for _, a := range allowed {
		if a == mime {
			return true
		}
	}
	return false
}
