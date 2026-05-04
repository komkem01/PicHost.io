package modules

import (
	"log/slog"
	"sync"

	"mcop/app/modules/entities"
	"mcop/app/modules/example"
	"mcop/app/modules/sentry"
	"mcop/app/modules/specs"
	"mcop/internal/config"
	"mcop/internal/database"
	"mcop/internal/log"
	"mcop/internal/otel/collector"

	exampletwo "mcop/app/modules/example-two"

	appConf "mcop/config"
	// "mcop/app/modules/kafka"
)

type Modules struct {
	Conf   *config.Module[appConf.Config]
	Specs  *specs.Module
	Log    *log.Module
	OTEL   *collector.Module
	Sentry *sentry.Module
	DB     *database.DatabaseModule
	ENT    *entities.Module
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
