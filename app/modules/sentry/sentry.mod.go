package sentry

import (
	"context"
	"time"

	"mcop/internal/provider"

	"mcop/internal/config"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var _ provider.Close = (*Service)(nil)

type Module struct {
	tracer trace.Tracer
	Svc    *Service
}

type Options struct {
	Config *config.Config[Config]
	tracer trace.Tracer
}

type Service struct {
	*Options
}

type Config struct {
	DSN string
}

func New(conf *config.Config[Config]) *Module {
	tracer := otel.Tracer("mcop.storage.sentry")
	svc := newService(&Options{
		Config: conf,
		tracer: tracer,
	})

	return &Module{
		tracer: tracer,
		Svc:    svc,
	}
}

// Close implements provider.Close.
func (s *Service) Close(ctx context.Context) error {
	sentry.Flush(2 * time.Second)
	return nil
}
