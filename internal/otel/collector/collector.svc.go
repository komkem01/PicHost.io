package collector

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"mcop/internal/config"
	"mcop/internal/provider"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ provider.Close = (*Service)(nil)

// Service represents a service for the OpenTelemetry Collector.
type Service struct {
	close func(context.Context) error
}

type Config struct {
	CollectorEndpoint string
	TraceMode         string
	MetricMode        string
	LogMode           string
	LogLevel          string
	TraceRatio        float64
}

func newService(conf *config.Config[Config]) *Service {
	return &Service{
		close: initProvider(conf),
	}
}

func initProvider(conf *config.Config[Config]) func(context.Context) error {
	handlePanic := func(err error, msg string) {
		if err != nil {
			panic(fmt.Sprintf("%s: %v", msg, err))
		}
	}

	createResource := func(conf *config.Config[Config]) (*resource.Resource, error) {
		hostname := conf.Hostname()
		if hostname == "" {
			hostname = hex.EncodeToString(big.NewInt(rand.Int63()).Bytes())
		}
		res, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				// the service name used to display traces in backends
				semconv.ServiceNameKey.String(conf.AppName()),
				semconv.HostNameKey.String(hostname),
			),
		)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	createTraceExporter := func(ctx context.Context, conn *grpc.ClientConn) (sdkTrace.SpanExporter, error) {
		return otlptracegrpc.New(ctx,
			otlptracegrpc.WithGRPCConn(conn),
			otlptracegrpc.WithCompressor("gzip"),
		)
	}

	createMetricExporter := func(ctx context.Context, conn *grpc.ClientConn) (sdkMetric.Exporter, error) {
		return otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithGRPCConn(conn),
			otlpmetricgrpc.WithCompressor("gzip"),
		)
	}

	createLogExporter := func(ctx context.Context, conn *grpc.ClientConn) (sdkLog.Exporter, error) {
		return otlploggrpc.New(ctx,
			otlploggrpc.WithGRPCConn(conn),
			otlploggrpc.WithCompressor("gzip"),
		)
	}

	ctx := context.Background()

	res, err := createResource(conf)
	handlePanic(err, "failed to create resource")

	var (
		traceExporter  sdkTrace.SpanExporter
		metricExporter sdkMetric.Exporter
		logExporter    sdkLog.Exporter
		logProcess     sdkLog.Processor
		tracerProvider trace.TracerProvider
		meterProvider  metric.MeterProvider
		loggerProvider log.LoggerProvider
		shutdowns      []func(context.Context) error
	)

	if conf.Val.CollectorEndpoint != "" {
		conn, err := grpc.NewClient(conf.Val.CollectorEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(25*1024*1024), // 25MB
			),
		)
		handlePanic(err, fmt.Sprintf("failed to create gRPC connection to collector: %s", err))

		traceExporter, err = createTraceExporter(ctx, conn)
		handlePanic(err, "failed to create trace exporter")
		metricExporter, err = createMetricExporter(ctx, conn)
		handlePanic(err, "failed to create metric exporter")
		logExporter, err = createLogExporter(ctx, conn)
		handlePanic(err, "failed to create log exporter")
		logProcess = sdkLog.NewBatchProcessor(logExporter)
	}

	if conf.Val.TraceMode == "stdout" {
		traceExporter, err = stdouttrace.New()
		handlePanic(err, "failed to create trace exporter")
	}

	if conf.Val.MetricMode == "stdout" {
		metricExporter, err = stdoutmetric.New()
		handlePanic(err, "failed to create metric exporter")
	}

	if conf.Val.LogMode == "stdout" {
		logExporter, err = stdoutlog.New()
		handlePanic(err, "failed to create log exporter")
		logProcess = sdkLog.NewSimpleProcessor(logExporter)
	}

	createTracerProvider := func(traceExporter sdkTrace.SpanExporter) trace.TracerProvider {
		traceOpts := []sdkTrace.TracerProviderOption{
			sdkTrace.WithResource(res),
		}

		if traceExporter != nil {
			traceOpts = append(traceOpts,
				sdkTrace.WithBatcher(traceExporter),
				sdkTrace.WithSampler(ErrorOrParentOrRatioSampler(conf.Val.TraceRatio)),
			)
		} else {
			// TraceMode is "noop"
			traceExporter = &noopSpanExporter{}
			traceOpts = append(traceOpts,
				sdkTrace.WithSyncer(traceExporter),
			)
		}
		tp := sdkTrace.NewTracerProvider(
			traceOpts...,
		)
		shutdowns = append(shutdowns, tp.Shutdown)
		return tp
	}
	createMetricProvider := func(metricExporter sdkMetric.Exporter) metric.MeterProvider {
		if metricExporter != nil {
			mp := sdkMetric.NewMeterProvider(
				sdkMetric.WithResource(res),
				sdkMetric.WithReader(sdkMetric.NewPeriodicReader(metricExporter)),
			)
			shutdowns = append(shutdowns, mp.Shutdown)
			return mp
		}
		return metricNoop.NewMeterProvider()
	}
	createLoggerProvider := func(logExporter sdkLog.Exporter) log.LoggerProvider {
		if logExporter != nil {
			lp := sdkLog.NewLoggerProvider(
				sdkLog.WithResource(res),
				sdkLog.WithProcessor(logProcess),
			)
			shutdowns = append(shutdowns, lp.Shutdown)
			return lp
		}
		return noop.NewLoggerProvider()
	}

	tracerProvider = createTracerProvider(traceExporter)
	otel.SetTracerProvider(tracerProvider)

	meterProvider = createMetricProvider(metricExporter)
	otel.SetMeterProvider(meterProvider)

	loggerProvider = createLoggerProvider(logExporter)
	global.SetLoggerProvider(loggerProvider)

	// Note: TextMapPropagator will be set by Sentry service to include Sentry propagation
	// If Sentry is not enabled, set default propagators
	if otel.GetTextMapPropagator() == nil {
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	}
	otel.SetErrorHandler(
		func(debug bool) otel.ErrorHandlerFunc {
			if debug {
				return func(err error) {
					fmt.Println("otel:", err)
				}
			}
			return func(_ error) {
				// do nothing
			}
		}(conf.Debug()))
	return func(ctx context.Context) error {
		var err error
		for _, fn := range shutdowns {
			if e := fn(ctx); e != nil {
				err = errors.Join(err, e)
			}
		}
		return err
	}
}

