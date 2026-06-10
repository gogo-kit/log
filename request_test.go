package log

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogsContractFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{Service: "svc", Environment: "test", Output: &buf})

	Request(l, RequestInfo{
		RequestID:    "req-1",
		Method:       "POST",
		Path:         "/api/orders",
		Status:       201,
		Duration:     42 * time.Millisecond,
		ClientIP:     "10.0.0.1",
		UserAgent:    "agent",
		RequestSize:  12,
		ResponseSize: 34,
	})

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "INFO", m["level"])
	assert.Equal(t, "request completed", m["message"])
	assert.Equal(t, "req-1", m["request_id"])
	assert.Equal(t, "POST", m["method"])
	assert.Equal(t, "/api/orders", m["path"])
	assert.EqualValues(t, 201, m["status_code"])
	assert.EqualValues(t, 42, m["duration_ms"])
	assert.Equal(t, "10.0.0.1", m["client_ip"])
	assert.EqualValues(t, 12, m["request_size"])
	assert.EqualValues(t, 34, m["response_size"])
}

func TestRequestOmitsNegativeRequestSize(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{Service: "svc", Output: &buf})

	Request(l, RequestInfo{RequestID: "r", RequestSize: -1})

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.NotContains(t, m, "request_size")
}

func TestNewRequestIDIsUniqueHex(t *testing.T) {
	a := NewRequestID()
	b := NewRequestID()
	assert.Len(t, a, 32)
	assert.NotEqual(t, a, b)
}
