package emample

type Module struct {
	Svc *Service
	Ctl *Controller
}

type (
	Service    struct{}
	Controller struct {
		svc *Service
	}
)

func New() *Module {
	svc := newService()
	return &Module{
		Svc: newService(),
		Ctl: newController(svc),
	}
}

func newService() *Service {
	return &Service{}
}

func newController(svc *Service) *Controller {
	return &Controller{
		svc: svc,
	}
}
