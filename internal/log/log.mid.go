package log

import "mcop/internal/config"

type Middleware struct {
	Config *config.Config[Option]
	Svc    *Service
}

func NewMiddleware(conf *config.Config[Option], svc *Service) *Middleware {
	return &Middleware{
		Config: conf,
		Svc:    svc,
	}
}
