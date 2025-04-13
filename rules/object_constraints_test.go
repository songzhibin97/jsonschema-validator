package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMinProperties(t *testing.T) {
	ctx := context.Background() // 不需要 validator，保持简单

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{
			name: "Valid more than min",
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			schemaValue: 1,
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Valid equal to min",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: 1,
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Invalid less than min",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: 2,
			path:        "root",
			expectValid: false,
			expectErr:   "object has 1 properties, which is less than minProperties 2",
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: 1,
			path:        "root",
			expectValid: false,
			expectErr:   "minProperties can only be applied to objects",
		},
		{
			name:        "Invalid schema not integer",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: "not an integer",
			path:        "root",
			expectValid: false,
			expectErr:   "minProperties must be a non-negative integer",
		},
		{
			name:        "Invalid negative min",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: -1,
			path:        "root",
			expectValid: false,
			expectErr:   "minProperties must be a non-negative integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMinProperties(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateMaxProperties(t *testing.T) {
	ctx := context.Background() // 不需要 validator，保持简单

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{
			name: "Valid less than max",
			value: map[string]interface{}{
				"name": "John",
			},
			schemaValue: 2,
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Valid equal to max",
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			schemaValue: 2,
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name: "Invalid more than max",
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
				"city": "NY",
			},
			schemaValue: 2,
			path:        "root",
			expectValid: false,
			expectErr:   "object has 3 properties, which is more than maxProperties 2",
		},
		{
			name:        "Invalid not an object",
			value:       "not an object",
			schemaValue: 1,
			path:        "root",
			expectValid: false,
			expectErr:   "maxProperties can only be applied to objects",
		},
		{
			name:        "Invalid schema not integer",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: "not an integer",
			path:        "root",
			expectValid: false,
			expectErr:   "maxProperties must be a non-negative integer",
		},
		{
			name:        "Invalid negative max",
			value:       map[string]interface{}{"name": "John"},
			schemaValue: -1,
			path:        "root",
			expectValid: false,
			expectErr:   "maxProperties must be a non-negative integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMaxProperties(ctx, tt.value, tt.schemaValue, tt.path)
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
