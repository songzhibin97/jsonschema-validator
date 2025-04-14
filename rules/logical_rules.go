package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// 注册逻辑组合相关规则
func registerLogicalRules(registry ValidatorRegistry) {
	registry.RegisterValidator("allOf", validateAllOf)
	registry.RegisterValidator("anyOf", validateAnyOf)
	registry.RegisterValidator("oneOf", validateOneOf)
	registry.RegisterValidator("not", validateNot)
}

// validateAllOf 验证数据满足所有指定的schema
func validateAllOf(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema数组
	schemas, ok := schemaValue.([]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "allOf must be an array",
			Value:   schemaValue,
			Tag:     "allOf",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "allOf",
		}
	}

	// 如果schemas为空，返回错误
	if len(schemas) == 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "allOf cannot be empty",
			Value:   schemaValue,
			Tag:     "allOf",
		}
	}

	// 验证数据满足所有schema
	for i, schema := range schemas {
		schemaObj, ok := schema.(map[string]interface{})
		if !ok {
			return false, &errors.ValidationError{
				Path:    fmt.Sprintf("%s.allOf[%d]", path, i),
				Message: "schema must be an object",
				Value:   schema,
				Tag:     "allOf",
			}
		}

		schemaPath := fmt.Sprintf("%s.allOf[%d]", path, i)

		// 遍历schema中的验证关键字
		for keyword, keywordValue := range schemaObj {
			if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
				continue
			}

			validator := registry.GetValidator(keyword)
			if validator == nil {
				continue
			}

			isValid, err := validator(ctx, value, keywordValue, schemaPath)
			if err != nil {
				return false, &errors.ValidationError{
					Path:    schemaPath,
					Message: fmt.Sprintf("failed to validate against schema at allOf[%d] for keyword '%s': %v", i, keyword, err),
					Value:   value,
					Tag:     "allOf",
				}
			}

			if !isValid {
				return false, &errors.ValidationError{
					Path:    schemaPath,
					Message: fmt.Sprintf("failed to validate against schema at allOf[%d] for keyword '%s'", i, keyword),
					Value:   value,
					Tag:     "allOf",
				}
			}
		}
	}

	return true, nil
}

// validateAnyOf 验证数据满足至少一个指定的schema
func validateAnyOf(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema数组
	schemas, ok := schemaValue.([]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "anyOf must be an array",
			Value:   schemaValue,
			Tag:     "anyOf",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "anyOf",
		}
	}

	// 如果schemas为空，返回错误
	if len(schemas) == 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "anyOf cannot be empty",
			Value:   schemaValue,
			Tag:     "anyOf",
		}
	}

	// 记录验证失败的错误
	var validationErrors []errors.ValidationError

	// 验证数据满足至少一个schema
	for i, schema := range schemas {
		schemaObj, ok := schema.(map[string]interface{})
		if !ok {
			validationErrors = append(validationErrors, errors.ValidationError{
				Path:    fmt.Sprintf("%s.anyOf[%d]", path, i),
				Message: "schema must be an object",
				Value:   schema,
				Tag:     "anyOf",
			})
			continue
		}

		schemaPath := fmt.Sprintf("%s.anyOf[%d]", path, i)

		// 使用通用的validateWithSchema函数
		valid, validErr := validateWithSchema(ctx, value, schemaObj, schemaPath, registry)
		if valid {
			// 只要有一个schema验证通过，整体就通过
			return true, nil
		}

		// 记录错误
		if validErr != nil {
			validationErrors = append(validationErrors, *validErr)
		}
	}

	// 如果所有schema都验证失败，返回错误
	return false, &errors.ValidationError{
		Path:    path,
		Message: "value does not match any schema in anyOf",
		Value:   value,
		Tag:     "anyOf",
	}
}

// validateOneOf 验证数据恰好满足一个指定的schema
func validateOneOf(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema数组
	schemas, ok := schemaValue.([]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "oneOf must be an array",
			Value:   schemaValue,
			Tag:     "oneOf",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "oneOf",
		}
	}

	// 如果schemas为空，返回错误
	if len(schemas) == 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "oneOf cannot be empty",
			Value:   schemaValue,
			Tag:     "oneOf",
		}
	}

	// 记录验证失败的错误
	var validationErrors []errors.ValidationError

	// 记录匹配的schema数量
	matchCount := 0

	// 验证数据恰好满足一个schema
	for i, schema := range schemas {
		schemaObj, ok := schema.(map[string]interface{})
		if !ok {
			validationErrors = append(validationErrors, errors.ValidationError{
				Path:    fmt.Sprintf("%s.oneOf[%d]", path, i),
				Message: "schema must be an object",
				Value:   schema,
				Tag:     "oneOf",
			})
			continue
		}

		schemaPath := fmt.Sprintf("%s.oneOf[%d]", path, i)

		// 使用通用的validateWithSchema函数
		valid, validErr := validateWithSchema(ctx, value, schemaObj, schemaPath, registry)
		if valid {
			matchCount++
			if matchCount > 1 {
				return false, &errors.ValidationError{
					Path:    path,
					Message: "value matches more than one schema in oneOf",
					Value:   value,
					Tag:     "oneOf",
				}
			}
		} else if validErr != nil {
			validationErrors = append(validationErrors, *validErr)
		}
	}

	// 检查匹配数量
	if matchCount == 1 {
		return true, nil
	} else {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value does not match any schema in oneOf",
			Value:   value,
			Tag:     "oneOf",
		}
	}
}

// validateNot 验证数据不满足指定的schema
func validateNot(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema
	schema, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "not must be an object",
			Value:   schemaValue,
			Tag:     "not",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "not",
		}
	}

	// 如果schema为空，返回错误
	if len(schema) == 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "not schema cannot be empty",
			Value:   schemaValue,
			Tag:     "not",
		}
	}

	// 使用通用的validateWithSchema函数，但结果取反
	valid, _ := validateWithSchema(ctx, value, schema, path, registry)

	// not验证：如果schema验证通过，则not验证失败；如果schema验证失败，则not验证通过
	if valid {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value must not validate against the schema in not",
			Value:   value,
			Tag:     "not",
		}
	}

	return true, nil
}
