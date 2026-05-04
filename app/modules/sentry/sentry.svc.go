// Package sentry provides Sentry integration with OpenTelemetry for error tracking and performance monitoring.
package sentry

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	sentryslog "github.com/getsentry/sentry-go/slog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// newService creates and configures a new Sentry service with OpenTelemetry integration.
func newService(opts *Options) *Service {
	svc := &Service{Options: opts}
	svc.initialize()
	return svc
}

// initialize sets up Sentry with all integrations.
func (s *Service) initialize() {
	s.initializeSentrySDK()
	s.setupOpenTelemetryIntegration()
	s.setupPropagation()
	s.setupLoggingIntegration()
}

// initializeSentrySDK initializes the Sentry SDK with basic configuration.
func (s *Service) initializeSentrySDK() {
	clientOptions := s.buildClientOptions()
	if err := sentry.Init(clientOptions); err != nil {
		panic(err)
	}
}

// buildClientOptions creates Sentry client options from service configuration.
func (s *Service) buildClientOptions() sentry.ClientOptions {
	return sentry.ClientOptions{
		Dsn:              s.Config.Val.DSN,
		ServerName:       s.Config.AppName(),
		Environment:      s.Config.Environment(),
		Release:          s.Config.Version(),
		TracesSampleRate: 0.01, // Default 1% sampling
		EnableTracing:    true,
		EnableLogs:       true,
		AttachStacktrace: true,
		SendDefaultPII:   true,
		// Custom sampler to increase sampling rate for errors
		TracesSampler: CreateErrorAwareTracesSampler(),
	}
}

// setupOpenTelemetryIntegration configures OpenTelemetry span processing with Sentry.
func (s *Service) setupOpenTelemetryIntegration() {
	tp := otel.GetTracerProvider()
	if sdkProvider, ok := tp.(*sdkTrace.TracerProvider); ok {
		// Register Sentry's span processor to automatically send OpenTelemetry spans to Sentry
		sdkProvider.RegisterSpanProcessor(sentryotel.NewSentrySpanProcessor())
	}
}

// setupPropagation configures trace context propagation by combining existing and Sentry propagators.
func (s *Service) setupPropagation() {
	compositePropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
		sentryotel.NewSentryPropagator())
	otel.SetTextMapPropagator(compositePropagator)
}

// setupLoggingIntegration configures structured logging with Sentry and OpenTelemetry context.
func (s *Service) setupLoggingIntegration() {
	ctx := context.Background()
	sentryHandler := s.createSentryLogHandler(ctx)
	teeHandler := NewTeeHandler(slog.Default().Handler(), sentryHandler)
	slog.SetDefault(slog.New(teeHandler))
}

// createSentryLogHandler creates a Sentry log handler with OpenTelemetry context awareness.
func (s *Service) createSentryLogHandler(ctx context.Context) slog.Handler {
	return sentryslog.Option{
		AddSource:  true,
		EventLevel: s.getEventLevels(),
		LogLevel:   s.getLogLevels(),
		AttrFromContext: []func(ctx context.Context) []slog.Attr{
			s.extractOpenTelemetryContext,
		},
	}.NewSentryHandler(ctx)
}

// getEventLevels returns log levels that should be sent as Sentry events.
func (s *Service) getEventLevels() []slog.Level {
	return []slog.Level{slog.LevelError, sentryslog.LevelFatal}
}

// getLogLevels returns log levels that should be sent as Sentry log entries.
func (s *Service) getLogLevels() []slog.Level {
	return []slog.Level{slog.LevelWarn, slog.LevelError, sentryslog.LevelFatal}
}

// extractOpenTelemetryContext extracts OpenTelemetry trace information from context.
func (s *Service) extractOpenTelemetryContext(ctx context.Context) []slog.Attr {
	var attrs []slog.Attr

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", spanCtx.TraceID().String()),
			slog.String("span_id", spanCtx.SpanID().String()),
		)
	}

	return attrs
}
