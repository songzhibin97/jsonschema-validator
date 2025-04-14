package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMessageFormats(t *testing.T) {
	tests := []struct {
		name           string
		error          ValidationError
		simpleFormat   string
		detailedFormat string
		jsonFormat     string
	}{
		{
			name: "Basic validation error",
			error: ValidationError{
				Path:    "user.name",
				Message: "must be at least 3 characters",
				Tag:     "minLength",
				Value:   "Jo",
				Param:   "3",
			},
			simpleFormat:   "must be at least 3 characters",
			detailedFormat: "validation error: must be at least 3 characters (path: user.name)",
			jsonFormat:     `{"path":"user.name","message":"must be at least 3 characters","value":"Jo","tag":"minLength","param":"3"}`,
		},
		{
			name: "Type validation error",
			error: ValidationError{
				Path:    "user.age",
				Message: "expected integer, got string",
				Tag:     "type",
				Value:   "thirty",
			},
			simpleFormat:   "expected integer, got string",
			detailedFormat: "validation error: expected integer, got string (path: user.age)",
			jsonFormat:     `{"path":"user.age","message":"expected integer, got string","value":"thirty","tag":"type"}`,
		},
		{
			name: "Required field error",
			error: ValidationError{
				Path:    "user.email",
				Message: "required property is missing",
				Tag:     "required",
			},
			simpleFormat:   "required property is missing",
			detailedFormat: "validation error: required property is missing (path: user.email)",
			jsonFormat:     `{"path":"user.email","message":"required property is missing","tag":"required"}`,
		},
		{
			name: "Pattern validation error",
			error: ValidationError{
				Path:    "user.email",
				Message: "does not match pattern ^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$",
				Tag:     "pattern",
				Value:   "invalid-email",
				Param:   "^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$",
			},
			simpleFormat:   "does not match pattern ^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$",
			detailedFormat: "validation error: does not match pattern ^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$ (path: user.email)",
			jsonFormat:     `{"path":"user.email","message":"does not match pattern ^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$","value":"invalid-email","tag":"pattern","param":"^[a-z0-9._%+-]+@[a-z0-9.-]+\\.[a-z]{2,}$"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test individual error
			assert.Equal(t, tt.detailedFormat, tt.error.Error(), "Error() method should return detailed format")

			// Test error collection
			errs := ValidationErrors{tt.error}

			// Test simple format
			simple := errs.FormatWithMode(FormattingModeSimple)
			assert.Equal(t, tt.simpleFormat, simple, "Simple format mismatch")

			// Test detailed format
			detailed := errs.FormatWithMode(FormattingModeDetailed)
			assert.Contains(t, detailed, tt.detailedFormat, "Detailed format should contain error details")

			// Test JSON format
			jsonOutput := errs.FormatWithMode(FormattingModeJSON)
			assert.Contains(t, jsonOutput, tt.jsonFormat, "JSON format should contain error details")

			// Test invalid mode falls back to detailed
			fallback := errs.FormatWithMode(FormattingMode(999))
			assert.Equal(t, detailed, fallback, "Invalid mode should fall back to detailed format")
		})
	}
}

func TestValidationErrorMap(t *testing.T) {
	// Create a map of errors
	errorMap := ValidationErrorMap{
		"name": ValidationErrors{
			{Path: "name", Message: "too short", Tag: "minLength"},
			{Path: "name", Message: "contains invalid characters", Tag: "pattern"},
		},
		"age": ValidationErrors{
			{Path: "age", Message: "must be positive", Tag: "minimum"},
		},
	}

	// Test error message
	errMsg := errorMap.Error()
	assert.Contains(t, errMsg, "validation failed for the following fields")
	assert.Contains(t, errMsg, "Field 'name'")
	assert.Contains(t, errMsg, "too short")
	assert.Contains(t, errMsg, "contains invalid characters")
	assert.Contains(t, errMsg, "Field 'age'")
	assert.Contains(t, errMsg, "must be positive")

	// Test empty map
	emptyMap := ValidationErrorMap{}
	assert.Equal(t, "", emptyMap.Error(), "Empty map should return empty string")
}

func TestNestedErrors(t *testing.T) {
	// Create deeply nested errors
	nested := ValidationErrors{
		{
			Path:    "user.address.street",
			Message: "street name too long",
			Tag:     "maxLength",
			Value:   "1234 Very Long Street Name That Exceeds The Maximum Length",
			Param:   "30",
		},
		{
			Path:    "user.contacts[0].phone",
			Message: "invalid phone format",
			Tag:     "pattern",
			Value:   "not-a-phone",
			Param:   "^\\d{3}-\\d{3}-\\d{4}$",
		},
	}

	// Test formatted output
	jsonOutput := nested.FormatWithMode(FormattingModeJSON)

	// Verify full paths are preserved
	assert.Contains(t, jsonOutput, "user.address.street")
	assert.Contains(t, jsonOutput, "user.contacts[0].phone")

	// Verify detailed error format handles nested structures
	detailed := nested.FormatWithMode(FormattingModeDetailed)
	assert.Contains(t, detailed, "user.address.street")
	assert.Contains(t, detailed, "user.contacts[0].phone")
}
