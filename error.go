package log

import (
	"errors"
	"reflect"
)

// stackError wraps an error with the application stack captured at its origin.
type stackError struct {
	err   error
	stack []StackFrame
}

func (e *stackError) Error() string { return e.err.Error() }
func (e *stackError) Unwrap() error { return e.err }

// stackTrace returns the frames captured when the error was wrapped.
func (e *stackError) stackTrace() []StackFrame { return e.stack }

// Wrap captures the application stack trace near the error origin and returns an
// error carrying it. Prefer wrapping at the failure site:
//
//	if err != nil {
//	    return log.Wrap(err)
//	}
//
// Wrap is nil-safe and idempotent: if the chain already carries a captured
// stack, the original error is returned unchanged so the origin is preserved.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	var se *stackError
	if errors.As(err, &se) {
		return err
	}
	return &stackError{err: err, stack: captureStack(1)}
}

// errorType returns the concrete type name of the root cause error.
func errorType(err error) string {
	if err == nil {
		return ""
	}
	root := err
	for {
		// Unwrap our own carrier to reach the underlying cause.
		var se *stackError
		if errors.As(root, &se) {
			root = se.err
			continue
		}
		next := errors.Unwrap(root)
		if next == nil {
			break
		}
		root = next
	}
	t := reflect.TypeOf(root)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if name := t.Name(); name != "" {
		return name
	}
	return t.String()
}

// stackFromError returns the captured stack if the error carries one.
func stackFromError(err error) []StackFrame {
	var se *stackError
	if errors.As(err, &se) {
		return se.stackTrace()
	}
	return nil
}
