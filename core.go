package log

import "context"

// logger is the backend-agnostic core implementation of Logger. It owns every
// concern that must behave identically regardless of backend — field merging,
// error enrichment, stack capture and context correlation — then delegates the
// raw write to a Driver. Swapping backends means swapping the driver, not this
// type.
type logger struct {
	driver Driver
	fields []Field // bound via With, replayed on every event
}

func (l *logger) Debug(msg string, fields ...Field) { l.log(DebugLevel, nil, msg, fields) }
func (l *logger) Info(msg string, fields ...Field)  { l.log(InfoLevel, nil, msg, fields) }
func (l *logger) Warn(msg string, fields ...Field)  { l.log(WarnLevel, nil, msg, fields) }

// Error enriches the event with error_type, error_message and an application
// stack trace before emitting. The stack is taken from the error when it was
// wrapped via Wrap (preserving the origin); otherwise it is captured here.
func (l *logger) Error(err error, msg string, fields ...Field) {
	l.log(ErrorLevel, err, msg, fields)
}

// log is the single emit path. Keeping it one frame deep keeps captureStack's
// skip count stable for the error fallback.
func (l *logger) log(level Level, err error, msg string, fields []Field) {
	if !l.driver.Enabled(level) {
		return
	}

	data := make(map[string]any, len(l.fields)+len(fields)+4)
	for _, f := range l.fields {
		data[f.Key] = f.Value
	}
	for _, f := range fields {
		data[f.Key] = f.Value
	}

	if err != nil {
		data[keyErrorType] = errorType(err)
		data[keyErrorMessage] = err.Error()

		stack := stackFromError(err)
		if stack == nil {
			stack = captureStack(2) // caller -> Error/log wrapper -> log -> captureStack
		}
		data[keyStack] = stack
		data[keyStackTrace] = formatStack(stack)
	}

	l.driver.Write(level, msg, data)
}

// Context-aware variants pull correlation fields (request_id, user_id) out of
// ctx and prepend them, so explicit fields still take precedence.

func (l *logger) DebugContext(ctx context.Context, msg string, fields ...Field) {
	l.log(DebugLevel, nil, msg, mergeContext(ctx, fields))
}

func (l *logger) InfoContext(ctx context.Context, msg string, fields ...Field) {
	l.log(InfoLevel, nil, msg, mergeContext(ctx, fields))
}

func (l *logger) WarnContext(ctx context.Context, msg string, fields ...Field) {
	l.log(WarnLevel, nil, msg, mergeContext(ctx, fields))
}

func (l *logger) ErrorContext(ctx context.Context, err error, msg string, fields ...Field) {
	l.log(ErrorLevel, err, msg, mergeContext(ctx, fields))
}

func (l *logger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return l
	}
	merged := make([]Field, 0, len(l.fields)+len(fields))
	merged = append(merged, l.fields...)
	merged = append(merged, fields...)
	return &logger{driver: l.driver, fields: merged}
}

// Sync flushes the underlying driver. Exposed for graceful shutdown (e.g. zap
// buffers); not part of the Logger interface. Use the package-level Sync to
// flush the default logger.
func (l *logger) Sync() error { return l.driver.Sync() }
