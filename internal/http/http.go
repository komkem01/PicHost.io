package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"mcop/app/modules"
	"mcop/internal/log"
	"mcop/internal/provider"
	"mcop/routes"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type ginLogger struct{}

func (l *ginLogger) Write(p []byte) (n int, err error) {
	log := log.With(slog.String("gin", "logger"))
	log.Infof("%s", string(p))
	return len(p), nil
}

// D is the main function for the HTTP server.
func D(isHTTPS bool) func(_ *cobra.Command, _ []string) {
	return func(_ *cobra.Command, _ []string) {
		ctx, cancel := NotifyContext()
		mod := modules.Get()
		conf := mod.Conf.Svc.Config()

		srv := serve(mod)
		log := log.With(log.String("gin", "logger"))

		go func() {
			srv.Addr = fmt.Sprintf("0.0.0.0:%d", conf.Port)
			log.Infof("Gin is running on %s.", srv.Addr)
			if isHTTPS {
				pk := conf.SslPrivatePath
				cert := conf.SslCertPath
				if err := srv.ListenAndServeTLS(cert, pk); !errors.Is(err, nil) && !errors.Is(err, http.ErrServerClosed) {
					log.With(log.Error(err)).Errf("Gin was failed to start %s.", srv.Addr)
					os.Exit(1)
				}
			} else if err := srv.ListenAndServe(); !errors.Is(err, nil) && !errors.Is(err, http.ErrServerClosed) {
				log.With(log.Error(err)).Errf("Gin was failed to start %s.", srv.Addr)
				os.Exit(1)
			}
		}()

		<-ctx.Done()
		cancel()
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(timeoutCtx); err != nil {
			log.With(log.Error(err)).Errf("Gin was failed to shutdown.")
			os.Exit(1)
		}

		log.Infof("Gin was successful shutdown.")
	}
}

func serve(mod *modules.Modules) *http.Server {
	conf := mod.Conf.Svc.Config()
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	app := gin.New(
		func(e *gin.Engine) {
			e.UseH2C = true
			e.RemoveExtraSlash = true
		})

	if conf.Debug {
		app.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Output: new(ginLogger),
			Formatter: func(params gin.LogFormatterParams) string {
				if params.ErrorMessage != "" {
					return params.ErrorMessage
				}
				return fmt.Sprintf("%d %s %s %s %s %s %s",
					params.StatusCode,
					params.Method,
					params.Path,
					params.Latency,
					params.ClientIP,
					params.Request.Proto,
					params.Request.UserAgent(),
				)
			},
		}), gin.Recovery())
		pprof.Register(app)
	}
	routes.Router(app, mod)

	srv := &http.Server{
		Handler: app,
	}
	srv.SetKeepAlivesEnabled(true)

	srv.RegisterOnShutdown(func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pv := provider.Config(modules.Map())
		pv.Close(timeoutCtx)
	})
	return srv
}
