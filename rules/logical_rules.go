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
			// 跳过非验证关键字
			if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
				continue
			}

			validator := registry.GetValidator(keyword)
			if validator == nil {
				continue
			}

			isValid, err := validator(ctx, value, keywordValue, schemaPath)
			if err != nil {
				// 包装子验证器的错误
				return false, &errors.ValidationError{
					Path:    schemaPath,
					Message: fmt.Sprintf("failed to validate against schema at allOf[%d] for keyword '%s' err:%e", i, keyword, err),
					Value:   value,
					Tag:     keyword,
				}
			}

			if !isValid {
				return false, &errors.ValidationError{
					Path:    schemaPath,
					Message: fmt.Sprintf("failed to validate against schema at allOf[%d] for keyword '%s'", i, keyword),
					Value:   value,
					Tag:     keyword,
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

		// 标记当前schema验证是否通过
		isSchemaValid := true
		var schemaError *errors.ValidationError

		// 遍历schema中的验证关键字
		for keyword, keywordValue := range schemaObj {
			// 跳过非验证关键字
			if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
				continue
			}

			validator := registry.GetValidator(keyword)
			if validator == nil {
				continue
			}

			isValid, err := validator(ctx, value, keywordValue, schemaPath)
			if err != nil {
				validErr, ok := err.(*errors.ValidationError)
				if ok {
					schemaError = validErr
					isSchemaValid = false
					break
				} else {
					return false, err
				}
			}

			if !isValid {
				isSchemaValid = false
				schemaError = &errors.ValidationError{
					Path:    schemaPath,
					Message: fmt.Sprintf("failed to validate against schema at anyOf[%d] for keyword '%s'", i, keyword),
					Value:   value,
					Tag:     keyword,
				}
				break
			}
		}

		// 如果当前schema验证通过，则整体验证通过
		if isSchemaValid {
			return true, nil
		}

		// 记录错误
		if schemaError != nil {
			validationErrors = append(validationErrors, *schemaError)
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

		// 标记当前schema验证是否通过
		isSchemaValid := true
		var schemaError *errors.ValidationError

		// 遍历schema中的验证关键字
		for keyword, keywordValue := range schemaObj {
			// 跳过非验证关键字
			if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
				continue
			}

			validator := registry.GetValidator(keyword)
			if validator == nil {
				continue
			}

			isValid, err := validator(ctx, value, keywordValue, schemaPath)
			if err != nil {
				validErr, ok := err.(*errors.ValidationError)
				if ok {
					schemaError = validErr
					isSchemaValid = false
					break
				} else {
					return false, err
				}
			}

			if !isValid {
				isSchemaValid = false
				schemaError = &errors.ValidationError{
					Path:    schemaPath,
					Message: fmt.Sprintf("failed to validate against schema at oneOf[%d] for keyword '%s'", i, keyword),
					Value:   value,
					Tag:     keyword,
				}
				break
			}
		}

		// 如果当前schema验证通过，增加匹配计数
		if isSchemaValid {
			matchCount++
			if matchCount > 1 {
				return false, &errors.ValidationError{
					Path:    path,
					Message: "value matches more than one schema in oneOf",
					Value:   value,
					Tag:     "oneOf",
				}
			}
		} else if schemaError != nil {
			validationErrors = append(validationErrors, *schemaError)
		}
	}

	// 检查匹配数量
	if matchCount == 1 {
		return true, nil
	} else if matchCount == 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value does not match any schema in oneOf",
			Value:   value,
			Tag:     "oneOf",
		}
	} else {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "value matches more than one schema in oneOf",
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

	// 遍历schema中的验证关键字
	hasValidation := false
	for keyword, keywordValue := range schema {
		// 跳过非验证关键字
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}

		validator := registry.GetValidator(keyword)
		if validator == nil {
			continue
		}

		hasValidation = true
		isValid, err := validator(ctx, value, keywordValue, path)
		if err != nil || !isValid {
			// 子验证器失败或返回false，表示value不满足schema，not验证通过
			return true, nil
		}
	}

	// 如果没有任何验证关键字，返回错误
	if !hasValidation {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "not schema must contain validation keywords",
			Value:   schemaValue,
			Tag:     "not",
		}
	}

	// 如果所有验证都通过，not验证失败
	return false, &errors.ValidationError{
		Path:    path,
		Message: "value must not validate against the schema in not",
		Value:   value,
		Tag:     "not",
	}
}
