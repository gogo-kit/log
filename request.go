package log

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// RequestIDHeader is the HTTP header used to read and propagate the request id.
// It lives in the core so every transport adapter (loghttp, loggin, ...) agrees
// on it without importing one another.
const RequestIDHeader = "X-Request-ID"

// RequestInfo holds the values logged for one request. Transport adapters
// (net/http, Gin, echo, fasthttp) populate it and call Request so the emitted
// field set stays identical across frameworks. It carries no transport types,
// so the core stays free of net/http and web frameworks.
type RequestInfo struct {
	RequestID    string
	Method       string
	Path         string
	Status       int
	Duration     time.Duration
	ClientIP     string
	UserAgent    string
	RequestSize  int64
	ResponseSize int64
}

// Request emits a single request-lifecycle log event from the given info.
func Request(l Logger, info RequestInfo) {
	if l == nil {
		l = Default()
	}
	fields := []Field{
		RequestID(info.RequestID),
		String(keyMethod, info.Method),
		String(keyPath, info.Path),
		Int(keyStatusCode, info.Status),
		Int64(keyDurationMs, info.Duration.Milliseconds()),
		String(keyClientIP, info.ClientIP),
		String(keyUserAgent, info.UserAgent),
		Int64(keyResponseSize, info.ResponseSize),
	}
	if info.RequestSize >= 0 {
		fields = append(fields, Int64(keyRequestSize, info.RequestSize))
	}
	l.Info("request completed", fields...)
}

// NewRequestID returns a random 16-byte hex request id (no external uuid dep).
func NewRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-" + time.Now().UTC().Format("20060102150405.000000")
	}
	return hex.EncodeToString(b)
}
