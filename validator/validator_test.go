package validator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/songzhibin97/jsonschema-validator/errors"
	"github.com/songzhibin97/jsonschema-validator/schema"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	v := New(
		WithTagName("custom"),
		WithValidationMode(schema.ModeLoose),
		WithErrorFormattingMode(errors.FormattingModeSimple),
		WithCaching(true),
		WithStopOnFirstError(true),
		WithRecursiveValidation(true),
		WithAllowUnknownFields(true),
	)
	assert.Equal(t, "custom", v.opts.TagName)
	assert.Equal(t, schema.ModeLoose, v.opts.ValidationMode)
	assert.Equal(t, errors.FormattingModeSimple, v.opts.ErrorFormattingMode)
	assert.True(t, v.opts.EnableCaching)
	assert.True(t, v.opts.StopOnFirstError)
	assert.True(t, v.opts.RecursiveValidation)
	assert.True(t, v.opts.AllowUnknownFields)
}

func TestValidateJSON(t *testing.T) {
	v := New(WithValidationMode(schema.ModeStrict))
	tests := []struct {
		name        string
		jsonData    string
		schemaJSON  string
		expectValid bool
		expectErr   bool
		errorCount  int
		errMsg      string
	}{
		{
			name:        "Valid object",
			jsonData:    `{"name":"John","age":30}`,
			schemaJSON:  `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer","minimum":18}},"required":["name"]}`,
			expectValid: true,
			errorCount:  0,
		},
		{
			name:        "Invalid type",
			jsonData:    `{"name":123,"age":30}`,
			schemaJSON:  `{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer","minimum":18}},"required":["name"]}`,
			expectValid: false,
			errorCount:  1,
			errMsg:      "expected string",
		},
		{
			name:        "Nested object",
			jsonData:    `{"user":{"name":"John"}}`,
			schemaJSON:  `{"type":"object","properties":{"user":{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer","minimum":18}},"required":["name"]}}}`,
			expectValid: true,
			errorCount:  0,
		},
		{
			name:        "Array items",
			jsonData:    `["apple","banana"]`,
			schemaJSON:  `{"type":"array","items":{"type":"string"}}`,
			expectValid: true,
			errorCount:  0,
		},
		{
			name:        "Unknown field strict",
			jsonData:    `{"name":"John","extra":true}`,
			schemaJSON:  `{"type":"object","properties":{"name":{"type":"string"}},"additionalProperties":false}`,
			expectValid: false,
			errorCount:  1,
			errMsg:      "unknown field",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.ValidateJSON(tt.jsonData, tt.schemaJSON)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Len(t, result.Errors, tt.errorCount)
			if tt.errMsg != "" && len(result.Errors) > 0 {
				assert.Contains(t, result.Errors[0].Message, tt.errMsg)
			}
		})
	}
}

func TestVar(t *testing.T) {
	v := New()
	tests := []struct {
		name      string
		value     interface{}
		tag       string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "Invalid enum",
			value:     "invalid",
			tag:       "enum=val1|val2",
			expectErr: true,
			errMsg:    "value must be one of: val1, val2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.value, tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWithSchema(t *testing.T) {
	v := New(WithValidationMode(schema.ModeLoose))

	schemaMap := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
			"age":  map[string]interface{}{"type": "integer", "minimum": 18},
		},
		"required": []interface{}{"name"},
	}

	tests := []struct {
		name        string
		value       interface{}
		path        string
		expectValid bool
		errorCount  int
		errMsg      string
	}{
		{
			name:        "Valid",
			value:       map[string]interface{}{"name": "John", "age": 30},
			path:        "root",
			expectValid: true,
		},
		{
			name:        "Missing required",
			value:       map[string]interface{}{"age": 30},
			path:        "root",
			expectValid: false,
			errorCount:  1,
			errMsg:      "required",
		},
		{
			name:        "Invalid type",
			value:       map[string]interface{}{"name": 123, "age": 30},
			path:        "root",
			expectValid: false,
			errorCount:  1,
			errMsg:      "expected string",
		},
		{
			name:        "Nested object",
			value:       map[string]interface{}{"name": "John"},
			path:        "root",
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.ValidateWithSchema(tt.value, schemaMap, tt.path)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Len(t, result.Errors, tt.errorCount)
			if tt.errMsg != "" && len(result.Errors) > 0 {
				assert.Contains(t, result.Errors[0].Message, tt.errMsg)
			}
		})
	}
}

func TestStruct(t *testing.T) {
	v := New(WithTagName("validate"), WithRecursiveValidation(true))

	type NestedStruct struct {
		Score int `validate:"minimum=0"`
	}

	type TestStruct struct {
		Name   string       `validate:"required,type=string"`
		Age    int          `validate:"minimum=18"`
		Nested NestedStruct `validate:"required"`
	}

	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
		errMsg    string
	}{
		{
			name:  "Valid struct",
			input: TestStruct{Name: "John", Age: 30, Nested: NestedStruct{Score: 10}},
		},
		{
			name:      "Missing required",
			input:     TestStruct{Age: 30},
			expectErr: true,
			errMsg:    "field is required",
		},
		{
			name:      "Invalid minimum",
			input:     TestStruct{Name: "John", Age: 15, Nested: NestedStruct{Score: 10}},
			expectErr: true,
			errMsg:    "less than minimum",
		},
		{
			name:      "Invalid nested",
			input:     TestStruct{Name: "John", Age: 30, Nested: NestedStruct{Score: -1}},
			expectErr: true,
			errMsg:    "less than minimum",
		},
		{
			name:      "Invalid input",
			input:     "not a struct",
			expectErr: true,
			errMsg:    "input must be a struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompileSchema(t *testing.T) {
	v := New(WithCaching(true))

	schemaJSON := `{"type":"object","properties":{"name":{"type":"string"}}}`

	s, err := v.CompileSchema(schemaJSON)
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.NotNil(t, s.Compiled)
	assert.Equal(t, "object", s.Raw["type"])

	// 验证缓存
	s2, err := v.CompileSchema(schemaJSON)
	assert.NoError(t, err)
	assert.Same(t, s, s2)

	// 清理缓存
	v.ClearCache()
	_, err = v.CompileSchema(schemaJSON)
	assert.NoError(t, err)

	// 无效 schema
	_, err = v.CompileSchema(`{`)
	assert.Error(t, err)
}

func TestCustomValidation(t *testing.T) {
	v := New()
	v.SetCustomValidateFunc(func(ctx context.Context, value interface{}, path string) (bool, error) {
		if str, ok := value.(string); ok && strings.HasPrefix(strings.ToUpper(str), "ADMIN_") {
			return true, nil
		}
		return false, nil
	})
	type TestStruct struct {
		Role string `validate:"required"`
	}
	err := v.Struct(TestStruct{Role: "admin_user"})
	assert.NoError(t, err)
	err = v.Struct(TestStruct{Role: "user"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "value must start with 'ADMIN_'")
}

func TestConcurrentValidation(t *testing.T) {
	v := New()

	type TestStruct struct {
		Name string `validate:"required,type=string"`
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := v.Struct(TestStruct{Name: fmt.Sprintf("User%d", i)})
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()
}
