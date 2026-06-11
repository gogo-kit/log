package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureDriver is a test Driver that records events in memory, proving a
// caller can plug a custom backend under the Logger interface.
type captureDriver struct {
	opts     DriverOptions
	events   []captured
	minLevel Level
	synced   bool
}

type captured struct {
	level  Level
	msg    string
	fields map[string]any
}

func newCaptureFactory(d *captureDriver) DriverFactory {
	return func(opts DriverOptions) Driver {
		d.opts = opts
		return d
	}
}

func (d *captureDriver) Write(level Level, msg string, fields map[string]any) {
	d.events = append(d.events, captured{level, msg, fields})
}
func (d *captureDriver) Enabled(level Level) bool { return level >= d.minLevel }
func (d *captureDriver) Sync() error              { d.synced = true; return nil }

func TestCustomDriverReceivesEnrichedEvent(t *testing.T) {
	d := &captureDriver{}
	l := New(Config{
		Service:     "svc",
		Environment: "test",
		Output:      nil, // unused by the custom driver
		Keys:        Keys{ErrorMessage: "error"},
		Driver:      newCaptureFactory(d),
	})

	l.With(RequestID("req-1")).Error(Wrap(errors.New("boom")), "failed", Event("X"))

	require.Len(t, d.events, 1)
	e := d.events[0]
	assert.Equal(t, ErrorLevel, e.level)
	assert.Equal(t, "failed", e.msg)

	// Core enrichment is backend-agnostic: the driver receives canonical keys.
	assert.Equal(t, "req-1", e.fields[keyRequestID])
	assert.Equal(t, "X", e.fields[keyEvent])
	assert.Equal(t, "boom", e.fields[keyErrorMessage])
	assert.NotEmpty(t, e.fields[keyStack])
	assert.NotEmpty(t, e.fields[keyStackTrace])

	// Key renames are passed through to the driver via options, not pre-applied.
	assert.Equal(t, "error", d.opts.Key(keyErrorMessage))
	assert.Equal(t, "test", d.opts.Environment)
}

func TestDisabledLevelSkipsWriteAndStackCapture(t *testing.T) {
	d := &captureDriver{minLevel: ErrorLevel} // Info/Warn/Debug disabled
	l := New(Config{Driver: newCaptureFactory(d)})

	l.Info("dropped")
	l.Warn("dropped")
	l.Error(errors.New("kept"), "kept")

	require.Len(t, d.events, 1)
	assert.Equal(t, ErrorLevel, d.events[0].level)
}

func TestPackageSyncDelegatesToDriver(t *testing.T) {
	d := &captureDriver{}
	SetDefault(New(Config{Driver: newCaptureFactory(d)}))

	require.NoError(t, Sync())
	assert.True(t, d.synced)
}
