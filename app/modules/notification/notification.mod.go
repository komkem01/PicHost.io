package notification

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
)

type Module struct {
	Svc *Service
	Ctl *Controller
}

type Config struct{}

func New(conf *config.Config[Config], notifEnt entitiesinf.NotificationEntity) *Module {
	tracer := otel.GetTracerProvider().Tracer(conf.AppName())

	svc := newService(notifEnt)
	ctl := newController(tracer, svc)
	return &Module{Svc: svc, Ctl: ctl}
}
