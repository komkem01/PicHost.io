package mailer

import (
	"pichost.io/internal/config"
)

type Module struct {
	Svc *Service
}

func New(cfg *config.Config[Config]) *Module {
	svc := NewService(cfg)
	return &Module{
		Svc: svc,
	}
}
