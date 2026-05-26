package modules

import (
	"log/slog"
	"sync"

	"pichost.io/app/modules/admin"
	"pichost.io/app/modules/auth"
	"pichost.io/app/modules/entities"
	"pichost.io/app/modules/example"
	"pichost.io/app/modules/image"
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

	exampletwo "pichost.io/app/modules/example-two"

	appConf "pichost.io/config"
	// "pichost.io/app/modules/kafka"
)

type Modules struct {
	Conf    *config.Module[appConf.Config]
	Specs   *specs.Module
	Log     *log.Module
	OTEL    *collector.Module
	Sentry  *sentry.Module
	DB      *database.DatabaseModule
	ENT     *entities.Module
	Auth    *auth.Module
	Admin   *admin.Module
	Users   *users.Module
	Storage *storage.Module
	Image   *image.Module
	Quota   *quota.Module
	Payment *payment.Module
	// Kafka *kafka.Module
	Example  *example.Module
	Example2 *exampletwo.Module
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
	authMod := auth.New(config.Conf[auth.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	usersMod := users.New(config.Conf[users.Config](confMod.Svc), entitiesMod.Svc)
	storageMod := storage.New(config.Conf[storage.Config](confMod.Svc), entitiesMod.Svc)
	quotaMod := quota.New(config.Conf[quota.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	imageMod := image.New(config.Conf[image.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, quotaMod.Svc)
	paymentMod := payment.New(config.Conf[payment.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	storageMod.SetImageService(imageMod.Svc)
	authMod.SetAuditEntity(entitiesMod.Svc)
	storageMod.SetAuditEntity(entitiesMod.Svc)
	adminMod := admin.New(config.Conf[admin.Config](confMod.Svc), entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc, entitiesMod.Svc)
	exampleMod := example.New(config.Conf[example.Config](confMod.Svc), entitiesMod.Svc)
	exampleMod2 := exampletwo.New(config.Conf[exampletwo.Config](confMod.Svc), entitiesMod.Svc)
	// kafka := kafka.New(&conf.Kafka)
	mod = &Modules{
		Conf:     confMod,
		Specs:    specsMod,
		Log:      logMod,
		OTEL:     otel,
		Sentry:   sentryMod,
		DB:       db,
		ENT:      entitiesMod,
		Auth:     authMod,
		Admin:    adminMod,
		Users:    usersMod,
		Storage:  storageMod,
		Image:    imageMod,
		Quota:    quotaMod,
		Payment:  paymentMod,
		Example:  exampleMod,
		Example2: exampleMod2,
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
