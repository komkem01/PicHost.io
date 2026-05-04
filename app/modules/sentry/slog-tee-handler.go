// Package sentry provides Sentry integration with OpenTelemetry for error tracking and performance monitoring.
package sentry

import (
	"context"
	"errors"
	"log/slog"
)

// TeeHandler implements slog.Handler and directs logs to multiple handlers.
// It allows sending the same log record to multiple handlers simultaneously,
// useful for sending logs to both console and Sentry.
type TeeHandler struct {
	handlers []slog.Handler
}

// NewTeeHandler creates a new TeeHandler with the given handlers.
// All provided handlers will receive the same log records.
func NewTeeHandler(handlers ...slog.Handler) *TeeHandler {
	return &TeeHandler{handlers: handlers}
}

// Enabled reports whether the handler handles records at the given level.
// Returns true if any of the underlying handlers can handle the level.
func (t *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle handles the Record by sending it to all underlying handlers.
// If any handler returns an error, all errors are joined and returned.
func (t *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs error
	for _, h := range t.handlers {
		if err := h.Handle(ctx, r); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

// WithAttrs returns a new handler whose attributes consist of
// the receiver's attributes followed by the given attributes.
// The new TeeHandler will have all underlying handlers updated with the same attributes.
func (t *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(t.handlers))
	for i, h := range t.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return NewTeeHandler(newHandlers...)
}

// WithGroup returns a new handler with the given group name.
// The new TeeHandler will have all underlying handlers updated with the same group.
func (t *TeeHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(t.handlers))
	for i, h := range t.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return NewTeeHandler(newHandlers...)
}
