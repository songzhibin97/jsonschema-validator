package comparators

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewComparator(t *testing.T) {
	fn := func(a, b interface{}) bool { return true }
	comp := NewComparator("testComp", fn)

	assert.Equal(t, "testComp", comp.Name(), "comparator name mismatch")
	assert.True(t, comp.Compare(1, 2), "comparator should return true")
}

func TestBaseComparator_NilFunc(t *testing.T) {
	comp := &BaseComparator{name: "nilComp", fn: nil}
	assert.False(t, comp.Compare(1, 2), "nil CompareFunc should return false")
}

func TestGetGeComparator(t *testing.T) {
	fn := GetGeComparator()
	tests := []struct {
		name        string
		a           interface{}
		b           interface{}
		expectValid bool
	}{
		{
			name:        "Int greater or equal",
			a:           10,
			b:           5,
			expectValid: true,
		},
		{
			name:        "Float equal",
			a:           5.0,
			b:           5.0,
			expectValid: true,
		},
		{
			name:        "Mixed types",
			a:           int64(10),
			b:           float32(5.5),
			expectValid: true,
		},
		{
			name:        "Non-numeric",
			a:           "string",
			b:           5,
			expectValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fn(tt.a, tt.b)
			assert.Equal(t, tt.expectValid, result, "comparison result mismatch")
		})
	}
}

func TestToFloat64_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{
			name:     "Max int64",
			input:    int64(math.MaxInt64),
			expected: float64(math.MaxInt64),
			ok:       true,
		},
		{
			name:     "Complex number",
			input:    complex(1, 2),
			expected: 0,
			ok:       false,
		},
		{
			name:     "Struct",
			input:    struct{}{},
			expected: 0,
			ok:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toFloat64(tt.input)
			assert.Equal(t, tt.ok, ok, "validity mismatch")
			if ok {
				assert.Equal(t, tt.expected, result, "value mismatch")
			}
		})
	}
}
