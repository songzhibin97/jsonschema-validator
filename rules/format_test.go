package rules

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFormat(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterValidator("format", validateFormat)
	ctxStrict := context.WithValue(context.Background(), "validator", registry)
	ctxStrict = context.WithValue(ctxStrict, "validationMode", 0) // ModeStrict
	ctxLoose := context.WithValue(context.Background(), "validator", registry)
	ctxLoose = context.WithValue(ctxLoose, "validationMode", 1) // Non-strict

	tests := []struct {
		name        string
		value       interface{}
		schemaValue interface{}
		path        string
		ctx         context.Context
		expectValid bool
		expectErr   string
	}{
		{
			name:        "Valid email",
			value:       "test@example.com",
			schemaValue: "email",
			path:        "root",
			ctx:         ctxStrict,
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid email",
			value:       "invalid-email",
			schemaValue: "email",
			path:        "root",
			ctx:         ctxStrict,
			expectValid: false,
			expectErr:   "invalid email format",
		},
		{
			name:        "Valid uuid",
			value:       "123e4567-e89b-12d3-a456-426614174000",
			schemaValue: "uuid",
			path:        "root",
			ctx:         ctxStrict,
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Invalid uuid",
			value:       "invalid-uuid",
			schemaValue: "uuid",
			path:        "root",
			ctx:         ctxStrict,
			expectValid: false,
			expectErr:   "invalid uuid format",
		},
		{
			name:        "Unknown format strict",
			value:       "test",
			schemaValue: "unknown",
			path:        "root",
			ctx:         ctxStrict,
			expectValid: false,
			expectErr:   "unknown format: unknown",
		},
		{
			name:        "Unknown format loose",
			value:       "test",
			schemaValue: "unknown",
			path:        "root",
			ctx:         ctxLoose,
			expectValid: true,
			expectErr:   "",
		},
		{
			name:        "Non-string value",
			value:       123,
			schemaValue: "email",
			path:        "root",
			ctx:         ctxStrict,
			expectValid: false,
			expectErr:   "value must be a string", // 更新为匹配代码
		},
		{
			name:        "Non-string schema",
			value:       "test@example.com",
			schemaValue: 123,
			path:        "root",
			ctx:         ctxStrict,
			expectValid: false,
			expectErr:   "format must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := validateFormat(tt.ctx, tt.value, tt.schemaValue, tt.path)
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

func TestRegisterFormatValidator(t *testing.T) {
	// 备份原始验证器映射
	originalMap := make(map[string]func(string) bool)
	for k, v := range formatValidatorMap {
		originalMap[k] = v
	}
	defer func() {
		formatValidatorMap = originalMap
	}()

	tests := []struct {
		name        string
		format      string
		validator   func(string) bool
		input       string
		expectValid bool
	}{
		{
			name:   "Register valid custom format",
			format: "custom",
			validator: func(s string) bool {
				return s == "valid"
			},
			input:       "valid",
			expectValid: true,
		},
		{
			name:   "Register invalid custom format",
			format: "custom",
			validator: func(s string) bool {
				return s == "valid"
			},
			input:       "invalid",
			expectValid: false,
		},
		{
			name:        "Register nil validator",
			format:      "custom",
			validator:   nil,
			input:       "valid",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清空映射以隔离测试
			formatValidatorMap = make(map[string]func(string) bool)
			registry := NewRegistry()
			registry.RegisterValidator("format", validateFormat)
			ctx := context.WithValue(context.Background(), "validator", registry)

			// 注册验证器
			RegisterFormatValidator(tt.format, tt.validator)

			// 验证
			valid, err := validateFormat(ctx, tt.input, tt.format, "root")
			if tt.validator == nil && tt.format != "" {
				// 如果注册了nil验证器，预期格式不存在
				assert.False(t, valid, "valid mismatch for %s", tt.name)
				assert.Error(t, err, "expected error for %s", tt.name)
				assert.Contains(t, err.Error(), "unknown format", "error message mismatch for %s", tt.name)
			} else {
				assert.Equal(t, tt.expectValid, valid, "valid mismatch for %s", tt.name)
				if tt.expectValid {
					assert.NoError(t, err, "unexpected error for %s", tt.name)
				} else {
					assert.Error(t, err, "expected error for %s", tt.name)
					assert.Contains(t, err.Error(), fmt.Sprintf("invalid %s format", tt.format), "error message mismatch for %s", tt.name)
				}
			}
		})
	}
}
