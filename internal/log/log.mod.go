package log

import "pichost.io/internal/config"

type Module struct {
	Svc *Service
	Mid *Middleware
}

func New(conf *config.Config[Option]) *Module {
	svc := newService(conf)
	mid := NewMiddleware(conf, svc)
	return &Module{
		Svc: svc,
		Mid: mid,
	}
}
