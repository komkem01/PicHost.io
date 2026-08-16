package routes

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"pichost.io/app/modules"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

// buildAllowOriginFunc parses the comma-separated CORS allowlist into a matcher.
//
// Entries are exact origins; a leading "*." (e.g. "*.vercel.app") matches any
// subdomain of that suffix. A bare "*" opens every origin, which cannot be
// safely combined with AllowCredentials and is therefore rejected in production.
// An empty list is rejected too: silently falling back to a localhost default
// would block the real frontend with an error that looks like a CORS bug.
func buildAllowOriginFunc(allowedOriginsRaw, environment string) (func(string) bool, error) {
	allowedOrigins := make(map[string]bool)
	wildcardSuffixes := make([]string, 0)
	allowAll := false

	for _, o := range strings.Split(allowedOriginsRaw, ",") {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		switch {
		case o == "":
			// An empty entry is not an instruction to allow everything — skip it.
			continue
		case o == "*":
			allowAll = true
		case strings.HasPrefix(o, "*."):
			wildcardSuffixes = append(wildcardSuffixes, strings.ToLower(o[1:]))
		default:
			allowedOrigins[strings.ToLower(o)] = true
		}
	}

	if allowAll && environment == "production" {
		return nil, errors.New("CORS_ALLOWED_ORIGINS=* is not allowed in production: list the exact frontend origins instead")
	}
	if !allowAll && len(allowedOrigins) == 0 && len(wildcardSuffixes) == 0 {
		return nil, errors.New("CORS_ALLOWED_ORIGINS is empty: set it to the frontend origin(s), e.g. https://pichost-web.vercel.app")
	}

	return func(origin string) bool {
		cleanOrigin := strings.ToLower(strings.TrimRight(strings.TrimSpace(origin), "/"))
		if cleanOrigin == "" {
			return false
		}
		if allowAll {
			return true
		}
		if allowedOrigins[cleanOrigin] {
			return true
		}
		for _, suffix := range wildcardSuffixes {
			if strings.HasSuffix(cleanOrigin, suffix) {
				return true
			}
		}
		return false
	}, nil
}

func Router(app *gin.Engine, mod *modules.Modules) {
	// 0.1: Restrict CORS origins using an explicit allowlist.
	allowOrigin, err := buildAllowOriginFunc(
		mod.Conf.Svc.Config().CorsAllowedOrigins,
		mod.Conf.Svc.Config().Environment,
	)
	if err != nil {
		panic(err.Error())
	}

	// 0.2: CORS middleware MUST be applied FIRST before any routes/telemetry so preflight OPTIONS requests return immediately without delay
	app.Use(cors.New(cors.Config{
		AllowOriginFunc:        allowOrigin,
		AllowMethods:           []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:           []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With", "X-Trace-ID"},
		ExposeHeaders:          []string{"X-Trace-ID", "Content-Disposition", "Content-Length"},
		AllowCredentials:       true,
		AllowBrowserExtensions: true,
		AllowWebSockets:        true,
		MaxAge:                 12 * time.Hour,
	}))

	// 0.3: Telemetry & Tracing middleware
	app.Use(otelgin.Middleware(mod.Conf.Svc.Config().AppName),
		func(ctx *gin.Context) {
			spanCtx := trace.SpanContextFromContext(ctx.Request.Context())
			if spanCtx.IsValid() {
				ctx.Header("X-Trace-ID", spanCtx.TraceID().String())
			}
			ctx.Next()
		},
	)

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
	apiNotifications(app.Group("/api/v1"), mod)
}

