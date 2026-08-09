package routes

import (
	"net/http"
	"strings"

	"pichost.io/app/modules"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

func Router(app *gin.Engine, mod *modules.Modules) {
	// 0.4: Global middleware applied FIRST before any routes
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

	// 0.3: Restrict CORS origins using exact match allowlist from config
	allowedOriginsRaw := mod.Conf.Svc.Config().CorsAllowedOrigins
	allowedOrigins := make(map[string]bool)
	for _, o := range strings.Split(allowedOriginsRaw, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowedOrigins[o] = true
		}
	}

	app.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return allowedOrigins[origin]
		},
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:           []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:          []string{"X-Trace-ID"},
		AllowCredentials:       true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
	}))

	// Public static routes
	app.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	app.GET("/readyz", func(ctx *gin.Context) {
		checks := gin.H{}
		isHealthy := true

		if mod.DB != nil && mod.DB.Svc != nil {
			if err := mod.DB.Svc.DB().PingContext(ctx.Request.Context()); err != nil {
				checks["database"] = "down"
				isHealthy = false
			} else {
				checks["database"] = "up"
			}
		} else {
			checks["database"] = "unknown"
		}

		if isHealthy {
			ctx.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"checks": checks,
			})
		} else {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "degraded",
				"checks": checks,
			})
		}
	})

	app.GET("/p/:code", mod.Storage.Ctl.OpenPublicByCode)
	app.HEAD("/p/:code", mod.Storage.Ctl.OpenPublicByCode)
	app.GET("/i/:id", mod.Storage.Ctl.OpenPublic)
	app.HEAD("/i/:id", mod.Storage.Ctl.OpenPublic)

	// API routes
	apiStorage(app.Group("/api/v1"), mod)
	apiImage(app.Group("/api/v1"), mod)
	apiPublic(app.Group("/api/v1"), mod)
	apiAuth(app.Group("/api/v1"), mod)
	apiBilling(app.Group("/api/v1"), mod)
	apiAdmin(app.Group("/api/v1"), mod)
}

