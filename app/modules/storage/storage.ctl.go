package storage

import (
	"go.opentelemetry.io/otel/trace"

	imagemod "pichost.io/app/modules/image"
)

type Controller struct {
	tracer trace.Tracer
	svc    *Service
	imgSvc *imagemod.Service
}

func newController(trace trace.Tracer, svc *Service) *Controller {
	return &Controller{
		tracer: trace,
		svc:    svc,
	}
}
