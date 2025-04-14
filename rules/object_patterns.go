package rules

import (
	"context"
	"fmt"
	"regexp"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

func validatePatternProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	patternProps, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "patternProperties must be an object", Value: schemaValue, Tag: "patternProperties"}
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "patternProperties can only be applied to objects", Value: value, Tag: "patternProperties"}
	}

	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "validator not found in context", Tag: "patternProperties"}
	}

	// 创建新的上下文，正确存储 patternProperties
	newCtx := context.WithValue(ctx, "patternProperties", patternProps)

	// 编译所有模式
	compiledPatterns, err := compilePatterns(patternProps)
	if err != nil {
		return false, &errors.ValidationError{Path: path, Message: err.Error(), Value: patternProps, Tag: "patternProperties"}
	}

	// 对每个属性检查所有模式
	for propName, propValue := range obj {
		for pattern, re := range compiledPatterns {
			if re.MatchString(propName) {
				propSchema, ok := patternProps[pattern]
				if !ok {
					continue
				}

				propSchemaObj, ok := propSchema.(map[string]interface{})
				if !ok {
					continue
				}

				propPath := fmt.Sprintf("%s.%s", path, propName)

				// 验证属性
				isValid, err := validatePropertyWithSchema(newCtx, propValue, propSchemaObj, propPath, registry)
				if !isValid || err != nil {
					return false, err
				}
			}
		}
	}

	return true, nil
}

func validateAdditionalProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "additionalProperties can only be applied to objects", Value: value, Tag: "additionalProperties"}
	}

	// 从上下文获取属性和模式属性
	properties, _ := ctx.Value("properties").(map[string]interface{})
	patternProperties, _ := ctx.Value("patternProperties").(map[string]interface{})

	// 编译模式属性的正则表达式
	var patterns []*regexp.Regexp
	if patternProperties != nil {
		compiledPatterns, err := compilePatterns(patternProperties)
		if err != nil {
			// 忽略无效的模式，继续处理
			patterns = make([]*regexp.Regexp, 0)
		} else {
			patterns = make([]*regexp.Regexp, 0, len(compiledPatterns))
			for _, re := range compiledPatterns {
				patterns = append(patterns, re)
			}
		}
	}

	// 找出额外的属性
	additionalProps := make(map[string]interface{})
	for propName, propValue := range obj {
		// 跳过在properties中定义的属性
		if _, exists := properties[propName]; exists {
			continue
		}

		// 检查是否匹配任何模式属性
		matched := false
		for _, re := range patterns {
			if re.MatchString(propName) {
				matched = true
				break
			}
		}

		if matched {
			continue
		}

		// 这是一个额外属性
		additionalProps[propName] = propValue
	}

	// 如果没有额外属性，验证通过
	if len(additionalProps) == 0 {
		return true, nil
	}

	// 处理不同类型的additionalProperties值
	switch schemaValue.(type) {
	case bool:
		// 布尔值：true允许额外属性，false禁止
		allowed, _ := schemaValue.(bool)
		if !allowed {
			return false, &errors.ValidationError{
				Path:    path,
				Message: "additional properties are not allowed",
				Value:   additionalProps,
				Tag:     "additionalProperties",
			}
		}
		return true, nil

	case map[string]interface{}:
		// 对象：使用schema验证额外属性
		schema, _ := schemaValue.(map[string]interface{})
		registry, ok := ctx.Value("validator").(ValidatorRegistry)
		if !ok {
			return false, &errors.ValidationError{
				Path:    path,
				Message: "validator not found in context",
				Tag:     "additionalProperties",
			}
		}

		// 验证每个额外属性
		for propName, propValue := range additionalProps {
			propPath := fmt.Sprintf("%s.%s", path, propName)

			// 直接遍历schema中的关键字，保持原始错误消息格式
			for keyword, keywordValue := range schema {
				if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
					continue
				}
				validator := registry.GetValidator(keyword)
				if validator == nil {
					continue
				}
				isValid, err := validator(ctx, propValue, keywordValue, propPath)
				if !isValid || err != nil {
					return false, &errors.ValidationError{
						Path:    propPath,
						Message: fmt.Sprintf("additional property validation failed for keyword '%s'", keyword),
						Value:   propValue,
						Tag:     keyword,
					}
				}
			}
		}
		return true, nil

	default:
		return false, &errors.ValidationError{
			Path:    path,
			Message: "additionalProperties must be a boolean or an object",
			Value:   schemaValue,
			Tag:     "additionalProperties",
		}
	}
}

func validateSchemaForProperty(ctx context.Context, value interface{}, schema map[string]interface{}, path string, registry ValidatorRegistry) (bool, error) {
	for keyword, keywordValue := range schema {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		validator := registry.GetValidator(keyword)
		if validator == nil {
			continue
		}
		isValid, err := validator(ctx, value, keywordValue, path)
		if !isValid || err != nil {
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed for keyword '%s'", keyword),
				Value:   value,
				Tag:     keyword,
			}
		}
	}
	return true, nil
}

// compilePatterns 编译正则表达式模式
func compilePatterns(patterns map[string]interface{}) (map[string]*regexp.Regexp, error) {
	result := make(map[string]*regexp.Regexp)
	for pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %s", err.Error())
		}
		result[pattern] = re
	}
	return result, nil
}

// validatePropertyWithSchema 使用schema验证属性
func validatePropertyWithSchema(ctx context.Context, propValue interface{}, propSchema map[string]interface{}, propPath string, registry ValidatorRegistry) (bool, error) {
	for keyword, keywordValue := range propSchema {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		validator := registry.GetValidator(keyword)
		if validator == nil {
			continue
		}
		isValid, err := validator(ctx, propValue, keywordValue, propPath)
		if !isValid || err != nil {
			return false, &errors.ValidationError{
				Path:    propPath,
				Message: fmt.Sprintf("property validation failed for keyword '%s'", keyword),
				Value:   propValue,
				Tag:     keyword,
			}
		}
	}
	return true, nil
}

func validateWithSchema(ctx context.Context, value interface{}, schema map[string]interface{}, path string, registry ValidatorRegistry) (bool, *errors.ValidationError) {
	validators := make(map[string]RuleFunc, len(schema))

	for keyword := range schema {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		if validator := registry.GetValidator(keyword); validator != nil {
			validators[keyword] = validator
		}
	}

	// 执行验证
	for keyword, validator := range validators {
		keywordValue := schema[keyword]
		isValid, err := validator(ctx, value, keywordValue, path)
		if err != nil {
			validErr, ok := err.(*errors.ValidationError)
			if ok {
				return false, validErr
			}
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed: %v", err),
				Value:   value,
				Tag:     keyword,
			}
		}
		if !isValid {
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed for keyword '%s'", keyword),
				Value:   value,
				Tag:     keyword,
			}
		}
	}
	return true, nil
}
