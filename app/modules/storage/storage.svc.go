package storage

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer trace.Tracer
	store  entitiesinf.StorageEntity
}

type Options struct {
	*config.Config[Config]
	tracer trace.Tracer
	store  entitiesinf.StorageEntity
}

func newService(opt *Options) *Service {
	return &Service{
		tracer: opt.tracer,
		store:  opt.store,
	}
}
