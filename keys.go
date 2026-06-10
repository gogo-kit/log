package log

// Standard field keys that make up the ELK logging contract. Centralizing them
// keeps the schema in one place: rename or add a key here and every producer
// (field constructors, formatter, middleware) follows.
const (
	// Mandatory on every event.
	keyTimestamp   = "timestamp"
	keyLevel       = "level"
	keyService     = "service"
	keyEnvironment = "environment"
	keyMessage     = "message"

	// Correlation / business context.
	keyRequestID  = "request_id"
	keyUserID     = "user_id"
	keyEvent      = "event"
	keyResourceID = "resource_id"

	// Error events.
	keyErrorType    = "error_type"
	keyErrorMessage = "error_message"
	keyStack        = "stack"
	keyStackTrace   = "stack_trace"

	// Request / interceptor events.
	keyMethod       = "method"
	keyPath         = "path"
	keyStatusCode   = "status_code"
	keyDurationMs   = "duration_ms"
	keyClientIP     = "client_ip"
	keyUserAgent    = "user_agent"
	keyRequestSize  = "request_size"
	keyResponseSize = "response_size"
)

// Keys overrides the output name of any contract field. A zero-value (empty)
// field keeps the default, so callers only set the keys they want to rename:
//
//	log.New(log.Config{Keys: log.Keys{ErrorMessage: "error"}})
//
// emits "error" instead of "error_message" while leaving every other key
// untouched.
type Keys struct {
	Timestamp   string
	Level       string
	Service     string
	Environment string
	Message     string

	RequestID  string
	UserID     string
	Event      string
	ResourceID string

	ErrorType    string
	ErrorMessage string
	Stack        string
	StackTrace   string

	Method       string
	Path         string
	StatusCode   string
	DurationMs   string
	ClientIP     string
	UserAgent    string
	RequestSize  string
	ResponseSize string
}

// rename builds a canonical -> effective lookup containing only the keys the
// caller actually overrode. Returns nil when there are no overrides so the
// common path stays allocation-free.
func (k Keys) rename() map[string]string {
	m := make(map[string]string)
	add := func(canonical, custom string) {
		if custom != "" && custom != canonical {
			m[canonical] = custom
		}
	}
	add(keyTimestamp, k.Timestamp)
	add(keyLevel, k.Level)
	add(keyService, k.Service)
	add(keyEnvironment, k.Environment)
	add(keyMessage, k.Message)
	add(keyRequestID, k.RequestID)
	add(keyUserID, k.UserID)
	add(keyEvent, k.Event)
	add(keyResourceID, k.ResourceID)
	add(keyErrorType, k.ErrorType)
	add(keyErrorMessage, k.ErrorMessage)
	add(keyStack, k.Stack)
	add(keyStackTrace, k.StackTrace)
	add(keyMethod, k.Method)
	add(keyPath, k.Path)
	add(keyStatusCode, k.StatusCode)
	add(keyDurationMs, k.DurationMs)
	add(keyClientIP, k.ClientIP)
	add(keyUserAgent, k.UserAgent)
	add(keyRequestSize, k.RequestSize)
	add(keyResponseSize, k.ResponseSize)
	if len(m) == 0 {
		return nil
	}
	return m
}
