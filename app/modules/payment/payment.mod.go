package payment

import (
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	WebhookSecret      string
	CheckoutBaseURL    string
	CheckoutTTLMinutes int
	// SubscriptionDays is the number of days a paid plan is active after payment confirmation.
	// Defaults to 30 if not set.
	SubscriptionDays int

	// Bank transfer / PromptPay info shown to the user on the checkout page.
	BankName          string
	BankAccountName   string
	BankAccountNumber string
	BankAccountType   string // e.g. "savings"
	PromptPayID       string // national ID or phone number
	BankLogoURL       string // optional logo image URL
}

type Module struct {
	tracer trace.Tracer
	Svc    *Service
	Ctl    *Controller
}

func New(
	conf *config.Config[Config],
	userEnt entitiesinf.UserEntity,
	planEnt entitiesinf.PlanSettingEntity,
	paymentEnt entitiesinf.PaymentTransactionEntity,
) *Module {
	tracer := otel.Tracer("pichost.io.modules.payment")
	svc := newService(&Options{
		Config:     conf,
		tracer:     tracer,
		userEnt:    userEnt,
		planEnt:    planEnt,
		paymentEnt: paymentEnt,
	})
	return &Module{
		tracer: tracer,
		Svc:    svc,
		Ctl:    newController(tracer, svc),
	}
}
