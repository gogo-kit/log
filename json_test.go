package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer) Logger {
	return New(Config{Service: "order-service", Environment: "test", Output: buf})
}

func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	out := buf.String()
	require.Equal(t, 1, strings.Count(out, "\n"), "log event must be a single line")
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &m))
	return m
}

func TestInfoEmitsMandatoryFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)

	l.Info("order created", Event("ORDER_CREATED"), String("order_id", "ord-001"))

	m := decodeLine(t, &buf)
	assert.Equal(t, "INFO", m["level"])
	assert.Equal(t, "order-service", m["service"])
	assert.Equal(t, "test", m["environment"])
	assert.Equal(t, "order created", m["message"])
	assert.Equal(t, "ORDER_CREATED", m["event"])
	assert.Equal(t, "ord-001", m["order_id"])
	assert.NotEmpty(t, m["timestamp"])
}

func TestLevelNames(t *testing.T) {
	tests := []struct {
		emit func(l Logger)
		want string
	}{
		{func(l Logger) { l.Info("m") }, "INFO"},
		{func(l Logger) { l.Warn("m") }, "WARN"},
		{func(l Logger) { l.Error(errors.New("x"), "m") }, "ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			var buf bytes.Buffer
			tt.emit(newTestLogger(&buf))
			assert.Equal(t, tt.want, decodeLine(t, &buf)["level"])
		})
	}
}

func TestWithChainingMergesFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf).With(RequestID("req-123")).With(UserID("u-1"))

	l.Info("hello")

	m := decodeLine(t, &buf)
	assert.Equal(t, "req-123", m["request_id"])
	assert.Equal(t, "u-1", m["user_id"])
}

func TestErrorLogContract(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)

	err := Wrap(&customErr{msg: "gateway timeout"})
	l.Error(err, "failed to create order", RequestID("req-123"))

	m := decodeLine(t, &buf)
	assert.Equal(t, "ERROR", m["level"])
	assert.Equal(t, "failed to create order", m["message"])
	assert.Equal(t, "req-123", m["request_id"])
	assert.Equal(t, "customErr", m["error_type"])
	assert.Equal(t, "gateway timeout", m["error_message"])
	assert.NotEmpty(t, m["stack_trace"])
	assert.NotEmpty(t, m["stack"])
}

func TestCustomKeys(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Service:     "order-service",
		Environment: "test",
		Output:      &buf,
		Keys: Keys{
			ErrorMessage: "error",
			Message:      "msg",
			RequestID:    "trace_id",
		},
	})

	l.Error(Wrap(&customErr{msg: "boom"}), "failed", RequestID("req-1"))

	m := decodeLine(t, &buf)
	// Renamed keys present...
	assert.Equal(t, "boom", m["error"])
	assert.Equal(t, "failed", m["msg"])
	assert.Equal(t, "req-1", m["trace_id"])
	// ...and the defaults are gone.
	assert.NotContains(t, m, "error_message")
	assert.NotContains(t, m, "message")
	assert.NotContains(t, m, "request_id")
	// Untouched keys keep defaults.
	assert.Equal(t, "ERROR", m["level"])
	assert.Equal(t, "boom", m["error"])
	assert.Contains(t, m, "stack_trace")
}

func TestDebugSuppressedAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{Service: "s", Output: &buf, Level: InfoLevel})
	l.Debug("should not appear")
	assert.Empty(t, buf.String())
}

type customErr struct{ msg string }

func (e *customErr) Error() string { return e.msg }
