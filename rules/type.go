package rules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// 注册类型相关规则
func registerTypeRules(registry ValidatorRegistry) {
	registry.RegisterValidator("type", validateType)
	registry.RegisterValidator("required", requiredValidator)
	registry.RegisterValidator("minimum", minimumValidator)
	registry.RegisterValidator("enum", enumValidator)
}

// validateType 验证值的类型
func validateType(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 处理多类型情况（type: ["string", "number"]）
	if types, ok := schemaValue.([]interface{}); ok {
		for _, t := range types {
			typeStr, ok := t.(string)
			if !ok {
				continue
			}
			if checkType(value, typeStr) {
				return true, nil
			}
		}

		typeNames := make([]string, 0, len(types))
		for _, t := range types {
			if ts, ok := t.(string); ok {
				typeNames = append(typeNames, ts)
			}
		}

		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("value type does not match any of the expected types: %s", strings.Join(typeNames, ", ")),
			Value:   value,
			Tag:     "type",
		}
	}

	// 处理单一类型情况（type: "string"）
	typeStr, ok := schemaValue.(string)
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "schema type must be a string or an array of strings",
			Value:   schemaValue,
			Tag:     "type",
		}
	}

	if !checkType(value, typeStr) {
		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("value is of type %T, expected %s", value, typeStr),
			Value:   value,
			Tag:     "type",
			Param:   typeStr,
		}
	}

	return true, nil
}

// checkType 检查值是否符合指定的类型
func checkType(value interface{}, typeName string) bool {
	if value == nil {
		return typeName == "null"
	}

	switch typeName {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch v := value.(type) {
		case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		case json.Number:
			_, err := v.Float64()
			return err == nil
		}
		return false
	case "integer":
		switch v := value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		case float64:
			return v == float64(int(v))
		case float32:
			return float32(int(v)) == v
		case json.Number:
			f, err := v.Float64()
			if err != nil {
				return false
			}
			return f == float64(int(f))
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}
