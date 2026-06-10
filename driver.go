package log

import "io"

// Driver is the low-level logging backend. Implement it to plug a new library
// (logrus, Uber zap, zerolog, ...) under the Logger interface. The high-level
// core Logger owns all backend-agnostic concerns — field typing, error
// enrichment, stack capture, context correlation — and hands the driver a
// fully-prepared event to format and write.
//
// Drivers are constructed through a DriverFactory and selected via Config.Driver.
// The built-in default is LogrusDriver. A driver lives wherever its dependency
// is acceptable: the logrus driver ships in this package (logrus is already a
// dependency), while heavier backends like zap belong in their own subpackage
// so callers only pull the dependency they choose.
type Driver interface {
	// Write emits a single event at level. fields carries every resolved
	// key/value for the event using canonical keys; the driver is responsible
	// for applying key renames (DriverOptions.Rename) and stamping the mandatory
	// contract fields (timestamp, level, service, environment, message).
	Write(level Level, msg string, fields map[string]any)

	// Enabled reports whether an event at level would be emitted, letting the
	// core skip expensive work (e.g. stack capture) for filtered levels.
	Enabled(level Level) bool

	// Sync flushes any buffered events. No-op for unbuffered backends.
	Sync() error
}

// DriverFactory builds a Driver from resolved options. Config.New passes a
// normalized DriverOptions (defaults applied, rename map computed) so factories
// stay simple and consistent across backends.
type DriverFactory func(DriverOptions) Driver

// DriverOptions is the normalized configuration a Driver needs. It is derived
// from Config by New: Output and Environment are defaulted, and Rename is the
// canonical -> effective key map (nil when there are no overrides).
type DriverOptions struct {
	Service     string
	Environment string
	Level       Level
	Output      io.Writer
	Rename      map[string]string
}

// Key resolves a canonical contract key to its configured output name. Drivers
// use it when stamping mandatory fields and renaming event data.
func (o DriverOptions) Key(canonical string) string {
	if effective, ok := o.Rename[canonical]; ok {
		return effective
	}
	return canonical
}
