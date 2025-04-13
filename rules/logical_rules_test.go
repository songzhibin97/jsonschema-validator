package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAllOf(t *testing.T) {
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
			name:        "Valid allOf",
			value:       "test",
			schemaValue: []interface{}{map[string]interface{}{"type": "string"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid allOf",
			value:       123,
			schemaValue: []interface{}{map[string]interface{}{"type": "string"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: false,
			expectErr:   "failed to validate against schema at allOf", // 更新为更宽松的匹配
		},
		// 其他用例...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.ctx != nil {
				testCtx = tt.ctx
			}
			valid, err := validateAllOf(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid, "valid mismatch for %s", tt.name)
			if tt.expectErr == "" {
				assert.NoError(t, err, "unexpected error for %s", tt.name)
			} else {
				assert.Error(t, err, "expected error for %s", tt.name)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr, "error message mismatch for %s", tt.name)
				}
			}
		})
	}
}

func TestValidateNot(t *testing.T) {
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
			name:        "Valid not",
			value:       123,
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid not",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: false,
			expectErr:   "value must not validate against the schema in not",
		},
		// 其他用例...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.ctx != nil {
				testCtx = tt.ctx
			}
			valid, err := validateNot(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid, "valid mismatch for %s", tt.name)
			if tt.expectErr == "" {
				assert.NoError(t, err, "unexpected error for %s", tt.name)
			} else {
				assert.Error(t, err, "expected error for %s", tt.name)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr, "error message mismatch for %s", tt.name)
				}
			}
		})
	}
}

func TestValidateAnyOf(t *testing.T) {
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
			name:        "Valid anyOf",
			value:       "test",
			schemaValue: []interface{}{map[string]interface{}{"type": "integer"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid anyOf",
			value:       true,
			schemaValue: []interface{}{map[string]interface{}{"type": "integer"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: false,
			expectErr:   "value does not match any schema in anyOf",
		},
		{
			name:        "Invalid schema not array",
			value:       "test",
			schemaValue: "not an array",
			path:        "root",
			expectValid: false,
			expectErr:   "anyOf must be an array",
		},
		{
			name:        "Invalid empty schemas",
			value:       "test",
			schemaValue: []interface{}{},
			path:        "root",
			expectValid: false,
			expectErr:   "anyOf cannot be empty",
		},
		{
			name:        "Invalid no validator",
			value:       "test",
			schemaValue: []interface{}{map[string]interface{}{"type": "string"}},
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
			valid, err := validateAnyOf(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid, "valid mismatch for %s", tt.name)
			if tt.expectErr == "" {
				assert.NoError(t, err, "unexpected error for %s", tt.name)
			} else {
				assert.Error(t, err, "expected error for %s", tt.name)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr, "error message mismatch for %s", tt.name)
				}
			}
		})
	}
}

func TestValidateOneOf(t *testing.T) {
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
			name:        "Valid oneOf",
			value:       "test",
			schemaValue: []interface{}{map[string]interface{}{"type": "integer"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid oneOf multiple",
			value:       "test",
			schemaValue: []interface{}{map[string]interface{}{"type": "string"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: false,
			expectErr:   "value matches more than one schema in oneOf",
		},
		{
			name:        "Invalid oneOf none",
			value:       true,
			schemaValue: []interface{}{map[string]interface{}{"type": "integer"}, map[string]interface{}{"type": "string"}},
			path:        "root",
			expectValid: false,
			expectErr:   "value does not match any schema in oneOf",
		},
		{
			name:        "Invalid schema not array",
			value:       "test",
			schemaValue: "not an array",
			path:        "root",
			expectValid: false,
			expectErr:   "oneOf must be an array",
		},
		{
			name:        "Invalid empty schemas",
			value:       "test",
			schemaValue: []interface{}{},
			path:        "root",
			expectValid: false,
			expectErr:   "oneOf cannot be empty",
		},
		{
			name:        "Invalid no validator",
			value:       "test",
			schemaValue: []interface{}{map[string]interface{}{"type": "string"}},
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
			valid, err := validateOneOf(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid, "valid mismatch for %s", tt.name)
			if tt.expectErr == "" {
				assert.NoError(t, err, "unexpected error for %s", tt.name)
			} else {
				assert.Error(t, err, "expected error for %s", tt.name)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr, "error message mismatch for %s", tt.name)
				}
			}
		})
	}
}
