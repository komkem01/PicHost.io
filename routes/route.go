package routes

import (
	"net/http"

	"pichost.io/app/modules"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

func Router(app *gin.Engine, mod *modules.Modules) {
	app.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})
	app.GET("/p/:code", mod.Storage.Ctl.OpenPublicByCode)
	app.HEAD("/p/:code", mod.Storage.Ctl.OpenPublicByCode)
	app.GET("/i/:id", mod.Storage.Ctl.OpenPublic)
	app.HEAD("/i/:id", mod.Storage.Ctl.OpenPublic)

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
		AllowOriginFunc: func(origin string) bool {
			// Allow all origins but reflect the actual origin (required when AllowCredentials=true)
			return true
		},
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:           []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:          []string{"X-Trace-ID"},
		AllowCredentials:       true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
	}))

	api(app.Group("/api/v1"), mod)
	apiUser(app.Group("/api/v1"), mod)
	apiStorage(app.Group("/api/v1"), mod)
	apiImage(app.Group("/api/v1"), mod)
	apiPublic(app.Group("/api/v1"), mod)
	apiAuth(app.Group("/api/v1"), mod)
}
