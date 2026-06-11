package log

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextRequestAndUserID(t *testing.T) {
	ctx := WithUserID(WithRequestID(context.Background(), "req-1"), "u-1")
	assert.Equal(t, "req-1", RequestIDFromContext(ctx))
	assert.Equal(t, "u-1", UserIDFromContext(ctx))
}

func TestContextEmptyDefaults(t *testing.T) {
	assert.Empty(t, RequestIDFromContext(context.Background()))
	assert.Empty(t, UserIDFromContext(context.Background()))
}

func TestFromContextEnrichesLogger(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)
	ctx := WithUserID(WithRequestID(context.Background(), "req-9"), "u-9")

	FromContext(ctx, l).Info("hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "req-9", m["request_id"])
	assert.Equal(t, "u-9", m["user_id"])
}

func TestFromContextWithoutFieldsReturnsSame(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)
	assert.Same(t, l, FromContext(context.Background(), l))

	var nilCtx context.Context
	var nilLogger Logger
	assert.Nil(t, FromContext(nilCtx, nilLogger))
}

func TestContextMethodsExtractCorrelation(t *testing.T) {
	ctx := WithUserID(WithRequestID(context.Background(), "req-7"), "u-7")

	tests := []struct {
		name  string
		emit  func(l Logger, buf *bytes.Buffer)
		level string
	}{
		{"Info", func(l Logger, _ *bytes.Buffer) { l.InfoContext(ctx, "m") }, "INFO"},
		{"Warn", func(l Logger, _ *bytes.Buffer) { l.WarnContext(ctx, "m") }, "WARN"},
		{"Error", func(l Logger, _ *bytes.Buffer) { l.ErrorContext(ctx, errors.New("x"), "m") }, "ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := newTestLogger(&buf)
			tt.emit(l, &buf)

			var m map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
			assert.Equal(t, tt.level, m["level"])
			assert.Equal(t, "req-7", m["request_id"])
			assert.Equal(t, "u-7", m["user_id"])
		})
	}
}

func TestContextMethodExplicitFieldWins(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf)
	ctx := WithRequestID(context.Background(), "from-ctx")

	// An explicit RequestID must override the one pulled from context.
	l.InfoContext(ctx, "m", RequestID("explicit"))

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "explicit", m["request_id"])
}

func TestPackageLevelContextHelper(t *testing.T) {
	var buf bytes.Buffer
	SetDefault(newTestLogger(&buf))

	InfoContext(WithRequestID(context.Background(), "req-pkg"), "hello")

	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "req-pkg", m["request_id"])
}
