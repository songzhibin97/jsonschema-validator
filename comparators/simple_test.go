package comparators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSimpleComparatorRegistry(t *testing.T) {
	registry := NewSimpleComparatorRegistry()
	assert.NotNil(t, registry.comparators, "comparators map should be initialized")
	assert.Empty(t, registry.comparators, "comparators map should be empty")
}

func TestRegisterComparator_EdgeCases(t *testing.T) {
	registry := NewSimpleComparatorRegistry()
	tests := []struct {
		name      string
		compName  string
		fn        CompareFunc
		expectErr bool
	}{
		{
			name:      "Empty name",
			compName:  "",
			fn:        func(a, b interface{}) bool { return true },
			expectErr: false, // Current impl allows empty names; consider if this should error
		},
		{
			name:      "Nil function",
			compName:  "nilFn",
			fn:        nil,
			expectErr: true,
		},
		{
			name:      "Duplicate registration",
			compName:  "dup",
			fn:        func(a, b interface{}) bool { return true },
			expectErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Duplicate registration" {
				_ = registry.RegisterComparator(tt.compName, tt.fn)
			}
			err := registry.RegisterComparator(tt.compName, tt.fn)
			if tt.expectErr {
				assert.Error(t, err, "expected error for %s", tt.name)
			} else {
				assert.NoError(t, err, "unexpected error for %s", tt.name)
			}
		})
	}
}

func TestGetComparator_Missing(t *testing.T) {
	registry := NewSimpleComparatorRegistry()
	fn := registry.GetComparator("missing")
	assert.Nil(t, fn, "should return nil for unregistered comparator")
}
