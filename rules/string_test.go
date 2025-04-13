package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMinLength(t *testing.T) {
	registry := NewRegistry()
	registerStringRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid above min", "hello", 3, "root", true, ""},
		{"Valid equal min", "abc", 3, "root", true, ""},
		{"Invalid below min", "ab", 3, "root", false, "length less than minimum"},
		{"Invalid type", 123, 3, "root", false, "must be a string"},
		{"Invalid schema type", "hello", "not a number", "root", false, "minLength must be a non-negative integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMinLength(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateMaxLength(t *testing.T) {
	registry := NewRegistry()
	registerStringRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid below max", "hi", 3, "root", true, ""},
		{"Valid equal max", "abc", 3, "root", true, ""},
		{"Invalid above max", "abcd", 3, "root", false, "length greater than maximum"},
		{"Invalid type", 123, 3, "root", false, "must be a string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateMaxLength(ctx, tt.value, tt.schemaValue, tt.path)
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

func TestValidatePattern(t *testing.T) {
	registry := NewRegistry()
	registerStringRules(registry)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		expectValid bool
		expectErr   string
	}{
		{"Valid match", "abc123", "^[a-z]+[0-9]+$", "root", true, ""},
		{"Invalid no match", "123abc", "^[a-z]+[0-9]+$", "root", false, "does not match pattern"},
		{"Invalid type", 123, "^[a-z]+$", "root", false, "must be a string"},
		{"Invalid pattern", "abc", "[", "root", false, "invalid pattern"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validatePattern(ctx, tt.value, tt.schemaValue, tt.path)
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
