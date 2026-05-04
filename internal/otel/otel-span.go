package otel

import (
	"context"
	"mcop/internal/log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SpanHandleFunc func(ctx context.Context, span trace.Span, log *log.Logger) error

func Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return otel.Tracer(name, opts...)
}

// LogSpanFromContext extracts the span context and logger from a context
func LogSpanFromContext(ctx context.Context) (trace.Span, *log.Logger) {
	return trace.SpanFromContext(ctx), log.WithCtx(ctx)
}

// NewLogSpan creates a new span with the given tracer, context and name, returning the span context, span and logger
func NewLogSpan(ctx context.Context, tracer trace.Tracer, name string) (context.Context, trace.Span, *log.Logger) {
	spanCtx, span := tracer.Start(ctx, name)
	return spanCtx, span, log.WithCtx(spanCtx)
}

// NewLogSpanWithKind creates a new span with the given tracer, context, name and span kind, returning the span context, span and logger
func NewLogSpanWithKind(ctx context.Context, tracer trace.Tracer, name string, kind trace.SpanKind) (context.Context, trace.Span, *log.Logger) {
	spanCtx, span := tracer.Start(ctx, name, trace.WithSpanKind(kind))
	return spanCtx, span, log.WithCtx(spanCtx)
}

func RunInSpan(ctx context.Context, tracer trace.Tracer, name string, f SpanHandleFunc) error {
	spanCtx, span, log := NewLogSpanWithKind(ctx, tracer, name, trace.SpanKindInternal)
	err := f(spanCtx, span, log)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End(trace.WithTimestamp(time.Now()))
	return err
}
