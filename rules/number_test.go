package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMinimum(t *testing.T) {
	registry := NewRegistry()
	registerNumberRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid above minimum", 10, 5, "root", true, ""},
		{"Valid equal minimum", 5, 5, "root", true, ""},
		{"Invalid below minimum", 3, 5, "root", false, "less than minimum"},
		{"Float valid", 5.5, 5.0, "root", true, ""},
		{"Float invalid", 4.9, 5.0, "root", false, "less than minimum"},
		{"Invalid type", "not a number", 5, "root", false, "must be a number"},
		{"Invalid schema type", 10, "not a number", "root", false, "minimum must be a number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMinimum(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateMaximum(t *testing.T) {
	registry := NewRegistry()
	registerNumberRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid below maximum", 5, 10, "root", true, ""},
		{"Valid equal maximum", 10, 10, "root", true, ""},
		{"Invalid above maximum", 15, 10, "root", false, "greater than maximum"},
		{"Float valid", 9.9, 10.0, "root", true, ""},
		{"Float invalid", 10.1, 10.0, "root", false, "greater than maximum"},
		{"Invalid type", "not a number", 10, "root", false, "must be a number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMaximum(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateExclusiveMinimum(t *testing.T) {
	registry := NewRegistry()
	registerNumberRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid above exclusive minimum", 6, 5, "root", true, ""},
		{"Invalid equal exclusive minimum", 5, 5, "root", false, "less than or equal to exclusive minimum"},
		{"Invalid below exclusive minimum", 4, 5, "root", false, "less than or equal to exclusive minimum"},
		{"Float valid", 5.1, 5.0, "root", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateExclusiveMinimum(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateExclusiveMaximum(t *testing.T) {
	registry := NewRegistry()
	registerNumberRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid below exclusive maximum", 4, 5, "root", true, ""},
		{"Invalid equal exclusive maximum", 5, 5, "root", false, "greater than or equal to exclusive maximum"},
		{"Invalid above exclusive maximum", 6, 5, "root", false, "greater than or equal to exclusive maximum"},
		{"Float valid", 4.9, 5.0, "root", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateExclusiveMaximum(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateMultipleOf(t *testing.T) {
	registry := NewRegistry()
	registerNumberRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid multiple", 10, 2, "root", true, ""},
		{"Invalid not multiple", 7, 2, "root", false, "not a multiple of"},
		{"Float valid", 1.5, 0.5, "root", true, ""},
		{"Float invalid", 1.6, 0.5, "root", false, "not a multiple of"},
		{"Zero divisor", 10, 0, "root", false, "multipleOf must be a positive number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMultipleOf(ctx, tt.value, tt.schemaValue, tt.path)
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
