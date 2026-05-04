// Package utils provides utility functions for OpenTelemetry tracing and logging
package utils

import (
	"context"

	"mcop/internal/log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// LogSpanFromGin extracts the span context and logger from a Gin context
func LogSpanFromGin(ginCtx *gin.Context) (trace.Span, *log.Logger) {
	ctx := ginCtx.Request.Context()
	return LogSpanFromContext(ctx)
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
