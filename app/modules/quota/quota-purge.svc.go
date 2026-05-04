package quota

import (
	"context"
	"time"

	internalotel "pichost.io/internal/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// PurgeExpiredGuestImages deletes all image records whose expires_at is in the past.
// Call this from a scheduled job or startup routine for batch cleanup.
// For real-time expiry enforcement, GetImage already handles lazy deletion.
func (s *Service) PurgeExpiredGuestImages(ctx context.Context) (int, error) {
	ctx, span, log := internalotel.NewLogSpan(ctx, s.tracer, "PurgeExpiredGuestImages")
	defer span.End()

	now := time.Now()
	span.SetAttributes(attribute.String("purge.before", now.Format(time.RFC3339)))

	images, err := s.imageEnt.ListExpiredImages(ctx, now)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		log.Errf("failed to list expired images: %v", err)
		return 0, err
	}

	deleted := 0
	for _, img := range images {
		if err := s.imageEnt.DeleteImage(ctx, img.ID); err != nil {
			log.Warnf("failed to delete expired image %s: %v", img.ID, err)
			continue
		}
		deleted++
		log.Infof("purged expired guest image %s (expired at %s)", img.ID, img.ExpiresAt)
	}

	span.SetAttributes(attribute.Int("purge.deleted_count", deleted))
	log.Infof("purge complete: deleted=%d / total=%d", deleted, len(images))
	return deleted, nil
}
