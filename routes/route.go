package routes

import (
	"net/http"

	"mcop/app/modules"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

func Router(app *gin.Engine, mod *modules.Modules) {
	app.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	app.Use(otelgin.Middleware(mod.Conf.Svc.Config().AppName),
		// Middleware add trace id to response header
		func(ctx *gin.Context) {
			spanCtx := trace.SpanContextFromContext(ctx.Request.Context())
			if spanCtx.IsValid() {
				ctx.Header("X-Trace-ID", spanCtx.TraceID().String())
			}
			ctx.Next()
		},
	)

	app.Use(cors.New(cors.Config{
		AllowAllOrigins:        true,
		AllowMethods:           []string{"*"},
		AllowHeaders:           []string{"*"},
		AllowCredentials:       true,
		AllowWildcard:          true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
		AllowFiles:             false,
	}))

	api(app.Group("/api/v1"), mod)
}
