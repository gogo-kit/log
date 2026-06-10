package log

import (
	"context"
)

// Logger is the high-level, backend-agnostic logging abstraction. Business code
// depends only on this interface, never on logrus/zap/zerolog/slog directly, so
// the implementation can be swapped without touching call sites.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)

	// Error logs a failure. The implementation automatically derives error_type,
	// error_message and an application-only stack trace from err; callers never
	// supply stack information themselves.
	Error(err error, msg string, fields ...Field)

	loggerWithContext

	// With returns a child logger that includes the given fields on every event.
	With(fields ...Field) Logger
}

// Context-aware variants automatically enrich the event with correlation
// fields (request_id, user_id) carried by ctx.
type loggerWithContext interface {
	DebugContext(ctx context.Context, msg string, fields ...Field)
	InfoContext(ctx context.Context, msg string, fields ...Field)
	WarnContext(ctx context.Context, msg string, fields ...Field)
	ErrorContext(ctx context.Context, err error, msg string, fields ...Field)
}

// defaultLogger holds the package-level default. A *Logger is stored (rather than the interface directly)
// so swapping in any concrete implementation never panics.
// Should not use concurrent.
var defaultLogger Logger

func init() {
	defaultLogger = New(Config{})
}

// SetDefault replaces the package-level default logger. It is safe to call
// concurrently with Default and the package-level logging helpers.
func SetDefault(l Logger) {
	if l != nil {
		defaultLogger = l
	}
}

// Default returns the package-level default logger. The read is lock-free.
func Default() Logger {
	return defaultLogger
}

// Package-level convenience wrappers around the default logger.

// Debug logs at DEBUG level using the default logger.
func Debug(msg string, fields ...Field) { Default().Debug(msg, fields...) }

// Info logs at INFO level using the default logger.
func Info(msg string, fields ...Field) { Default().Info(msg, fields...) }

// Warn logs at WARN level using the default logger.
func Warn(msg string, fields ...Field) { Default().Warn(msg, fields...) }

// Error logs at ERROR level using the default logger.
func Error(err error, msg string, fields ...Field) { Default().Error(err, msg, fields...) }

// DebugContext logs at DEBUG level via the default logger, enriching the event
// with correlation fields carried by ctx.
func DebugContext(ctx context.Context, msg string, fields ...Field) {
	Default().DebugContext(ctx, msg, fields...)
}

// InfoContext logs at INFO level via the default logger, enriching the event
// with correlation fields carried by ctx.
func InfoContext(ctx context.Context, msg string, fields ...Field) {
	Default().InfoContext(ctx, msg, fields...)
}

// WarnContext logs at WARN level via the default logger, enriching the event
// with correlation fields carried by ctx.
func WarnContext(ctx context.Context, msg string, fields ...Field) {
	Default().WarnContext(ctx, msg, fields...)
}

// ErrorContext logs at ERROR level via the default logger, enriching the event
// with correlation fields carried by ctx.
func ErrorContext(ctx context.Context, err error, msg string, fields ...Field) {
	Default().ErrorContext(ctx, err, msg, fields...)
}

// Sync flushes the default logger's underlying driver. Call it before exit when
// using a buffered backend (e.g. zap). No-op for unbuffered backends.
func Sync() error {
	if s, ok := Default().(interface{ Sync() error }); ok {
		return s.Sync()
	}
	return nil
}
