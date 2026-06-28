package image

import (
	"context"

	"github.com/google/uuid"
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/app/modules/quota"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer   trace.Tracer
	image    entitiesinf.ImageEntity
	store    entitiesinf.StorageEntity
	quotaSvc *quota.Service
}

type Options struct {
	*config.Config[Config]
	tracer   trace.Tracer
	image    entitiesinf.ImageEntity
	store    entitiesinf.StorageEntity
	quotaSvc *quota.Service
}

func newService(opt *Options) *Service {
	return &Service{
		tracer:   opt.tracer,
		image:    opt.image,
		store:    opt.store,
		quotaSvc: opt.quotaSvc,
	}
}

// IsFreeUser delegates to quotaSvc to check if a user is on the Free plan.
func (s *Service) IsFreeUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	return s.quotaSvc.IsFreePlan(ctx, userID)
}

