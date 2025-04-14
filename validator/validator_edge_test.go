package validator

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/songzhibin97/jsonschema-validator/comparators"
	"github.com/songzhibin97/jsonschema-validator/rules"
	"github.com/stretchr/testify/assert"
)

// 测试 RegisterValidator 和 RegisterComparator 的边缘情况
func TestRegisterValidatorEdgeCases(t *testing.T) {
	v := New()
	tests := []struct {
		name      string
		validator string
		fn        rules.RuleFunc
		expectErr bool
		errMsg    string
	}{
		{
			name:      "空名称",
			validator: "",
			fn: func(ctx context.Context, value interface{}, schema interface{}, path string) (bool, error) {
				return true, nil
			},
			expectErr: true,
			errMsg:    "validator name cannot be empty",
		},
		{
			name:      "空函数",
			validator: "test",
			fn:        nil,
			expectErr: true,
			errMsg:    "validator function cannot be nil",
		},
		{
			name:      "重复注册",
			validator: "dup",
			fn: func(ctx context.Context, value interface{}, schema interface{}, path string) (bool, error) {
				return true, nil
			},
			expectErr: true,
			errMsg:    "validator dup already registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "重复注册" {
				_ = v.RegisterValidator(tt.validator, tt.fn)
			}
			err := v.RegisterValidator(tt.validator, tt.fn)
			if tt.expectErr {
				assert.Error(t, err, "预期返回错误")
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMsg, "错误消息不匹配")
				}
			} else {
				assert.NoError(t, err, "未预期错误")
			}
		})
	}
}

func TestRegisterComparatorEdgeCases(t *testing.T) {
	v := New()
	tests := []struct {
		name       string
		comparator string
		fn         comparators.CompareFunc
		expectErr  bool
		errMsg     string
	}{
		{
			name:       "空名称",
			comparator: "",
			fn:         func(a, b interface{}) bool { return true },
			expectErr:  true,
			errMsg:     "comparator name cannot be empty",
		},
		{
			name:       "空函数",
			comparator: "test",
			fn:         nil,
			expectErr:  true,
			errMsg:     "comparator function cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.RegisterComparator(tt.comparator, tt.fn)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 测试 parseTag 的边缘情况
func TestParseTag(t *testing.T) {
	v := New()
	tests := []struct {
		name     string
		tag      string
		expected map[string]interface{}
	}{
		{
			name:     "空标签",
			tag:      "",
			expected: map[string]interface{}{},
		},
		{
			name: "复杂标签",
			tag:  "required,min=5,max=10,enum=a|b|c",
			expected: map[string]interface{}{
				"required": true,
				"min":      5,
				"max":      10,
				"enum":     []string{"a", "b", "c"},
			},
		},
		{
			name: "无效数字",
			tag:  "min=abc",
			expected: map[string]interface{}{
				"min": "abc",
			},
		},
		{
			name: "特殊字符",
			tag:  "pattern=^a.*b$,format=email",
			expected: map[string]interface{}{
				"pattern": "^a.*b$",
				"format":  "email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.parseTag(tt.tag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 测试 isZero 的复杂类型
func TestIsZero(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "空字符串",
			value:    "",
			expected: true,
		},
		{
			name:     "零结构体",
			value:    struct{}{},
			expected: true,
		},
		{
			name:     "非空切片",
			value:    []int{1},
			expected: false,
		},
		{
			name:     "空指针",
			value:    (*int)(nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isZero(reflect.ValueOf(tt.value))
			assert.Equal(t, tt.expected, result)
		})
	}
}

// 测试缓存清理
func TestClearCache(t *testing.T) {
	v := New(WithCaching(true))
	schemaJSON := `{"type":"object"}`

	// 缓存 schema
	_, err := v.CompileSchema(schemaJSON)
	assert.NoError(t, err)
	_, ok := v.cache.Load(schemaJSON)
	assert.True(t, ok, "缓存应存在")

	// 清理缓存
	v.ClearCache()
	_, ok = v.cache.Load(schemaJSON)
	assert.False(t, ok, "缓存应被清理")
}

// 测试并发注册
func TestConcurrentRegistration(t *testing.T) {
	v := New()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := v.RegisterValidator(fmt.Sprintf("rule%d", i), func(ctx context.Context, value interface{}, schema interface{}, path string) (bool, error) {
				return true, nil
			})
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	// 验证注册结果
	for i := 0; i < 10; i++ {
		fn := v.GetValidator(fmt.Sprintf("rule%d", i))
		assert.NotNil(t, fn, "验证器 rule%d 应已注册", i)
	}
}
