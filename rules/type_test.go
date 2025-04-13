package rules

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateType(t *testing.T) {
	registry := NewRegistry()
	registerTypeRules(registry)
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
			name:        "Valid string",
			value:       "hello",
			schemaValue: "string",
			path:        "root",
			expectValid: true,
		},
		{
			name:        "Invalid string",
			value:       42,
			schemaValue: "string",
			path:        "root",
			expectValid: false,
			expectErr:   "value is of type int, expected string",
		},
		{
			name:        "Valid multi-type",
			value:       42,
			schemaValue: []interface{}{"string", "number"},
			path:        "root",
			expectValid: true,
		},
		{
			name:        "Invalid multi-type",
			value:       true,
			schemaValue: []interface{}{"string", "number"},
			path:        "root",
			expectValid: false,
			expectErr:   "value type does not match any of the expected types: string, number",
		},
		{
			name:        "Null value",
			value:       nil,
			schemaValue: "null",
			path:        "root",
			expectValid: true,
		},
		{
			name:        "Invalid schema type",
			value:       "hello",
			schemaValue: 123,
			path:        "root",
			expectValid: false,
			expectErr:   "schema type must be a string or an array of strings",
		},
		{
			name:        "Invalid multi-type schema",
			value:       "hello",
			schemaValue: []interface{}{123, "string"},
			path:        "root",
			expectValid: true, // Should match string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateType(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestCheckType(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		typeName string
		expected bool
	}{
		{"String", "hello", "string", true},
		{"Not string", 42, "string", false},
		{"Number", 42.5, "number", true},
		{"Integer", 42, "integer", true},
		{"Float as integer", 42.5, "integer", false},
		{"Boolean", true, "boolean", true},
		{"Object", map[string]interface{}{"a": 1}, "object", true},
		{"Array", []interface{}{1, 2}, "array", true},
		{"Null", nil, "null", true},
		{"Not null", "something", "null", false},
		{"JSON number", json.Number("42"), "number", true},
		{"Invalid JSON number", json.Number("invalid"), "number", false},
		{"Unknown type", "hello", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkType(tt.value, tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
