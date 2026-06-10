package log

import (
	"encoding/json"
	"sync"
	"time"
)

// JSONDriver is the built-in, zero-dependency DriverFactory and the default when
// Config.Driver is nil. It writes one line of contract JSON per event straight
// to the configured output using only the standard library, so importing the
// core adds no third-party logging dependency. Backends with richer ecosystems
// (logrus via loglogrus, zap via logzap) are opt-in modules selected through
// Config.Driver.
func JSONDriver(opts DriverOptions) Driver {
	return &jsonDriver{opts: opts}
}

type jsonDriver struct {
	opts DriverOptions
	mu   sync.Mutex // serializes writes to the shared output
}

func (d *jsonDriver) Enabled(level Level) bool { return level >= d.opts.Level }

// Sync is a no-op: writes go straight to the output.
func (d *jsonDriver) Sync() error { return nil }

func (d *jsonDriver) Write(level Level, msg string, fields map[string]any) {
	out := make(map[string]any, len(fields)+5)

	// Data carries canonical keys; resolve each to its configured output name.
	for k, v := range fields {
		out[d.opts.Key(k)] = v
	}

	out[d.opts.Key(KeyTimestamp)] = time.Now().UTC().Format(time.RFC3339)
	out[d.opts.Key(KeyLevel)] = levelName(level)
	out[d.opts.Key(KeyService)] = d.opts.Service
	out[d.opts.Key(KeyEnvironment)] = d.opts.Environment
	out[d.opts.Key(KeyMessage)] = msg

	b, err := json.Marshal(out)
	if err != nil {
		return
	}
	b = append(b, '\n')

	d.mu.Lock()
	defer d.mu.Unlock()
	_, _ = d.opts.Output.Write(b)
}

// levelName renders a Level using the UPPERCASE contract names.
func levelName(l Level) string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "INFO"
	}
}
