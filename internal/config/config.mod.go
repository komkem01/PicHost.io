package config

type Module[T any] struct {
	Svc *Service[T]
}

func New[T any](dConf *T) *Module[T] {
	return &Module[T]{
		Svc: newService(dConf),
	}
}
