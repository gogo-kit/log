package log

import (
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

// StackFrame is a single application stack frame, structured for Elasticsearch
// querying and aggregation.
type StackFrame struct {
	Func string `json:"func"`
	File string `json:"file"`
	Line int    `json:"line"`
}

var (
	moduleOnce sync.Once
	moduleName string
)

// appModule returns the path of the main application module, used to keep only
// application-owned frames in captured stacks. It is memoized.
func appModule() string {
	moduleOnce.Do(func() {
		if info, ok := debug.ReadBuildInfo(); ok &&
			info.Main.Path != "" && info.Main.Path != "command-line-arguments" {
			moduleName = info.Main.Path
			return
		}
		// Fallback (e.g. `go run` outside a module): derive the module prefix
		// from this package's own fully-qualified function name.
		if pc, _, _, ok := runtime.Caller(0); ok {
			fn := runtime.FuncForPC(pc).Name()
			if idx := strings.LastIndex(fn, "/log."); idx >= 0 {
				moduleName = fn[:idx]
			}
		}
	})
	return moduleName
}

// captureStack records the application call stack at the call site. skip is the
// number of frames to drop beyond captureStack itself (1 = the direct caller is
// dropped, so the stack starts at its caller / the error origin). Only frames
// belonging to the application module are kept, which naturally excludes
// runtime.*, reflect.*, net/http.*, grpc.* and third-party internals.
func captureStack(skip int) []StackFrame {
	const maxDepth = 64
	pcs := make([]uintptr, maxDepth)
	// +2 accounts for runtime.Callers and captureStack themselves.
	n := runtime.Callers(skip+2, pcs)
	if n == 0 {
		return nil
	}

	module := appModule()
	frames := runtime.CallersFrames(pcs[:n])
	out := make([]StackFrame, 0, n)
	for {
		frame, more := frames.Next()
		if isAppFrame(frame.Function, module) {
			out = append(out, parseFrame(frame.Function, frame.File, frame.Line, module))
		}
		if !more {
			break
		}
	}
	return out
}

// isAppFrame reports whether a frame belongs to the application. Module-owned
// packages carry the module path; the program's entrypoint package is reported
// by the runtime as "main.*" (no module prefix), so it is kept explicitly.
// Everything else (runtime.*, reflect.*, net/http.*, grpc.*, third-party) is
// excluded.
func isAppFrame(fn, module string) bool {
	if fn == "main" || strings.HasPrefix(fn, "main.") {
		return true
	}
	return module != "" && strings.HasPrefix(fn, module)
}

// parseFrame converts a raw runtime frame into a module-relative StackFrame.
//
//	github.com/acme/svc/internal/payment.(*Service).Charge
//	  -> Func: internal/payment.(*Service).Charge
//	  -> File: internal/payment/service.go
func parseFrame(fullFunc, file string, line int, module string) StackFrame {
	strippedFunc := strings.TrimPrefix(fullFunc, module+"/")

	pkgPath := fullFunc
	if lastSlash := strings.LastIndex(fullFunc, "/"); lastSlash >= 0 {
		seg := fullFunc[lastSlash+1:] // e.g. "payment.(*Service).Charge"
		pkgName, _, _ := strings.Cut(seg, ".")
		pkgPath = fullFunc[:lastSlash+1] + pkgName
	} else if pkg, _, found := strings.Cut(fullFunc, "."); found {
		pkgPath = pkg
	}
	relPkgPath := strings.TrimPrefix(pkgPath, module+"/")

	return StackFrame{
		Func: strippedFunc,
		File: relPkgPath + "/" + filepath.Base(file),
		Line: line,
	}
}

// formatStack renders frames as a human-readable, one-frame-per-line trace:
//
//	internal/payment.(*Service).Charge(internal/payment/service.go:87)
//	internal/order.(*Service).CreateOrder(internal/order/service.go:42)
func formatStack(frames []StackFrame) string {
	var b strings.Builder
	for i, f := range frames {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(f.Func)
		b.WriteByte('(')
		b.WriteString(f.File)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(f.Line))
		b.WriteByte(')')
	}
	return b.String()
}
