package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// validateRequired 验证对象必需包含的属性
func validateRequired(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	reqFields, ok := schemaValue.([]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "required must be an array of strings",
			Value:   schemaValue,
			Tag:     "required",
		}
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "required can only be applied to objects",
			Value:   value,
			Tag:     "required",
		}
	}

	for _, field := range reqFields {
		fieldStr, ok := field.(string)
		if !ok {
			continue
		}

		if _, exists := obj[fieldStr]; !exists {
			return false, &errors.ValidationError{
				Path:    fmt.Sprintf("%s.%s", path, fieldStr),
				Message: fmt.Sprintf("required property '%s' is missing", fieldStr),
				Value:   obj,
				Tag:     "required",
				Param:   fieldStr,
			}
		}
	}

	return true, nil
}

// validateProperties 验证对象的属性
func validateProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	properties, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "properties must be an object",
			Value:   schemaValue,
			Tag:     "properties",
		}
	}

	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "properties can only be applied to objects",
			Value:   value,
			Tag:     "properties",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "properties",
		}
	}

	// 将属性放入上下文，便于additionalProperties使用
	ctx = context.WithValue(ctx, "properties", properties)

	// 遍历对象的属性
	for propName, propSchema := range properties {
		propValue, exists := obj[propName]
		if !exists {
			// 属性不存在，跳过验证（required会处理必需属性）
			continue
		}

		propSchemaObj, ok := propSchema.(map[string]interface{})
		if !ok {
			continue
		}

		propPath := fmt.Sprintf("%s.%s", path, propName)

		// 遍历属性schema中的验证关键字
		for keyword, keywordValue := range propSchemaObj {
			// 跳过非验证关键字
			if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
				continue
			}

			validator := registry.GetValidator(keyword)
			if validator == nil {
				// 未知的关键字
				continue
			}

			isValid, err := validator(ctx, propValue, keywordValue, propPath)
			if err != nil {
				return false, err
			}

			if !isValid {
				return false, &errors.ValidationError{
					Path:    propPath,
					Message: fmt.Sprintf("property validation failed for keyword '%s'", keyword),
					Value:   propValue,
					Tag:     keyword,
				}
			}
		}
	}

	return true, nil
}
