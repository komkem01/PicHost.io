package users

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	tracer trace.Tracer
	user   entitiesinf.UserEntity
}

type Options struct {
	*config.Config[Config]
	tracer trace.Tracer
	user   entitiesinf.UserEntity
}

func newService(opt *Options) *Service {
	return &Service{
		tracer: opt.tracer,
		user:   opt.user,
	}
}
