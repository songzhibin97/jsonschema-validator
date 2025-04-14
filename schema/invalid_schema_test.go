package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidSchemas(t *testing.T) {
	tests := []struct {
		name        string
		schemaJSON  string
		parseErr    bool
		compileErr  bool
		errContains string
	}{
		{
			name:        "Invalid JSON syntax",
			schemaJSON:  `{"type": "object", "properties": {`,
			parseErr:    true,
			errContains: "failed to parse",
		},
		{
			name:        "Invalid type value",
			schemaJSON:  `{"type": 123}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "invalid type value",
		},
		{
			name:        "Invalid type array",
			schemaJSON:  `{"type": ["string", 123]}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "non-string value",
		},
		{
			name:        "Negative minLength",
			schemaJSON:  `{"type": "string", "minLength": -1}`,
			parseErr:    false,
			compileErr:  false,
			errContains: "",
		},
		{
			name:        "Invalid pattern property",
			schemaJSON:  `{"patternProperties": {"(": {"type": "string"}}}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "invalid pattern",
		},
		{
			name:        "Invalid property schema",
			schemaJSON:  `{"properties": {"name": 123}}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "must be an object",
		},
		{
			name:        "Invalid items schema",
			schemaJSON:  `{"items": 123}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "invalid items value",
		},
		{
			name:        "Invalid required field",
			schemaJSON:  `{"required": [123]}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "must be a string",
		},
		{
			name:        "Unknown keyword in strict mode",
			schemaJSON:  `{"unknown_keyword": true}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "unknown keyword",
		},
		{
			name:        "Invalid dependency format",
			schemaJSON:  `{"dependencies": {"field1": 123}}`,
			parseErr:    false,
			compileErr:  true,
			errContains: "invalid dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parsing
			s, err := Parse(tt.schemaJSON)
			if tt.parseErr {
				assert.Error(t, err, "Schema parsing should fail")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
				return
			} else {
				assert.NoError(t, err, "Schema parsing should succeed")
			}

			// Test compilation
			err = s.Compile()
			if tt.compileErr {
				assert.Error(t, err, "Schema compilation should fail")
				assert.Contains(t, err.Error(), tt.errContains, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Schema compilation should succeed")
			}
		})
	}
}

func TestMalformedSchemas(t *testing.T) {
	// Empty or nil schema
	t.Run("Empty schema", func(t *testing.T) {
		s := &Schema{}
		err := s.Compile()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema raw data is nil")
	})

	// Schema with circular references
	t.Run("Circular reference", func(t *testing.T) {
		s := &Schema{
			Raw: map[string]interface{}{
				"properties": map[string]interface{}{
					"self": map[string]interface{}{
						"$ref": "#",
					},
				},
			},
			Mode: ModeStrict,
		}

		err := s.Compile()
		assert.Error(t, err, "应该因不支持的 $ref 关键字而报错")
		if err != nil {
			assert.Contains(t, err.Error(), "unsupported keyword '$ref' in strict mode", "错误信息应包含 $ref 相关内容")
		}
	})

	// Schema with broken references
	t.Run("Broken reference", func(t *testing.T) {
		s := &Schema{
			Raw: map[string]interface{}{
				"properties": map[string]interface{}{
					"field": map[string]interface{}{
						"$ref": "#/definitions/nonexistent",
					},
				},
			},
			Mode: ModeStrict,
		}

		err := s.Compile()
		assert.Error(t, err, "应该因不支持的 $ref 关键字而报错")
		if err != nil {
			assert.Contains(t, err.Error(), "unsupported keyword '$ref' in strict mode", "错误信息应包含 $ref 相关内容")
		}
	})
}
