// Package loghttp is a net/http add-on for gkit/log. It logs the request
// lifecycle and propagates a correlation id, building on the transport-agnostic
// log.Request contract. It depends only on the standard library and gkit/log,
// so importing it adds no third-party dependencies.
package loghttp

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/trtuandat98/gkit/log"
)

// Middleware returns net/http middleware that logs each request and enriches
// the request context with a correlation id (read from, or written to, the
// X-Request-ID header).
func Middleware(l log.Logger) func(http.Handler) http.Handler {
	if l == nil {
		l = log.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := r.Header.Get(log.RequestIDHeader)
			if requestID == "" {
				requestID = log.NewRequestID()
			}
			w.Header().Set(log.RequestIDHeader, requestID)
			r = r.WithContext(log.WithRequestID(r.Context(), requestID))

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			log.Request(l, log.RequestInfo{
				RequestID:    requestID,
				Method:       r.Method,
				Path:         r.URL.Path,
				Status:       rec.status,
				Duration:     time.Since(start),
				ClientIP:     ClientIP(r),
				UserAgent:    r.UserAgent(),
				RequestSize:  r.ContentLength,
				ResponseSize: rec.written,
			})
		})
	}
}

// statusRecorder captures the response status code and body size.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.status = code
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

// ClientIP resolves the originating client IP, preferring X-Forwarded-For and
// X-Real-IP over the raw connection address.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if first, _, ok := strings.Cut(xff, ","); ok {
			return strings.TrimSpace(first)
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
