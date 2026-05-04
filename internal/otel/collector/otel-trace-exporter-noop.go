package collector

import (
	"context"

	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

type noopSpanExporter struct{}

var _ sdkTrace.SpanExporter = (*noopSpanExporter)(nil)

// ExportSpans implements trace.SpanExporter.
func (n *noopSpanExporter) ExportSpans(ctx context.Context, spans []sdkTrace.ReadOnlySpan) error {
	return nil
}

// Shutdown implements trace.SpanExporter.
func (n *noopSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}
