package rules

import (
	"context"
	"fmt"
	"strings"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// RuleFunc 定义了验证规则函数的签名
type RuleFunc func(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error)

// Rule 接口定义了验证规则的行为
type Rule interface {
	Name() string
	Validate(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error)
}

// BaseRule 是所有规则的基础实现
type BaseRule struct {
	name string
	fn   RuleFunc
}

// NewRule 创建一个新的规则
func NewRule(name string, fn RuleFunc) Rule {
	return &BaseRule{
		name: name,
		fn:   fn,
	}
}

// Name 返回规则的名称
func (r *BaseRule) Name() string {
	return r.name
}

// Validate 执行验证
func (r *BaseRule) Validate(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	return r.fn(ctx, value, schemaValue, path)
}

// typeValidator 验证类型
func typeValidator(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	schemaType, ok := schemaValue.(string)
	if !ok {
		return false, fmt.Errorf("schema type must be string")
	}
	switch schemaType {
	case "string":
		if _, ok := value.(string); !ok {
			return false, &errors.ValidationError{
				Path:    path,
				Message: "expected string",
				Tag:     "type",
			}
		}
	case "integer":
		if _, ok := value.(float64); !ok {
			return false, &errors.ValidationError{
				Path:    path,
				Message: "expected integer",
				Tag:     "type",
			}
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return false, &errors.ValidationError{
				Path:    path,
				Message: "expected object",
				Tag:     "type",
			}
		}
	}
	return true, nil
}

// requiredValidator 验证必需字段
func requiredValidator(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	if schemaValue == nil {
		return true, nil
	}
	requiredFields, ok := schemaValue.([]string)
	if !ok {
		return false, fmt.Errorf("required must be an array of strings")
	}
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value must be an object",
			Tag:     "required",
		}
	}
	for _, field := range requiredFields {
		if _, exists := obj[field]; !exists {
			return false, &errors.ValidationError{
				Path:    path + "." + field,
				Message: fmt.Sprintf("required property '%s' is missing", field),
				Tag:     "required",
			}
		}
	}
	return true, nil
}

// minimumValidator 验证最小值
func minimumValidator(ctx context.Context, value interface{}, schema interface{}, path string) (bool, error) {
	var schemaNum float64
	switch v := schema.(type) {
	case int:
		schemaNum = float64(v)
	case float64:
		schemaNum = v
	default:
		return false, &errors.ValidationError{
			Path:    path,
			Message: "minimum must be a number",
			Tag:     "minimum",
		}
	}
	valueNum, ok := toFloat64(value)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value must be a number",
			Tag:     "minimum",
			Value:   value,
		}
	}
	if valueNum < schemaNum {
		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("value %v is less than minimum %v", valueNum, schemaNum),
			Tag:     "minimum",
			Value:   value,
		}
	}
	return true, nil
}

// enumValidator 验证枚举值
func enumValidator(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	enumValues, ok := schemaValue.([]string)
	if !ok {
		return false, fmt.Errorf("enum must be an array of strings")
	}
	strVal, ok := value.(string)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value must be a string",
			Tag:     "enum",
		}
	}
	for _, v := range enumValues {
		if v == strVal {
			return true, nil
		}
	}
	return false, &errors.ValidationError{
		Path:    path,
		Message: fmt.Sprintf("value must be one of: %s", strings.Join(enumValues, ", ")),
		Tag:     "enum",
	}
}

// ValidateNotNil 验证值不为nil
func ValidateNotNil(value interface{}, path string, msg string) (bool, error) {
	if value == nil {
		return false, &errors.ValidationError{
			Path:    path,
			Message: msg,
			Tag:     "not_nil",
		}
	}
	return true, nil
}

// ValidationFunc 是验证函数的简化类型
type ValidationFunc func(value interface{}, path string) (bool, error)

// WrapValidation 包装验证函数
func WrapValidation(fn ValidationFunc, nilMsg string) ValidationFunc {
	return func(value interface{}, path string) (bool, error) {
		if value == nil {
			return false, &errors.ValidationError{
				Path:    path,
				Message: nilMsg,
				Tag:     "not_nil",
			}
		}
		return fn(value, path)
	}
}

// DefaultErrorMessage 获取默认错误消息
func DefaultErrorMessage(keyword string) string {
	return "failed to validate " + keyword
}
