package storage

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Module struct {
	tracer trace.Tracer
	Svc    *Service
	Ctl    *Controller
}

type Config struct{}

func New(conf *config.Config[Config], storageEnt entitiesinf.StorageEntity) *Module {
	tracer := otel.Tracer("pichost.io.modules.storage")
	svc := newService(&Options{
		Config: conf,
		tracer: tracer,
		store:  storageEnt,
	})

	return &Module{
		tracer: tracer,
		Svc:    svc,
		Ctl:    newController(tracer, svc),
	}
}
