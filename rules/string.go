package rules

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// 注册字符串相关规则
func registerStringRules(registry ValidatorRegistry) {
	registry.RegisterValidator("minLength", validateMinLength)
	registry.RegisterValidator("maxLength", validateMaxLength)
	registry.RegisterValidator("pattern", validatePattern)
}

// validateMinLength 验证字符串最小长度
func validateMinLength(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	if reflect.TypeOf(value).Kind() != reflect.String {
		return false, &errors.ValidationError{Path: path, Message: "must be a string", Tag: "minLength"}
	}
	str := value.(string)
	min, ok := toInt(schemaValue)
	if !ok || min < 0 {
		return false, &errors.ValidationError{Path: path, Message: "minLength must be a non-negative integer", Tag: "minLength"}
	}
	if len(str) < min {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("length less than minimum %d", min), Tag: "minLength", Param: fmt.Sprintf("%d", min)}
	}
	return true, nil
}

// validateMaxLength 验证字符串最大长度
func validateMaxLength(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	if reflect.TypeOf(value).Kind() != reflect.String {
		return false, &errors.ValidationError{Path: path, Message: "must be a string", Tag: "maxLength"}
	}
	str := value.(string)
	max, ok := toInt(schemaValue)
	if !ok || max < 0 {
		return false, &errors.ValidationError{Path: path, Message: "maxLength must be a non-negative integer", Tag: "maxLength"}
	}
	if len(str) > max {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("length greater than maximum %d", max), Tag: "maxLength", Param: fmt.Sprintf("%d", max)}
	}
	return true, nil
}

// validatePattern 验证字符串是否匹配正则表达式
func validatePattern(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	if reflect.TypeOf(value).Kind() != reflect.String {
		return false, &errors.ValidationError{Path: path, Message: "must be a string", Tag: "pattern"}
	}
	str := value.(string)
	pattern, ok := toString(schemaValue)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "pattern must be a string", Tag: "pattern"}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("invalid pattern: %v", err), Tag: "pattern"}
	}
	if !re.MatchString(str) {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("does not match pattern %s", pattern), Tag: "pattern", Param: pattern}
	}

	return true, nil
}
