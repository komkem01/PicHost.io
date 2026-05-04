package log

import (
	"context"
	"log/slog"
	"os"

	"mcop/internal/config"

	"go.elastic.co/ecszap"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

// Service struct
type Service struct {
	*Logger
}
type Option struct {
	Level string
}

func newService(conf *config.Config[Option]) *Service {
	zLog := initializeZapLogger(conf)
	newSLog := initializeSLogLogger(zLog, conf)

	slog.SetDefault(newSLog)
	log := defaultLogger.Load()
	log.Logger = newSLog
	return &Service{
		Logger: log,
	}
}

func initializeZapLogger(conf *config.Config[Option]) *zap.Logger {
	if conf.Environment() != "local" {
		return createProductionZapLogger(conf)
	}
	return createDevelopmentZapLogger()
}

func createProductionZapLogger(conf *config.Config[Option]) *zap.Logger {
	zapOptions := []zap.Option{zap.AddCaller()}
	if conf.Debug() {
		zapOptions = append(zapOptions, zap.Development())
	}

	encoderConfig := ecszap.NewDefaultEncoderConfig()
	core := ecszap.NewCore(encoderConfig, os.Stdout, zap.DebugLevel)
	return zap.New(core, zapOptions...)
}

func createDevelopmentZapLogger() *zap.Logger {
	zLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return zLog
}

func initializeSLogLogger(zLog *zap.Logger, conf *config.Config[Option]) *slog.Logger {
	coreHandler := zapslog.NewHandler(zLog.Core(),
		zapslog.WithName(conf.AppName()),
		zapslog.WithCaller(true),
	)

	otelHandler := zapslog.NewHandler(otelzap.NewCore(conf.AppName()),
		zapslog.WithName(conf.AppName()),
		zapslog.WithCaller(true),
	)

	return slog.New(&zapHandler{
		stdout: coreHandler,
		otel:   otelHandler,
	})
}

type zapHandler struct {
	stdout slog.Handler
	otel   slog.Handler
}

// Enabled implements slog.Handler.
func (z *zapHandler) Enabled(context.Context, slog.Level) bool {
	return true // Always enabled for simplicity; adjust as needed.
}

// Handle implements slog.Handler.
func (z *zapHandler) Handle(ctx context.Context, r slog.Record) error {
	if z.stdout.Enabled(ctx, r.Level) {
		spanCtx := trace.SpanContextFromContext(ctx)
		rc := r.Clone()
		if spanCtx.IsValid() {
			rc.Add(slog.String("TraceID", spanCtx.TraceID().String()))
			rc.Add(slog.String("SpanID", spanCtx.SpanID().String()))
		}
		z.stdout.Handle(ctx, rc)
	}
	if z.otel.Enabled(ctx, r.Level) {
		r.Add(slog.Any("ctx", ctx))
		z.otel.Handle(ctx, r)
	}
	return nil
}

// WithAttrs implements slog.Handler.
func (z *zapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nz := &zapHandler{
		stdout: z.stdout.WithAttrs(attrs),
		otel:   z.otel.WithAttrs(attrs),
	}
	return nz
}

// WithGroup implements slog.Handler.
func (z *zapHandler) WithGroup(name string) slog.Handler {
	nz := &zapHandler{
		stdout: z.stdout.WithGroup(name),
		otel:   z.otel.WithGroup(name),
	}
	return nz
}

var _ slog.Handler = (*zapHandler)(nil)
