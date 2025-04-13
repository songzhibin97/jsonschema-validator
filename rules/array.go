package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// 注册数组相关规则
func registerArrayRules(registry ValidatorRegistry) {
	registry.RegisterValidator("items", validateItems)
	registry.RegisterValidator("minItems", validateMinItems)
	registry.RegisterValidator("maxItems", validateMaxItems)
	registry.RegisterValidator("uniqueItems", validateUniqueItems)
}

// validateItems 验证数组的元素
func validateItems(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取数组
	arr, ok := value.([]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "items can only be applied to arrays",
			Value:   value,
			Tag:     "items",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "items",
		}
	}

	// 处理两种items模式：对象模式和数组模式
	switch schema := schemaValue.(type) {
	case map[string]interface{}:
		// 对象模式：所有元素都使用同一个schema验证
		for i, item := range arr {
			itemPath := fmt.Sprintf("%s[%d]", path, i)

			// 遍历schema中的验证关键字
			for keyword, keywordValue := range schema {
				// 跳过非验证关键字
				if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
					continue
				}

				validator := registry.GetValidator(keyword)
				if validator == nil {
					// 未知的关键字
					continue
				}

				isValid, err := validator(ctx, item, keywordValue, itemPath)
				if err != nil {
					return false, err
				}

				if !isValid {
					return false, &errors.ValidationError{
						Path:    itemPath,
						Message: fmt.Sprintf("array item validation failed for keyword '%s'", keyword),
						Value:   item,
						Tag:     keyword,
					}
				}
			}
		}

	case []interface{}:
		// 数组模式：每个元素都使用对应位置的schema验证
		for i, itemSchema := range schema {
			if i >= len(arr) {
				// 数组元素数量不足
				break
			}

			itemPath := fmt.Sprintf("%s[%d]", path, i)
			item := arr[i]

			itemSchemaObj, ok := itemSchema.(map[string]interface{})
			if !ok {
				continue
			}

			// 遍历schema中的验证关键字
			for keyword, keywordValue := range itemSchemaObj {
				// 跳过非验证关键字
				if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
					continue
				}

				validator := registry.GetValidator(keyword)
				if validator == nil {
					// 未知的关键字
					continue
				}

				isValid, err := validator(ctx, item, keywordValue, itemPath)
				if err != nil {
					return false, err
				}

				if !isValid {
					return false, &errors.ValidationError{
						Path:    itemPath,
						Message: fmt.Sprintf("array item validation failed for keyword '%s'", keyword),
						Value:   item,
						Tag:     keyword,
					}
				}
			}
		}

	default:
		return false, &errors.ValidationError{
			Path:    path,
			Message: "items must be an object or array",
			Value:   schemaValue,
			Tag:     "items",
		}
	}

	return true, nil
}

// validateMinItems 验证数组最小长度
func validateMinItems(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be an array", Tag: "minItems"}
	}
	min, ok := toInt(schemaValue)
	if !ok || min < 0 {
		return false, &errors.ValidationError{Path: path, Message: "minItems must be a non-negative integer", Tag: "minItems"}
	}
	if len(arr) < min {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("fewer items than minimum %d", min), Tag: "minItems", Param: fmt.Sprintf("%d", min)}
	}
	return true, nil
}

// validateMaxItems 验证数组最大长度
func validateMaxItems(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be an array", Tag: "maxItems"}
	}
	max, ok := toInt(schemaValue)
	if !ok || max < 0 {
		return false, &errors.ValidationError{Path: path, Message: "maxItems must be a non-negative integer", Tag: "maxItems"}
	}
	if len(arr) > max {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("more items than maximum %d", max), Tag: "maxItems", Param: fmt.Sprintf("%d", max)}
	}
	return true, nil
}

// validateUniqueItems 验证数组元素的唯一性
func validateUniqueItems(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	enabled, ok := toBool(schemaValue)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "uniqueItems must be a boolean", Tag: "uniqueItems"}
	}
	if !enabled {
		return true, nil
	}
	arr, ok := value.([]interface{})
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "must be an array", Tag: "uniqueItems"}
	}
	seen := make(map[interface{}]struct{})
	for _, item := range arr {
		if _, exists := seen[item]; exists {
			return false, &errors.ValidationError{Path: path, Message: "contains duplicate items", Tag: "uniqueItems"}
		}
		seen[item] = struct{}{}
	}
	return true, nil
}
