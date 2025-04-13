package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateItems(t *testing.T) {
	registry := NewRegistry()
	registerArrayRules(registry)
	registerTypeRules(registry) // items 需要类型验证器
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid items", []interface{}{"a", "b"}, map[string]interface{}{"type": "string"}, "root", true, ""},
		{"Invalid items", []interface{}{"a", 1}, map[string]interface{}{"type": "string"}, "root", false, "expected string"},
		{"Array of schemas", []interface{}{1, 2}, []interface{}{map[string]interface{}{"type": "integer"}, map[string]interface{}{"type": "integer"}}, "root", true, ""},
		{"Invalid type", "not an array", map[string]interface{}{"type": "string"}, "root", false, "items can only be applied to arrays"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateItems(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateMinItems(t *testing.T) {
	registry := NewRegistry()
	registerArrayRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid above min", []interface{}{1, 2, 3}, 2, "root", true, ""},
		{"Valid equal min", []interface{}{1, 2}, 2, "root", true, ""},
		{"Invalid below min", []interface{}{1}, 2, "root", false, "fewer items than minimum"},
		{"Invalid type", "not an array", 2, "root", false, "must be an array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMinItems(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateMaxItems(t *testing.T) {
	registry := NewRegistry()
	registerArrayRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid below max", []interface{}{1}, 2, "root", true, ""},
		{"Valid equal max", []interface{}{1, 2}, 2, "root", true, ""},
		{"Invalid above max", []interface{}{1, 2, 3}, 2, "root", false, "more items than maximum"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMaxItems(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateUniqueItems(t *testing.T) {
	registry := NewRegistry()
	registerArrayRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid unique", []interface{}{1, 2, 3}, true, "root", true, ""},
		{"Invalid duplicates", []interface{}{1, 1, 2}, true, "root", false, "contains duplicate items"},
		{"No check", []interface{}{1, 1}, false, "root", true, ""},
		{"Invalid type", "not an array", true, "root", false, "must be an array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateUniqueItems(ctx, tt.value, tt.schemaValue, tt.path)
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
