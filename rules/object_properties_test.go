package rules

import (
	"context"
	"fmt"
	"testing"

	"github.com/songzhibin97/jsonschema-validator/errors"
	"github.com/stretchr/testify/assert"
)

func mockTypeValidator(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
	expectedType, ok := schemaValue.(string)
	if !ok {
		return false, &errors.ValidationError{Path: path, Message: "type must be a string", Tag: "type"}
	}
	actualType := ""
	switch value.(type) {
	case string:
		actualType = "string"
	case int, int64:
		actualType = "integer"
	case map[string]interface{}:
		actualType = "object"
	default:
		actualType = "unknown"
	}
	if actualType != expectedType {
		return false, &errors.ValidationError{Path: path, Message: fmt.Sprintf("expected type %s, got %s", expectedType, actualType), Tag: "type"}
	}
	return true, nil
}

func TestValidateRequired(t *testing.T) {
	registry := NewRegistry()
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{
			name:        "Valid with all required fields",
			value:       map[string]interface{}{"name": "John", "age": 30},
			schemaValue: []interface{}{"name", "age"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid missing required field",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: []interface{}{"name", "age"},
			path:        "root",
			expectValid: false,
			expectErr:   "required property 'age' is missing",
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: []interface{}{"name"},
			path:        "root",
			expectValid: false,
			expectErr:   "required can only be applied to objects",
		},
		{
			name:        "Invalid schema not an array",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: "not an array",
			path:        "root",
			expectValid: false,
			expectErr:   "required must be an array of strings",
		},
		{
			name:        "Valid with non-string ignored",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: []interface{}{"name", 123}, // 非字符串被跳过
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateRequired(ctx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid)
			if tt.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
			}
		})
	}
}

func TestValidateProperties(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterValidator("type", mockTypeValidator)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
		ctx         context.Context
	}{
		{
			name: "Valid properties",
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type": "integer",
				},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Invalid property type",
			value: map[string]interface{}{
				"name": 123, // 应该是字符串
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "expected type string, got integer",
		},
		{
			name: "Valid with missing optional property",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type": "integer",
				},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: map[string]interface{}{"name": map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: false,
			expectErr:   "properties can only be applied to objects",
		},
		{
			name:        "Invalid schema not an object",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: "not an object",
			path:        "root",
			expectValid: false,
			expectErr:   "properties must be an object",
		},
		{
			name: "Invalid no validator in context",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "validator not found in context",
			// 使用空的 context
			ctx: context.Background(),
		},
		{
			name: "Valid with non-schema keywords",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "User name", // 非验证关键字
				},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.ctx != nil {
				testCtx = tt.ctx
			}
			valid, err := validateProperties(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid)
			if tt.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
			}
		})
	}
}
