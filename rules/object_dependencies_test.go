package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDependencies(t *testing.T) {
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
			name: "Valid property dependency",
			value: map[string]interface{}{
				"credit_card":     "1234",
				"billing_address": "123 Main St",
			},
			schemaValue: map[string]interface{}{
				"credit_card": []interface{}{"billing_address"},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Invalid missing property dependency",
			value: map[string]interface{}{
				"credit_card": "1234",
			},
			schemaValue: map[string]interface{}{
				"credit_card": []interface{}{"billing_address"},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "property 'credit_card' depends on 'billing_address', but it is missing",
		},
		{
			name: "Invalid schema dependency",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type": "object",
				},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "dependency validation failed for property 'name' with keyword 'type'",
		},
		{
			name: "Valid schema dependency",
			value: map[string]interface{}{
				"name": map[string]interface{}{"key": "value"},
			},
			schemaValue: map[string]interface{}{
				"name": map[string]interface{}{
					"type": "object",
				},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Valid property not present",
			value: map[string]interface{}{
				"other": "value",
			},
			schemaValue: map[string]interface{}{
				"credit_card": []interface{}{"billing_address"},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: map[string]interface{}{"name": []interface{}{"age"}},
			path:        "root",
			expectValid: false,
			expectErr:   "dependencies can only be applied to objects",
		},
		{
			name:        "Invalid schema not an object",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: "not an object",
			path:        "root",
			expectValid: false,
			expectErr:   "dependencies must be an object",
		},
		{
			name: "Invalid dependency type",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: map[string]interface{}{
				"name": "invalid type",
			},
			path:        "root",
			expectValid: false,
			expectErr:   "dependency for property 'name' must be an array or an object",
		},
		{
			name: "Valid with non-string in array",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: map[string]interface{}{
				"name": []interface{}{"age", 123},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "property 'name' depends on 'age', but it is missing",
		},
		{
			name: "Invalid no validator",
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
			ctx:         context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.ctx != nil {
				testCtx = tt.ctx
			}
			valid, err := validateDependencies(testCtx, tt.value, tt.schemaValue, tt.path)
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
