package config

import (
	"pichost.io/app/modules/admin"
	"pichost.io/app/modules/auth"
	"pichost.io/app/modules/image"
	"pichost.io/app/modules/quota"

	"pichost.io/app/modules/example"
	exampletwo "pichost.io/app/modules/example-two"
	"pichost.io/app/modules/sentry"
	"pichost.io/app/modules/specs"
	"pichost.io/app/modules/storage"
	"pichost.io/app/modules/users"
	"pichost.io/internal/kafka"
	"pichost.io/internal/log"
	"pichost.io/internal/otel/collector"
)

// Config is a struct that contains all the configuration of the application.
type Config struct {
	Database Database

	AppName     string
	AppKey      string
	Environment string
	Specs       specs.Config
	Debug       bool

	Port           int
	HttpJsonNaming string

	SslCaPath      string
	SslPrivatePath string
	SslCertPath    string

	Otel   collector.Config
	Sentry sentry.Config

	Kafka kafka.Config
	Log   log.Option

	Example example.Config

	ExampleTwo exampletwo.Config

	Auth    auth.Config
	User    users.Config
	Storage storage.Config
	Image   image.Config
	Quota   quota.Config
	Admin   admin.Config
}

var App = Config{
	Specs: specs.Config{
		Version: "v1",
	},
	Database: database,
	Kafka:    kafkaConf,

	AppName: "go_app",
	Port:    8080,
	AppKey:  "secret",
	Debug:   false,

	HttpJsonNaming: "snake_case",

	SslCaPath:      "pichost.io/cert/ca.pem",
	SslPrivatePath: "pichost.io/cert/server.pem",
	SslCertPath:    "pichost.io/cert/server-key.pem",

	Otel: collector.Config{
		CollectorEndpoint: "",
		LogMode:           "noop",
		TraceMode:         "noop",
		MetricMode:        "noop",
		TraceRatio:        0.01,
	},
	Auth: auth.Config{
		JWTSecret:              "change-me-in-production",
		AccessTokenTTLSeconds:  300,
		RefreshTokenTTLSeconds: 2592000,
		JWTIssuer:              "pichost.io",
		RefreshCookieName:      "refresh_token",
		RefreshCookieDomain:    "",
		RefreshCookieSecure:    false,
		GoogleClientID:         "",
		GoogleClientSecret:     "",
		GoogleRedirectURL:      "",
		GoogleStateTTLSeconds:  300,
		FrontendURL:            "http://localhost:3000",
	},
	User:  users.Config{},
	Image: image.Config{},
	Quota: quota.Config{},
}
