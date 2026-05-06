package admin

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Config is intentionally empty — admin needs no extra env vars.
type Config struct{}

type Module struct {
	tracer trace.Tracer
	Svc    *Service
	Ctl    *Controller
}

func New(
	_ *config.Config[Config],
	userEnt entitiesinf.UserEntity,
	quotaEnt entitiesinf.UserQuotaEntity,
	auditEnt entitiesinf.AuditEntity,
	imageEnt entitiesinf.ImageEntity,
	planEnt entitiesinf.PlanSettingEntity,
) *Module {
	tracer := otel.Tracer("pichost.io.modules.admin")
	svc := newService(userEnt, quotaEnt, imageEnt, planEnt)
	return &Module{
		tracer: tracer,
		Svc:    svc,
		Ctl:    newController(tracer, svc, auditEnt),
	}
}
