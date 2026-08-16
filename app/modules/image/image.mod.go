package image

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/app/modules/quota"
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

func New(
	conf *config.Config[Config],
	imageEnt entitiesinf.ImageEntity,
	storageEnt entitiesinf.StorageEntity,
	quotaSvc *quota.Service,
) *Module {
	tracer := otel.Tracer("pichost.io.modules.image")
	svc := newService(&Options{
		Config:   conf,
		tracer:   tracer,
		image:    imageEnt,
		store:    storageEnt,
		quotaSvc: quotaSvc,
	})

	return &Module{
		tracer: tracer,
		Svc:    svc,
		Ctl:    newController(tracer, svc),
	}
}

func (m *Module) SetAuditEntity(auditEnt entitiesinf.AuditEntity) {
	m.Ctl.auditEnt = auditEnt
}

