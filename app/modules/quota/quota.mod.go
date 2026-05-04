package quota

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Module struct {
	tracer trace.Tracer
	Svc    *Service
}

type Config struct{}

func New(
	conf *config.Config[Config],
	userEnt entitiesinf.UserEntity,
	quotaEnt entitiesinf.UserQuotaEntity,
	imageEnt entitiesinf.ImageEntity,
) *Module {
	tracer := otel.Tracer("pichost.io.modules.quota")
	svc := newService(&Options{
		Config:   conf,
		tracer:   tracer,
		userEnt:  userEnt,
		quotaEnt: quotaEnt,
		imageEnt: imageEnt,
	})

	return &Module{
		tracer: tracer,
		Svc:    svc,
	}
}
