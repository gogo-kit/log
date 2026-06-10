package log

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapNilReturnsNil(t *testing.T) {
	assert.Nil(t, Wrap(nil))
}

func TestWrapCapturesStack(t *testing.T) {
	err := originDeep()
	assert.NotEmpty(t, stackFromError(err))
}

func TestWrapIsIdempotent(t *testing.T) {
	first := Wrap(errors.New("boom"))
	second := Wrap(first)
	require.Same(t, first, second, "re-wrapping must preserve the original origin")
}

func TestWrapPreservesUnwrapChain(t *testing.T) {
	sentinel := errors.New("sentinel")
	err := Wrap(fmt.Errorf("context: %w", sentinel))
	assert.ErrorIs(t, err, sentinel)
}

func TestErrorType(t *testing.T) {
	assert.Equal(t, "customErr", errorType(Wrap(&customErr{msg: "x"})))
	assert.Equal(t, "customErr", errorType(&customErr{msg: "x"}))
	assert.Empty(t, errorType(nil))
}

// originDeep simulates an error originating a couple of calls deep.
func originDeep() error  { return originDeeper() }
func originDeeper() error { return Wrap(errors.New("boom")) }
