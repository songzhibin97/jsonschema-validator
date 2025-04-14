package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// 注册格式验证相关规则
func registerFormatRules(registry ValidatorRegistry) {
	registry.RegisterValidator("format", validateFormat)
}

// formatValidatorMap 保存所有支持的格式验证函数
var formatValidatorMap = map[string]func(string) bool{
	"email":     validateEmail,
	"date-time": validateDateTime,
	"date":      validateDate,
	"time":      validateTime,
	"uri":       validateURI,
	"hostname":  validateHostname,
	"ipv4":      validateIPv4,
	"ipv6":      validateIPv6,
	"uuid":      validateUUID,
}

// validateFormat 验证字符串格式
func validateFormat(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema中的格式
	format, ok := schemaValue.(string)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "format must be a string",
			Value:   schemaValue,
			Tag:     "format",
		}
	}

	// 获取待验证的字符串
	str, ok := value.(string)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value must be a string",
			Value:   value,
			Tag:     "format",
		}
	}

	// 查找格式验证函数
	validator, exists := formatValidatorMap[format]
	if !exists {
		// 默认严格模式
		mode, _ := ctx.Value("validationMode").(int)
		if mode != 1 { // 非宽松模式，视为严格模式
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("unknown format: %s", format),
				Value:   value,
				Tag:     "format",
				Param:   format,
			}
		}
		return true, nil
	}

	// 执行格式验证
	if !validator(str) {
		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("invalid %s format", format),
			Value:   value,
			Tag:     "format",
			Param:   format,
		}
	}

	return true, nil
}

// RegisterFormatValidator 注册自定义格式验证器
func RegisterFormatValidator(name string, validator func(string) bool) {
	if validator != nil {
		formatValidatorMap[name] = validator
	}
}
