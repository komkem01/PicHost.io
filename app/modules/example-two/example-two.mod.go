package exampletwo

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"
)

type (
	Module struct {
		Svc *Service
		Ctl *Controller
	}
	Service    struct{}
	Controller struct{}

	Config struct{}
)

func New(conf *config.Config[Config], db entitiesinf.ExampleEntity) *Module {
	return &Module{
		Svc: &Service{},
		Ctl: &Controller{},
	}
}
