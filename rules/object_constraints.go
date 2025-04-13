package rules

import (
	"context"
	"fmt"

	"github.com/songzhibin97/jsonschema-validator/errors"
)

// validateMinProperties 验证对象最小属性数量
func validateMinProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema中的最小属性数量
	minProperties, ok := toInt(schemaValue)
	if !ok || minProperties < 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "minProperties must be a non-negative integer",
			Value:   schemaValue,
			Tag:     "minProperties",
		}
	}

	// 获取对象
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "minProperties can only be applied to objects",
			Value:   value,
			Tag:     "minProperties",
		}
	}

	if len(obj) < minProperties {
		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("object has %d properties, which is less than minProperties %d", len(obj), minProperties),
			Value:   value,
			Tag:     "minProperties",
			Param:   fmt.Sprintf("%d", minProperties),
		}
	}

	return true, nil
}

// validateMaxProperties 验证对象最大属性数量
func validateMaxProperties(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	// 获取schema中的最大属性数量
	maxProperties, ok := toInt(schemaValue)
	if !ok || maxProperties < 0 {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "maxProperties must be a non-negative integer",
			Value:   schemaValue,
			Tag:     "maxProperties",
		}
	}

	// 获取对象
	obj, ok := value.(map[string]interface{})
	if !ok {
		return false, &errors.ValidationError{
			Path:    path,
			Message: "maxProperties can only be applied to objects",
			Value:   value,
			Tag:     "maxProperties",
		}
	}

	if len(obj) > maxProperties {
		return false, &errors.ValidationError{
			Path:    path,
			Message: fmt.Sprintf("object has %d properties, which is more than maxProperties %d", len(obj), maxProperties),
			Value:   value,
			Tag:     "maxProperties",
			Param:   fmt.Sprintf("%d", maxProperties),
		}
	}

	return true, nil
}
