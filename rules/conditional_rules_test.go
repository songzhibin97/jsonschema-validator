package rules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateIf(t *testing.T) {
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
		expectCond  bool
		ctx         context.Context
	}{
		{
			name:        "Condition met",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			expectCond:  true,
		},
		{
			name:        "Condition not met",
			value:       123,
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			expectCond:  false,
		},
		{
			name:        "Invalid schema not object",
			value:       "test",
			schemaValue: "not an object",
			path:        "root",
			expectValid: false,
			expectErr:   "if must be an object",
		},
		{
			name:        "Invalid no validator",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: false,
			expectErr:   "validator not found in context",
			ctx:         context.Background(), // 确保无 validator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.ctx != nil {
				testCtx = tt.ctx // 使用指定的上下文（如 context.Background()）
			}
			valid, err := validateIf(testCtx, tt.value, tt.schemaValue, tt.path)
			assert.Equal(t, tt.expectValid, valid, "valid mismatch for %s", tt.name)
			if tt.expectErr == "" {
				assert.NoError(t, err, "unexpected error for %s", tt.name)
			} else {
				assert.Error(t, err, "expected error for %s", tt.name)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectErr, "error message mismatch for %s", tt.name)
				}
			}
			if tt.expectErr == "" {
				condMet, ok := testCtx.Value("ifConditionMet").(bool)
				if ok {
					assert.Equal(t, tt.expectCond, condMet, "condition met mismatch for %s", tt.name)
				}
			}
		})
	}
}

func TestValidateThen(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterValidator("type", mockTypeValidator)
	ctxTrue := context.WithValue(context.Background(), "validator", registry)
	ctxTrue = context.WithValue(ctxTrue, "ifConditionMet", true)
	ctxFalse := context.WithValue(context.Background(), "validator", registry)
	ctxFalse = context.WithValue(ctxFalse, "ifConditionMet", false)

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
			name:        "Valid condition met",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxTrue,
		},
		{
			name:        "Valid condition not met",
			value:       123,
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxFalse,
		},
		{
			name:        "Invalid then failure",
			value:       123,
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: false,
			expectErr:   "validation failed against then schema for keyword 'type'",
			ctx:         ctxTrue,
		},
		{
			name:        "Invalid schema not object",
			value:       "test",
			schemaValue: "not an object",
			path:        "root",
			expectValid: false,
			expectErr:   "then must be an object",
			ctx:         ctxTrue,
		},
		{
			name:        "Invalid no validator",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: false,
			expectErr:   "validator not found in context",
			ctx:         context.Background(), // 确保无 validator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := tt.ctx
			if testCtx == nil {
				testCtx = ctxTrue // 默认使用 ctxTrue
			}
			valid, err := validateThen(testCtx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateElse(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterValidator("type", mockTypeValidator)
	ctxTrue := context.WithValue(context.Background(), "validator", registry)
	ctxTrue = context.WithValue(ctxTrue, "ifConditionMet", true)
	ctxFalse := context.WithValue(context.Background(), "validator", registry)
	ctxFalse = context.WithValue(ctxFalse, "ifConditionMet", false)

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
			name:        "Valid condition not met",
			value:       123,
			schemaValue: map[string]interface{}{"type": "integer"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxFalse,
		},
		{
			name:        "Valid condition met",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "string"},
			path:        "root",
			expectValid: true,
			expectErr:   "",
			ctx:         ctxTrue,
		},
		{
			name:        "Invalid else failure",
			value:       "test",
			schemaValue: map[string]interface{}{"type": "integer"},
			path:        "root",
			expectValid: false,
			expectErr:   "validation failed against else schema for keyword 'type'",
			ctx:         ctxFalse,
		},
		{
			name:        "Invalid schema not object",
			value:       123,
			schemaValue: "not an object",
			path:        "root",
			expectValid: false,
			expectErr:   "else must be an object",
			ctx:         ctxFalse,
		},
		{
			name:        "Invalid no validator",
			value:       123,
			schemaValue: map[string]interface{}{"type": "integer"},
			path:        "root",
			expectValid: false,
			expectErr:   "validator not found in context",
			ctx:         context.Background(), // 确保无 validator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := tt.ctx
			if testCtx == nil {
				testCtx = ctxFalse // 默认使用 ctxFalse
			}
			valid, err := validateElse(testCtx, tt.value, tt.schemaValue, tt.path)
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

func TestValidateConditional(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterValidator("type", mockTypeValidator)
	ctx := context.WithValue(context.Background(), "validator", registry)

	tests := []struct {
		name              string
		value             interface{}
		conditionalSchema map[string]interface{}
		path              string
		expectValid       bool
		expectErr         string
	}{
		{
			name:  "Valid if-then",
			value: "test",
			conditionalSchema: map[string]interface{}{
				"if":   map[string]interface{}{"type": "string"},
				"then": map[string]interface{}{"type": "string"},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:  "Valid if-else",
			value: 123,
			conditionalSchema: map[string]interface{}{
				"if":   map[string]interface{}{"type": "string"},
				"else": map[string]interface{}{"type": "integer"},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
		{
			name:  "Invalid then failure",
			value: "test",
			conditionalSchema: map[string]interface{}{
				"if":   map[string]interface{}{"type": "string"},
				"then": map[string]interface{}{"type": "integer"},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "validation failed against then schema for keyword 'type'",
		},
		{
			name:  "Invalid else failure",
			value: 123,
			conditionalSchema: map[string]interface{}{
				"if":   map[string]interface{}{"type": "string"},
				"else": map[string]interface{}{"type": "string"},
			},
			path:        "root",
			expectValid: false,
			expectErr:   "validation failed against else schema for keyword 'type'",
		},
		{
			name:  "Valid no if",
			value: "test",
			conditionalSchema: map[string]interface{}{
				"then": map[string]interface{}{"type": "string"},
			},
			path:        "root",
			expectValid: true,
			expectErr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateConditional(ctx, tt.value, tt.conditionalSchema, tt.path)
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
