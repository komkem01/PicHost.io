// Package log provides logging functionality with structured logging support using slog
package log

import (
	"log/slog"
	"time"
)

var (
	// Any creates an Any slog.Attr with the given key and value.
	Any = slog.Any
	// String creates a String slog.Attr with the given key and value.
	String = slog.String
	// Int creates an Int slog.Attr with the given key and value.
	Int = slog.Int
	// Int64 creates an Int64 slog.Attr with the given key and value.
	Int64 = slog.Int64
	// Uint64 creates a Uint64 slog.Attr with the given key and value.
	Uint64 = slog.Uint64
	// Float64 creates a Float64 slog.Attr with the given key and value.
	Float64 = slog.Float64
	// Duration creates a Duration slog.Attr with the given key and value.
	Duration = slog.Duration
	// Bool creates a Bool slog.Attr with the given key and value.
	Bool = slog.Bool
	// Time creates a Time slog.Attr with the given key and value.
	Time = slog.Time
)

// Error creates a slog.Attr from an error value.
func Error(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return String("error", err.Error())
}

// WithAttrs returns a new Logger with the given attributes added to its context.
func (l *Logger) WithAttrs(attrs ...slog.Attr) *Logger {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return l.With(args...)
}

// Error creates a slog.Attr from an error value.
func (l *Logger) Error(err error) slog.Attr {
	return Error(err)
}

// String creates a String slog.Attr with the given key and value.
func (l *Logger) String(key, value string) slog.Attr {
	return String(key, value)
}

// Int creates an Int slog.Attr with the given key and value.
func (l *Logger) Int(key string, value int) slog.Attr {
	return Int(key, value)
}

// Int64 creates an Int64 slog.Attr with the given key and value.
func (l *Logger) Int64(key string, value int64) slog.Attr {
	return Int64(key, value)
}

// Uint64 creates a Uint64 slog.Attr with the given key and value.
func (l *Logger) Uint64(key string, value uint64) slog.Attr {
	return Uint64(key, value)
}

// Float64 creates a Float64 slog.Attr with the given key and value.
func (l *Logger) Float64(key string, value float64) slog.Attr {
	return Float64(key, value)
}

// Duration creates a Duration slog.Attr with the given key and value.
func (l *Logger) Duration(key string, value time.Duration) slog.Attr {
	return Duration(key, value)
}

// Bool creates a Bool slog.Attr with the given key and value.
func (l *Logger) Bool(key string, value bool) slog.Attr {
	return Bool(key, value)
}

// Time creates a Time slog.Attr with the given key and value.
func (l *Logger) Time(key string, value time.Time) slog.Attr {
	return Time(key, value)
}
