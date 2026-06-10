package log

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureStackKeepsOnlyAppFrames(t *testing.T) {
	frames := stackFromError(originDeep())
	require.NotEmpty(t, frames)

	for _, f := range frames {
		// runtime.* / testing.* / reflect.* are excluded by module filtering.
		assert.False(t, strings.HasPrefix(f.Func, "runtime."), "unexpected runtime frame: %s", f.Func)
		assert.False(t, strings.HasPrefix(f.Func, "testing."), "unexpected testing frame: %s", f.Func)
		assert.NotEmpty(t, f.File)
		assert.Positive(t, f.Line)
	}
}

func TestCaptureStackStartsAtOrigin(t *testing.T) {
	frames := stackFromError(originDeep())
	require.NotEmpty(t, frames)
	// Wrap is called inside originDeeper, so it must be the first app frame.
	assert.Contains(t, frames[0].Func, "originDeeper")
}

func TestFormatStackShape(t *testing.T) {
	frames := []StackFrame{
		{Func: "internal/order.(*Service).CreateOrder", File: "internal/order/service.go", Line: 42},
		{Func: "api.(*Handler).Create", File: "api/handler.go", Line: 28},
	}
	got := formatStack(frames)
	lines := strings.Split(got, "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "internal/order.(*Service).CreateOrder(internal/order/service.go:42)", lines[0])
	assert.Equal(t, "api.(*Handler).Create(api/handler.go:28)", lines[1])
}

func TestIsAppFrame(t *testing.T) {
	const module = "github.com/acme/svc"
	tests := []struct {
		fn   string
		want bool
	}{
		{"github.com/acme/svc/internal/order.Create", true},
		{"main.main", true},
		{"main", true},
		{"runtime.goexit", false},
		{"net/http.(*conn).serve", false},
		{"github.com/gin-gonic/gin.(*Engine).handleHTTPRequest", false},
	}
	for _, tt := range tests {
		t.Run(tt.fn, func(t *testing.T) {
			assert.Equal(t, tt.want, isAppFrame(tt.fn, module))
		})
	}
}

func TestParseFrameStripsModule(t *testing.T) {
	module := "github.com/acme/svc"
	f := parseFrame(module+"/internal/payment.(*Service).Charge", "/abs/path/service.go", 87, module)
	assert.Equal(t, "internal/payment.(*Service).Charge", f.Func)
	assert.Equal(t, "internal/payment/service.go", f.File)
	assert.Equal(t, 87, f.Line)
}
