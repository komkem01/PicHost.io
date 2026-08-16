package audit

import (
	"context"

	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
)

type Module struct {
	Svc *Service
	Ctl *Controller
}

type Config struct{}

func New(conf *config.Config[Config], auditEnt entitiesinf.AuditEntity) *Module {
	tracer := otel.GetTracerProvider().Tracer(conf.AppName())

	svc := newService(auditEnt)
	svc.StartRetentionCleaner(context.Background(), 30)

	ctl := newController(tracer, svc)
	return &Module{Svc: svc, Ctl: ctl}
}
