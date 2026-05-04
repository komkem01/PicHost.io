package entities

import "github.com/uptrace/bun"

type Module struct {
	Svc *Service
}

func New(db *bun.DB) *Module {
	return &Module{
		Svc: newService(db),
	}
}
