package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// validateDependencies 验证对象属性的依赖关系
func validateDependencies(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取依赖映射
	dependencies, ok := schemaValue.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "dependencies must be an object",
			Value:   schemaValue,
			Tag:     "dependencies",
		}
	}

	// 获取对象
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "dependencies can only be applied to objects",
			Value:   value,
			Tag:     "dependencies",
		}
	}

	// 获取validator实例
	registry, ok := ctx.Value("validator").(ValidatorRegistry)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "validator not found in context",
			Tag:     "dependencies",
		}
	}

	// 遍历所有依赖项
	for propName, dependency := range dependencies {
		// 如果对象不包含属性，则跳过该依赖项
		propValue, exists := obj[propName] // 定义 propValue
		if !exists {
			continue
		}

		// 处理两种依赖形式：数组和对象
		switch dep := dependency.(type) {
		case []interface{}:
			// 属性依赖：当属性存在时，依赖的其他属性也必须存在
			for _, depProp := range dep {
				depPropStr, ok := depProp.(string)
				if !ok {
					continue
				}
				if _, exists := obj[depPropStr]; !exists {
					return false, &errors.ValidationError{
						Path:    path,
						Message: fmt.Sprintf("property '%s' depends on '%s', but it is missing", propName, depPropStr),
						Value:   obj,
						Tag:     "dependencies",
						Param:   depPropStr,
					}
				}
			}

		case map[string]interface{}:
			// Schema依赖：当属性存在时，属性值必须验证通过指定的schema
			for keyword, keywordValue := range dep {
				if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
					continue
				}
				validator := registry.GetValidator(keyword)
				if validator == nil {
					continue
				}
				isValid, err := validator(ctx, propValue, keywordValue, path)
				if !isValid || err != nil { // 合并 !isValid 和 err != nil 的处理
					// 如果有原始错误，附加其详细信息
					msg := fmt.Sprintf("dependency validation failed for property '%s' with keyword '%s'", propName, keyword)
					if err != nil {
						msg += fmt.Sprintf(": %v", err)
					}
					return false, &errors.ValidationError{
						Path:    path,
						Message: msg,
						Value:   propValue,
						Tag:     keyword,
					}
				}
			}

		default:
			return false, &errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("dependency for property '%s' must be an array or an object", propName),
				Value:   dependency,
				Tag:     "dependencies",
			}
		}
	}

	return true, nil
}
