package comparators

import (
	"fmt"
	"reflect"
)

// RegisterBuiltInComparators 注册内置比较器
func RegisterBuiltInComparators(registry ComparatorRegistry) error {
	comparators := []struct {
		name string
		fn   CompareFunc
	}{
		{name: "eq", fn: equal},
		{name: "ne", fn: notEqual},
		{name: "gt", fn: greaterThan},
		{name: "ge", fn: greaterThanOrEqual},
		{name: "lt", fn: lessThan},
		{name: "le", fn: lessThanOrEqual},
	}

	// 注册比较器
	for _, c := range comparators {
		if err := registry.RegisterComparator(c.name, c.fn); err != nil {
			return fmt.Errorf("failed to register comparator %s: %w", c.name, err)
		}
	}
	return nil
}

// equal 比较两个值是否相等
func equal(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	return reflect.DeepEqual(a, b)
}

// notEqual 比较两个值是否不相等
func notEqual(a, b interface{}) bool {
	return !equal(a, b)
}

// greaterThan 比较 a > b
func greaterThan(a, b interface{}) bool {
	return compareNumeric(a, b, func(fa, fb float64) bool { return fa > fb })
}

// greaterThanOrEqual 比较 a >= b
func greaterThanOrEqual(a, b interface{}) bool {
	return compareNumeric(a, b, func(fa, fb float64) bool { return fa >= fb })
}

// lessThan 比较 a < b
func lessThan(a, b interface{}) bool {
	return compareNumeric(a, b, func(fa, fb float64) bool { return fa < fb })
}

// lessThanOrEqual 比较 a <= b
func lessThanOrEqual(a, b interface{}) bool {
	return compareNumeric(a, b, func(fa, fb float64) bool { return fa <= fb })
}

// compareNumeric 辅助函数，处理数值比较
func compareNumeric(a, b interface{}, cmp func(float64, float64) bool) bool {
	fa, ok := toFloat64(a)
	if !ok {
		return false
	}
	fb, ok := toFloat64(b)
	if !ok {
		return false
	}
	return cmp(fa, fb)
}

// toFloat64 将 interface{} 转换为 float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}
