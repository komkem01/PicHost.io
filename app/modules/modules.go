package modules

import (
	"log/slog"
	"sync"

	"pichost.io/app/modules/admin"
	"pichost.io/app/modules/audit"
	"pichost.io/app/modules/auth"
	"pichost.io/app/modules/entities"
	"pichost.io/app/modules/image"
	"pichost.io/app/modules/mailer"
	"pichost.io/app/modules/notification"
	"pichost.io/app/modules/payment"
	"pichost.io/app/modules/quota"
	"pichost.io/app/modules/sentry"
	"pichost.io/app/modules/specs"
	"pichost.io/app/modules/storage"
	"pichost.io/app/modules/users"
	"pichost.io/internal/config"
	"pichost.io/internal/database"
	"pichost.io/internal/log"
	"pichost.io/internal/otel/collector"

	appConf "pichost.io/config"
	// "pichost.io/app/modules/kafka"
)

type Modules struct {
	Conf         *config.Module[appConf.Config]
	Specs        *specs.Module
	Log          *log.Module
	OTEL         *collector.Module
	Sentry       *sentry.Module
	DB           *database.DatabaseModule
	ENT          *entities.Module
	Mailer       *mailer.Module
	Auth         *auth.Module
	Admin        *admin.Module
	Users        *users.Module
	Storage      *storage.Module
	Image        *image.Module
	Quota        *quota.Module
	Payment      *payment.Module
	Audit        *audit.Module
	Notification *notification.Module
	// Kafka *kafka.Module
}

func modulesInit() {

	confMod := config.New(&appConf.App)
	specsMod := specs.New(config.Conf[specs.Config](confMod.Svc))
	conf := confMod.Svc.Config()

	logMod := log.New(config.Conf[log.Option](confMod.Svc))
	otel := collector.New(config.Conf[collector.Config](confMod.Svc))
	log := log.With(slog.String("module", "modules"))

	sentryMod := sentry.New(config.Conf[sentry.Config](confMod.Svc))

	db := database.New(conf.Database.Sql)
	entitiesMod := entities.New(db.Svc.DB())
	mailerMod := mailer.New(config.Conf[mailer.Config](confMod.Svc))

	authMod := auth.New(config.Conf[auth.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	authMod.SetMailer(mailerMod.Svc)
	authMod.SetImageEntity(entitiesMod.Svc)

	usersMod := users.New(config.Conf[users.Config](confMod.Svc), entitiesMod.Svc)
	storageMod := storage.New(config.Conf[storage.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	quotaMod := quota.New(config.Conf[quota.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	imageMod := image.New(config.Conf[image.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, quotaMod.Svc)
	paymentMod := payment.New(config.Conf[payment.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	paymentMod.SetMailer(mailerMod.Svc)
	paymentMod.SetNotificationEntity(entitiesMod.Svc)
	authMod.SetEntitlementActivator(paymentMod.Svc)

	notificationMod := notification.New(config.Conf[notification.Config](confMod.Svc), entitiesMod.Svc)

	storageMod.SetImageService(imageMod.Svc)
	authMod.SetAuditEntity(entitiesMod.Svc)
	storageMod.SetAuditEntity(entitiesMod.Svc)
	paymentMod.SetAuditEntity(entitiesMod.Svc)
	imageMod.SetAuditEntity(entitiesMod.Svc)
	adminMod := admin.New(config.Conf[admin.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)

	auditMod := audit.New(config.Conf[audit.Config](confMod.Svc), entitiesMod.Svc)
	// kafka := kafka.New(&conf.Kafka)
	mod = &Modules{
		Conf:         confMod,
		Specs:        specsMod,
		Log:          logMod,
		OTEL:         otel,
		Sentry:       sentryMod,
		DB:           db,
		ENT:          entitiesMod,
		Mailer:       mailerMod,
		Auth:         authMod,
		Admin:        adminMod,
		Users:        usersMod,
		Storage:      storageMod,
		Image:        imageMod,
		Quota:        quotaMod,
		Payment:      paymentMod,
		Audit:        auditMod,
		Notification: notificationMod,
	}

	log.Infof("all modules initialized")
}

var (
	once sync.Once
	mod  *Modules
)

func Get() *Modules {
	once.Do(modulesInit)

	return mod
}
