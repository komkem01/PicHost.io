package specs

import (
	"mcop/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	Conf = config.Conf[Config]
)

type Config struct {
	Version string
}

type Options struct {
	Config *config.Config[Config]
	tracer trace.Tracer
}

type Module struct {
	tracer  trace.Tracer
	version string
}

func New(conf *config.Config[Config]) *Module {
	tracer := otel.Tracer("mcop.storage.specs")

	return &Module{
		tracer:  tracer,
		version: conf.Val.Version,
	}
}

func (mod *Module) Version() string {
	return mod.version
}
