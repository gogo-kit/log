package log

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LogrusDriver is the built-in DriverFactory, backed by sirupsen/logrus. It is
// used by New when Config.Driver is left nil. logrus is already a dependency of
// this package, so the default driver adds no new imports for callers.
func LogrusDriver(opts DriverOptions) Driver {
	l := logrus.New()
	l.SetOutput(opts.Output)
	l.SetLevel(opts.Level.toLogrus())
	l.SetFormatter(&contractFormatter{opts: opts})
	return &logrusDriver{logger: l}
}

// logrusDriver adapts logrus to the Driver interface. It only formats and
// writes events; all enrichment happens in the core logger.
type logrusDriver struct {
	logger *logrus.Logger
}

func (d *logrusDriver) Write(level Level, msg string, fields map[string]any) {
	d.logger.WithFields(fields).Log(level.toLogrus(), msg)
}

func (d *logrusDriver) Enabled(level Level) bool {
	return d.logger.IsLevelEnabled(level.toLogrus())
}

// Sync is a no-op: logrus writes synchronously.
func (d *logrusDriver) Sync() error { return nil }

// contractFormatter renders each entry as one line of JSON matching the ELK
// contract, replacing logrus' default JSON formatter (which lowercases the
// level and names fields "time"/"msg"). Key renames come from the driver opts.
type contractFormatter struct {
	opts DriverOptions
}

func (f *contractFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	out := make(map[string]any, len(entry.Data)+5)

	// Field data first so mandatory keys below always win. Data carries
	// canonical keys (from the field constructors), so translate as we copy.
	for k, v := range entry.Data {
		out[f.opts.key(k)] = v
	}

	out[f.opts.key(keyTimestamp)] = entry.Time.UTC().Format(time.RFC3339)
	out[f.opts.key(keyLevel)] = levelString(entry.Level)
	out[f.opts.key(keyService)] = f.opts.Service
	out[f.opts.key(keyEnvironment)] = f.opts.Environment
	out[f.opts.key(keyMessage)] = entry.Message

	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func levelString(l logrus.Level) string {
	switch l {
	case logrus.WarnLevel:
		return "WARN"
	case logrus.PanicLevel:
		return "PANIC"
	case logrus.FatalLevel:
		return "FATAL"
	default:
		return strings.ToUpper(l.String())
	}
}
