package log

import (
	"io"
	"os"
)

// Level is the minimum severity a logger will emit.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Config controls construction of a Logger. The zero value is usable: it logs
// at INFO to stdout and reads the environment from APP_ENV.
type Config struct {
	// Service is the service name stamped on every event ("service" field).
	Service string
	// Environment is stamped on every event ("environment" field). When empty,
	// it falls back to the APP_ENV variable.
	Environment string
	// Level is the minimum severity to emit. Defaults to InfoLevel.
	Level Level
	// Output is where logs are written. Defaults to os.Stdout.
	Output io.Writer
	// Keys overrides the output name of any contract field. Unset keys keep
	// their defaults.
	Keys Keys
	// Driver selects the backend. When nil, the built-in LogrusDriver is used.
	// Provide another factory (e.g. from a zap subpackage) to switch backends
	// without touching call sites.
	Driver DriverFactory
}

// New builds a Logger that enforces the ELK logging contract. The concrete
// backend is chosen by Config.Driver (default: LogrusDriver). All contract
// logic lives in the backend-agnostic core; the driver only formats and writes.
func New(cfg Config) Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	if cfg.Environment == "" {
		cfg.Environment = os.Getenv("APP_ENV")
	}

	factory := cfg.Driver
	if factory == nil {
		factory = JSONDriver
	}

	driver := factory(DriverOptions{
		Service:     cfg.Service,
		Environment: cfg.Environment,
		Level:       cfg.Level,
		Output:      cfg.Output,
		Rename:      cfg.Keys.rename(),
	})

	return &logger{driver: driver}
}
