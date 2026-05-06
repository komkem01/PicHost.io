package auth

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Module struct {
	tracer trace.Tracer
	Svc    *Service
	Ctl    *Controller
}

type Config struct {
	JWTSecret              string
	AccessTokenTTLSeconds  int
	RefreshTokenTTLSeconds int
	JWTIssuer              string
	RefreshCookieName      string
	RefreshCookieDomain    string
	RefreshCookieSecure    bool
	GoogleClientID         string
	GoogleClientSecret     string
	GoogleRedirectURL      string
	GoogleStateTTLSeconds  int
	FrontendURL            string
}

func New(conf *config.Config[Config], userEnt entitiesinf.UserEntity, authEnt entitiesinf.AuthEntity, quotaEnt entitiesinf.UserQuotaEntity, planEnt entitiesinf.PlanSettingEntity) *Module {
	tracer := otel.Tracer("pichost.io.modules.auth")
	if conf.Val.AccessTokenTTLSeconds <= 0 {
		conf.Val.AccessTokenTTLSeconds = 300
	}
	if conf.Val.RefreshTokenTTLSeconds <= 0 {
		conf.Val.RefreshTokenTTLSeconds = 2592000
	}
	if conf.Val.JWTIssuer == "" {
		conf.Val.JWTIssuer = "pichost.io"
	}
	if conf.Val.RefreshCookieName == "" {
		conf.Val.RefreshCookieName = "refresh_token"
	}
	if conf.Val.JWTSecret == "" {
		conf.Val.JWTSecret = "change-me-in-production"
	}
	if conf.Val.GoogleStateTTLSeconds <= 0 {
		conf.Val.GoogleStateTTLSeconds = 300
	}
	if conf.Val.FrontendURL == "" {
		conf.Val.FrontendURL = "http://localhost:3000"
	}

	svc := newService(&Options{
		Config:   conf,
		tracer:   tracer,
		user:     userEnt,
		auth:     authEnt,
		quotaEnt: quotaEnt,
		planEnt:  planEnt,
	})

	return &Module{
		tracer: tracer,
		Svc:    svc,
		Ctl:    newController(tracer, svc),
	}
}

// SetAuditEntity injects the audit logger into the auth controller (fire-and-forget).
// Must be called after New().
func (m *Module) SetAuditEntity(auditEnt entitiesinf.AuditEntity) {
	m.Ctl.auditEnt = auditEnt
}
