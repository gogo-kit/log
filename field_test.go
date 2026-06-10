package log

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFieldConstructors(t *testing.T) {
	tests := []struct {
		name      string
		field     Field
		wantKey   string
		wantValue any
	}{
		{"String", String("k", "v"), "k", "v"},
		{"Int", Int("n", 5), "n", 5},
		{"Int64", Int64("n", int64(7)), "n", int64(7)},
		{"Float64", Float64("f", 1.5), "f", 1.5},
		{"Bool", Bool("b", true), "b", true},
		{"Duration", Duration("d", 1500*time.Millisecond), "d", int64(1500)},
		{"Any", Any("a", []int{1}), "a", []int{1}},
		{"RequestID", RequestID("req-1"), "request_id", "req-1"},
		{"UserID", UserID("u-1"), "user_id", "u-1"},
		{"Event", Event("ORDER_CREATED"), "event", "ORDER_CREATED"},
		{"ResourceID", ResourceID("ord-1"), "resource_id", "ord-1"},
		{"Err", Err(errors.New("boom")), "error_message", "boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKey, tt.field.Key)
			assert.Equal(t, tt.wantValue, tt.field.Value)
		})
	}
}

func TestFieldsToMap(t *testing.T) {
	m := fieldsToMap(String("a", "1"), Int("b", 2))
	assert.Equal(t, map[string]any{"a": "1", "b": 2}, m)
}
