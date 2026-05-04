package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"
)

type Logger struct {
	Logger *slog.Logger
	ctx    context.Context
}

var defaultLogger atomic.Pointer[Logger]

func init() {
	defaultLogger.Store(&Logger{slog.Default(), context.Background()})
}

func Default() *Logger {
	return &Logger{slog.Default(), context.Background()}
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.Logger.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.Background()
	}
	_ = l.Logger.Handler().Handle(ctx, r)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...), l.ctx}
}

func (l *Logger) WithCtx(ctx context.Context) *Logger {
	return &Logger{l.Logger, ctx}
}

func (l *Logger) Debugf(format string, args ...any) {
	l.log(l.ctx, slog.LevelDebug, fmt.Sprintf(format, args...))
}

func (l *Logger) Infof(format string, args ...any) {
	l.log(l.ctx, slog.LevelInfo, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(format string, args ...any) {
	l.log(l.ctx, slog.LevelWarn, fmt.Sprintf(format, args...))
}

func (l *Logger) Errf(format string, args ...any) {
	l.log(l.ctx, slog.LevelError, fmt.Sprintf(format, args...))
}

func With(args ...any) *Logger {
	return Default().With(args...)
}

func WithCtx(ctx context.Context) *Logger {
	return Default().WithCtx(ctx)
}
