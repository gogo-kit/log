package log

import "context"

type contextKey int

const (
	requestIDKey contextKey = iota
	userIDKey
)

// WithRequestID returns a context carrying the request correlation id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext extracts the request id, or "" if absent.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// WithUserID returns a context carrying the user id.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext extracts the user id, or "" if absent.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

// contextFields extracts the correlation fields (request_id, user_id) carried
// by ctx. Returns nil when none are present.
func contextFields(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	var fields []Field
	if id := RequestIDFromContext(ctx); id != "" {
		fields = append(fields, RequestID(id))
	}
	if id := UserIDFromContext(ctx); id != "" {
		fields = append(fields, UserID(id))
	}
	return fields
}

// mergeContext prepends the context's correlation fields to fields so that
// explicit fields, applied later, take precedence.
func mergeContext(ctx context.Context, fields []Field) []Field {
	cf := contextFields(ctx)
	if len(cf) == 0 {
		return fields
	}
	return append(cf, fields...)
}

// FromContext returns a child of l bound to any correlation fields present in
// ctx (request_id, user_id). Useful to capture the context once and reuse the
// bound logger; for one-off calls prefer the *Context methods.
func FromContext(ctx context.Context, l Logger) Logger {
	if l == nil {
		return l
	}
	fields := contextFields(ctx)
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}
