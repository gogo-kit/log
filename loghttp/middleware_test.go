package loghttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trtuandat98/gkit/log"
)

func newTestLogger(buf *bytes.Buffer) log.Logger {
	return log.New(log.Config{Service: "order-service", Environment: "test", Output: buf})
}

func serve(t *testing.T, l log.Logger, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	handler := Middleware(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestMiddlewareLogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)

	req := httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader("body"))
	req.RemoteAddr = "10.0.0.1:5555"
	req.Header.Set("User-Agent", "test-agent")

	rec := serve(t, l, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "INFO", m["level"])
	assert.Equal(t, "POST", m["method"])
	assert.Equal(t, "/api/orders", m["path"])
	assert.EqualValues(t, 201, m["status_code"])
	assert.Equal(t, "10.0.0.1", m["client_ip"])
	assert.Equal(t, "test-agent", m["user_agent"])
	assert.Contains(t, m, "duration_ms")
	assert.NotEmpty(t, rec.Header().Get(log.RequestIDHeader))
}

func TestMiddlewareGeneratesRequestID(t *testing.T) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := serve(t, newTestLogger(&buf), req)

	id := rec.Header().Get(log.RequestIDHeader)
	assert.NotEmpty(t, id)

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, id, m["request_id"])
}

func TestMiddlewarePropagatesRequestID(t *testing.T) {
	var buf bytes.Buffer
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(log.RequestIDHeader, "req-incoming")
	rec := serve(t, newTestLogger(&buf), req)

	assert.Equal(t, "req-incoming", rec.Header().Get(log.RequestIDHeader))
	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "req-incoming", m["request_id"])
}

func TestClientIPPrefersForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5, 10.0.0.1")
	assert.Equal(t, "203.0.113.5", ClientIP(req))
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:1234"
	assert.Equal(t, "10.0.0.9", ClientIP(req))
}
