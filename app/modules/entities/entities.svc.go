package entities

import (
	"github.com/uptrace/bun"
)

type Service struct {
	db *bun.DB
}

func newService(db *bun.DB) *Service {
	return &Service{
		db: db,
	}
}
