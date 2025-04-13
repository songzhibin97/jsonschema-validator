package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePatternProperties(t *testing.T) {
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
			name: "Valid pattern match",
			value: map[string]interface{}{
				"abc123": "test",
			},
			schemaValue: map[string]interface{}{
				"^[a-z]+[0-9]+$": map[string]interface{}{
					"type": "string",
				},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Valid no matching properties",
			value: map[string]interface{}{
				"xyz": "test",
			},
			schemaValue: map[string]interface{}{
				"^[a-z]+[0-9]+$": map[string]interface{}{
					"type": "string",
				},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Invalid pattern match type",
			value: map[string]interface{}{
				"abc123": 123,
			},
			schemaValue: map[string]interface{}{
				"^[a-z]+[0-9]+$": map[string]interface{}{
					"type": "string",
				},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "property validation failed for keyword 'type'",
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: map[string]interface{}{"^[a-z]+$": map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: false,
			expectErr:   "patternProperties can only be applied to objects",
		},
		{
			name:        "Invalid schema not an object",
			value:       map[string]interface{}{"abc": "test"},
			schemaValue: "not an object",
			path:        "root",
			expectValid: false,
			expectErr:   "patternProperties must be an object",
		},
		{
			name: "Invalid pattern",
			value: map[string]interface{}{
				"abc": "test",
			},
			schemaValue: map[string]interface{}{
				"[": map[string]interface{}{"type": "string"}, // 无效正则表达式
			},
			path:        "root",
			expectValid: false,
			expectErr:   "invalid pattern: error parsing regexp",
		},
		{
			name: "Invalid no validator",
			value: map[string]interface{}{
				"abc123": "test",
			},
			schemaValue: map[string]interface{}{
				"^[a-z]+[0-9]+$": map[string]interface{}{
					"type": "string",
				},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "validator not found in context",
			ctx:         context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.ctx != nil {
				testCtx = tt.ctx
			}
			valid, err := validatePatternProperties(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid)
			if tt.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
			}
		})
	}
}

func TestValidateAdditionalProperties(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterValidator("type", mockTypeValidator)
	baseCtx := context.WithValue(context.Background(), "validator", registry)
	ctxWithProps := context.WithValue(baseCtx, "properties", map[string]interface{}{
		"name": map[string]interface{}{"type": "string"},
	})
	ctxWithPatterns := context.WithValue(ctxWithProps, "patternProperties", map[string]interface{}{
		"^[a-z]+[0-9]+$": map[string]interface{}{"type": "string"},
	})

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
			name: "Valid allowed additional",
			value: map[string]interface{}{
				"name":  "John",
				"extra": "value",
			},
			schemaValue: true,
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxWithProps,
		},
		{
			name: "Valid no additional",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: false,
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxWithProps,
		},
		{
			name: "Valid pattern match no additional",
			value: map[string]interface{}{
				"name":   "John",
				"abc123": "test",
			},
			schemaValue: false,
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxWithPatterns,
		},
		{
			name: "Invalid not allowed additional",
			value: map[string]interface{}{
				"name":  "John",
				"extra": "value",
			},
			schemaValue: false,
			path:        "root",
			expectValid: false,
			expectErr:   "additional properties are not allowed",
			ctx:         ctxWithProps,
		},
		{
			name: "Valid additional with schema",
			value: map[string]interface{}{
				"name":  "John",
				"extra": "value",
			},
			schemaValue: map[string]interface{}{
				"type": "string",
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxWithProps,
		},
		{
			name: "Invalid additional schema failure",
			value: map[string]interface{}{
				"name":  "John",
				"extra": 123,
			},
			schemaValue: map[string]interface{}{
				"type": "string",
			},
			path:        "root",
			expectValid: false,
			expectErr:   "additional property validation failed for keyword 'type'",
			ctx:         ctxWithProps,
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: true,
			path:        "root",
			expectValid: false,
			expectErr:   "additionalProperties can only be applied to objects",
			ctx:         ctxWithProps,
		},
		{
			name: "Invalid schema type",
			value: map[string]interface{}{
				"name":  "John",
				"extra": "value",
			},
			schemaValue: "not a bool or object",
			path:        "root",
			expectValid: false,
			expectErr:   "additionalProperties must be a boolean or an object",
			ctx:         ctxWithProps,
		},
		{
			name: "Invalid no validator",
			value: map[string]interface{}{
				"name":  "John",
				"extra": "value",
			},
			schemaValue: map[string]interface{}{
				"type": "string",
			},
			path:        "root",
			expectValid: false,
			expectErr:   "validator not found in context",
			ctx:         context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := tt.ctx
			if testCtx == nil {
				testCtx = ctxWithProps // 默认上下文
			}
			valid, err := validateAdditionalProperties(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid)
			if tt.expectErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr)
				}
			}
		})
	}
}
