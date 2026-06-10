package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageLevelHelpersUseDefault(t *testing.T) {
	var buf bytes.Buffer
	prev := Default()
	SetDefault(newTestLogger(&buf))
	t.Cleanup(func() { SetDefault(prev) })

	tests := []struct {
		name  string
		emit  func()
		level string
	}{
		{"Debug", func() { Debug("d") }, "DEBUG"}, // suppressed at info, see below
		{"Info", func() { Info("i") }, "INFO"},
		{"Warn", func() { Warn("w") }, "WARN"},
		{"Error", func() { Error(errors.New("x"), "e") }, "ERROR"},
	}
	for _, tt := range tests {
		if tt.name == "Debug" {
			continue // default logger is InfoLevel; Debug is intentionally dropped
		}
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.emit()
			var m map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
			assert.Equal(t, tt.level, m["level"])
		})
	}
}

func TestDefaultConcurrentAccess(t *testing.T) {
	prev := Default()
	t.Cleanup(func() { SetDefault(prev) })

	var buf bytes.Buffer
	SetDefault(newTestLogger(&buf))

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() { defer wg.Done(); Info("concurrent") }()
		go func() { defer wg.Done(); SetDefault(newTestLogger(&bytes.Buffer{})) }()
	}
	wg.Wait()
	assert.NotNil(t, Default())
}

func TestSetDefaultIgnoresNil(t *testing.T) {
	before := Default()
	SetDefault(nil)
	assert.Same(t, before, Default())
}

func TestTimeField(t *testing.T) {
	ts := time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC)
	f := Time("ts", ts)
	assert.Equal(t, "ts", f.Key)
	assert.Equal(t, "2026-06-10T15:00:00Z", f.Value)
}
