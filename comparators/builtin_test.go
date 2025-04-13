package comparators

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterBuiltInComparators(t *testing.T) {
	registry := NewSimpleComparatorRegistry()

	// 注册内置比较器
	err := RegisterBuiltInComparators(registry)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		comparator  string
		a           interface{}
		b           interface{}
		expectValid bool
	}{
		{
			name:        "Equal integers",
			comparator:  "eq",
			a:           42,
			b:           42,
			expectValid: true,
		},
		{
			name:        "Not equal integers",
			comparator:  "ne",
			a:           42,
			b:           43,
			expectValid: true,
		},
		{
			name:        "Greater than",
			comparator:  "gt",
			a:           10,
			b:           5,
			expectValid: true,
		},
		{
			name:        "Greater than or equal",
			comparator:  "ge",
			a:           5,
			b:           5,
			expectValid: true,
		},
		{
			name:        "Less than",
			comparator:  "lt",
			a:           5,
			b:           10,
			expectValid: true,
		},
		{
			name:        "Less than or equal",
			comparator:  "le",
			a:           5,
			b:           5,
			expectValid: true,
		},
		{
			name:        "Invalid greater than non-numeric",
			comparator:  "gt",
			a:           "string",
			b:           5,
			expectValid: false,
		},
		{
			name:        "Equal nil",
			comparator:  "eq",
			a:           nil,
			b:           nil,
			expectValid: true,
		},
		{
			name:        "Mixed numeric types",
			comparator:  "gt",
			a:           int32(10),
			b:           float64(5.5),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := registry.GetComparator(tt.comparator)
			assert.NotNil(t, fn, "comparator %s should be registered", tt.comparator)
			result := fn(tt.a, tt.b)
			assert.Equal(t, tt.expectValid, result, "comparison result mismatch for %s", tt.name)
		})
	}
}

func TestRegisterBuiltInComparators_Conflict(t *testing.T) {
	registry := NewSimpleComparatorRegistry()

	// 先手动注册一个比较器
	err := registry.RegisterComparator("eq", func(a, b interface{}) bool { return true })
	assert.NoError(t, err)

	// 注册内置比较器，预期失败
	err = RegisterBuiltInComparators(registry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comparator eq already registered")

	// 验证 "eq" 仍保留原始注册
	fn := registry.GetComparator("eq")
	assert.NotNil(t, fn)
	result := fn(1, 2) // 原始 "eq" 总是返回 true
	assert.True(t, result, "original eq comparator should remain")
}

func TestComparatorFunctions(t *testing.T) {
	tests := []struct {
		name        string
		fn          CompareFunc
		a           interface{}
		b           interface{}
		expectValid bool
	}{
		{
			name:        "Equal strings",
			fn:          equal,
			a:           "test",
			b:           "test",
			expectValid: true,
		},
		{
			name:        "Not equal floats",
			fn:          notEqual,
			a:           1.23,
			b:           4.56,
			expectValid: true,
		},
		{
			name:        "Greater than float",
			fn:          greaterThan,
			a:           5.5,
			b:           3.3,
			expectValid: true,
		},
		{
			name:        "Invalid less than string",
			fn:          lessThan,
			a:           "abc",
			b:           "def",
			expectValid: false,
		},
		{
			name:        "Equal zero values",
			fn:          equal,
			a:           0,
			b:           0,
			expectValid: true,
		},
		{
			name:        "Greater than uint",
			fn:          greaterThan,
			a:           uint(10),
			b:           uint(5),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.a, tt.b)
			assert.Equal(t, tt.expectValid, result, "comparison result mismatch for %s", tt.name)
		})
	}
}

func TestConcurrentRegistration(t *testing.T) {
	registry := NewSimpleComparatorRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := RegisterBuiltInComparators(registry)
			if err != nil {
				// 允许部分注册失败（冲突）
				assert.Contains(t, err.Error(), "already registered")
			}
		}()
	}
	wg.Wait()

	// 验证所有比较器已注册
	for _, comp := range []string{"eq", "ne", "gt", "ge", "lt", "le"} {
		fn := registry.GetComparator(comp)
		assert.NotNil(t, fn, "comparator %s should be registered", comp)
	}
}
