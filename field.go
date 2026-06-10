package log

import "time"

// Field is a single structured key/value pair attached to a log event.
// Business code builds fields through the typed constructors below so the
// concrete backend never leaks into call sites.
type Field struct {
	Key   string
	Value any
}

// String creates a string field.
func String(key, value string) Field { return Field{Key: key, Value: value} }

// Int creates an int field.
func Int(key string, value int) Field { return Field{Key: key, Value: value} }

// Int64 creates an int64 field.
func Int64(key string, value int64) Field { return Field{Key: key, Value: value} }

// Float64 creates a float64 field.
func Float64(key string, value float64) Field { return Field{Key: key, Value: value} }

// Bool creates a bool field.
func Bool(key string, value bool) Field { return Field{Key: key, Value: value} }

// Duration records a duration as integer milliseconds, matching the ELK contract.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.Milliseconds()}
}

// Time records a timestamp in RFC3339 (UTC).
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.UTC().Format(time.RFC3339)}
}

// Any creates a field for an arbitrary value.
func Any(key string, value any) Field { return Field{Key: key, Value: value} }

// Err attaches an error message under the standard "error_message" key.
func Err(err error) Field {
	if err == nil {
		return Field{Key: keyErrorMessage, Value: nil}
	}
	return Field{Key: keyErrorMessage, Value: err.Error()}
}

// Standard correlation / business fields from the logging contract.

// RequestID sets the "request_id" correlation field.
func RequestID(value string) Field { return Field{Key: keyRequestID, Value: value} }

// UserID sets the "user_id" field.
func UserID(value string) Field { return Field{Key: keyUserID, Value: value} }

// Event sets the "event" field (e.g. ORDER_CREATED).
func Event(value string) Field { return Field{Key: keyEvent, Value: value} }

// ResourceID sets the "resource_id" field.
func ResourceID(value string) Field { return Field{Key: keyResourceID, Value: value} }

// fieldsToMap flattens fields into a map for the backend adapter.
func fieldsToMap(fields ...Field) map[string]any {
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}
