package errors

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected string
	}{
		{
			name:     "Basic error",
			err:      ValidationError{Path: "root.field", Message: "invalid value"},
			expected: "validation error: invalid value (path: root.field)",
		},
		{
			name:     "With tag and param",
			err:      ValidationError{Path: "root", Message: "too small", Tag: "minimum", Param: "10"},
			expected: "validation error: too small (path: root)",
		},
		{
			name:     "Empty path",
			err:      ValidationError{Message: "missing field"},
			expected: "validation error: missing field (path: )",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errs     ValidationErrors
		expected string
	}{
		{
			name:     "Empty errors",
			errs:     ValidationErrors{},
			expected: "",
		},
		{
			name: "Single error",
			errs: ValidationErrors{
				{Path: "field1", Message: "too short"},
			},
			expected: "validation failed with the following errors:\n[1] validation error: too short (path: field1)\n",
		},
		{
			name: "Multiple errors",
			errs: ValidationErrors{
				{Path: "field1", Message: "too short"},
				{Path: "field2", Message: "invalid format"},
			},
			expected: "validation failed with the following errors:\n[1] validation error: too short (path: field1)\n[2] validation error: invalid format (path: field2)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.errs.Error())
		})
	}
}

func TestValidationErrors_FormatWithMode(t *testing.T) {
	errs := ValidationErrors{
		{Path: "field1", Message: "too short", Tag: "minLength", Param: "5"},
		{Path: "field2", Message: "invalid format", Tag: "email"},
	}

	tests := []struct {
		name     string
		mode     FormattingMode
		expected string
	}{
		{
			name:     "Simple mode",
			mode:     FormattingModeSimple,
			expected: "too short; invalid format",
		},
		{
			name:     "Detailed mode",
			mode:     FormattingModeDetailed,
			expected: "validation failed with the following errors:\n[1] validation error: too short (path: field1)\n[2] validation error: invalid format (path: field2)\n",
		},
		{
			name:     "JSON mode",
			mode:     FormattingModeJSON,
			expected: `[{"path":"field1","message":"too short","tag":"minLength","param":"5"},{"path":"field2","message":"invalid format","tag":"email"}]`,
		},
		{
			name:     "Unknown mode",
			mode:     FormattingMode(999),
			expected: "validation failed with the following errors:\n[1] validation error: too short (path: field1)\n[2] validation error: invalid format (path: field2)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errs.FormatWithMode(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test empty errors
	t.Run("Empty JSON", func(t *testing.T) {
		empty := ValidationErrors{}
		assert.Equal(t, "[]", empty.FormatWithMode(FormattingModeJSON))
	})
}

func TestValidationErrorMap_Error(t *testing.T) {
	tests := []struct {
		name     string
		errMap   ValidationErrorMap
		expected string
	}{
		{
			name:     "Empty map",
			errMap:   ValidationErrorMap{},
			expected: "",
		},
		{
			name: "Single field",
			errMap: ValidationErrorMap{
				"name": {{Path: "name", Message: "required field missing"}},
			},
			expected: "validation failed for the following fields:\nField 'name':\n  [1] required field missing\n",
		},
		{
			name: "Multiple fields",
			errMap: ValidationErrorMap{
				"name": {{Path: "name", Message: "too short"}},
				"age":  {{Path: "age", Message: "must be positive"}},
			},
			expected: "validation failed for the following fields:\nField 'name':\n  [1] too short\nField 'age':\n  [1] must be positive\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errMap.Error()
			// Normalize line endings for cross-platform compatibility
			result = strings.ReplaceAll(result, "\r\n", "\n")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNew(t *testing.T) {
	err := New("test error")
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}
