package rules

import (
	"context"
	"fmt"
	"math"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// 注册数值相关规则
func registerNumberRules(registry ValidatorRegistry) {
	registry.RegisterValidator("minimum", validateMinimum)
	registry.RegisterValidator("maximum", validateMaximum)
	registry.RegisterValidator("exclusiveMinimum", validateExclusiveMinimum)
	registry.RegisterValidator("exclusiveMaximum", validateExclusiveMaximum)
	registry.RegisterValidator("multipleOf", validateMultipleOf)
}

// validateMinimum 验证数值最小值
func validateMinimum(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be a number", Tag: "minimum"}
	}
	min, ok := toFloat64(schemaValue)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "minimum must be a number", Tag: "minimum"}
	}
	if v < min {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("less than minimum %v", min), Tag: "minimum", Param: fmt.Sprintf("%v", min)}
	}
	return true, nil
}

// validateMaximum 验证数值最大值
func validateMaximum(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be a number", Tag: "maximum"}
	}
	max, ok := toFloat64(schemaValue)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "maximum must be a number", Tag: "maximum"}
	}
	if v > max {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("greater than maximum %v", max), Tag: "maximum", Param: fmt.Sprintf("%v", max)}
	}
	return true, nil
}

// validateExclusiveMinimum 验证数值严格大于最小值
func validateExclusiveMinimum(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be a number", Tag: "exclusiveMinimum"}
	}
	min, ok := toFloat64(schemaValue)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "exclusiveMinimum must be a number", Tag: "exclusiveMinimum"}
	}
	if v <= min {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("less than or equal to exclusive minimum %v", min), Tag: "exclusiveMinimum", Param: fmt.Sprintf("%v", min)}
	}
	return true, nil
}

// validateExclusiveMaximum 验证数值严格小于最大值
func validateExclusiveMaximum(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	v, ok := toFloat64(value)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be a number", Tag: "exclusiveMaximum"}
	}
	max, ok := toFloat64(schemaValue)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "exclusiveMaximum must be a number", Tag: "exclusiveMaximum"}
	}
	if v >= max {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("greater than or equal to exclusive maximum %v", max), Tag: "exclusiveMaximum", Param: fmt.Sprintf("%v", max)}
	}
	return true, nil
}

// validateMultipleOf 验证数值是否是指定值的倍数
func validateMultipleOf(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema中的除数
	divisor, ok := toFloat64(schemaValue)
	if !ok || divisor <= 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "multipleOf must be a positive number",
			Value:   schemaValue,
			Tag:     "multipleOf",
		}
	}

	// 获取待验证的值
	val, ok := toFloat64(value)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "multipleOf can only be applied to numbers",
			Value:   value,
			Tag:     "multipleOf",
		}
	}

	// 处理浮点数精度问题
	ratio := val / divisor
	if math.Abs(ratio-math.Round(ratio)) > 1e-10 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("value %v is not a multiple of %v", value, divisor),
			Value:   value,
			Tag:     "multipleOf",
			Param:   fmt.Sprintf("%v", divisor),
		}
	}

	return true, nil
}