var _ sdkTrace.Sampler = (*errorOrParentOrRatioSampler)(nil)

type errorOrParentOrRatioSampler struct {
	sdkTrace.Sampler
}

// ShouldSample implements trace.Sampler.
func (e *errorOrParentOrRatioSampler) ShouldSample(parameters sdkTrace.SamplingParameters) sdkTrace.SamplingResult {
	if parameters.Attributes != nil {
		for _, attr := range parameters.Attributes {
			if attr.Key == semconv.ErrorTypeKey || attr.Key == semconv.ExceptionTypeKey {
				return sdkTrace.SamplingResult{
					Decision:   sdkTrace.RecordAndSample,
					Attributes: parameters.Attributes,
				}
			}
		}
	}
	return e.Sampler.ShouldSample(parameters)
}

func (e *errorOrParentOrRatioSampler) Description() string {
	return "ErrorOrParentOrRatioSampler"
}

// ErrorOrParentOrRatioSampler creates a sampler that samples traces based on error status,
// parent trace sampling decision, or a given ratio when neither condition applies.
func ErrorOrParentOrRatioSampler(ratio float64) sdkTrace.Sampler {
	return &errorOrParentOrRatioSampler{
		sdkTrace.ParentBased(sdkTrace.TraceIDRatioBased(ratio)),
	}
}

// Close implements the provider.Close interface and shuts down the telemetry providers.
func (m *Service) Close(ctx context.Context) error {
	if m.close == nil {
		return nil
	}
	return m.close(ctx)
}
