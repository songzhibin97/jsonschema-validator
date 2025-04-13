package schema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		jsonSchema  string
		expectID    string
		expectTitle string
		expectDesc  string
		expectErr   bool
	}{
		{
			name:        "Valid schema",
			jsonSchema:  `{"$id":"test-schema","title":"Test Schema","description":"A test schema","type":"object"}`,
			expectID:    "test-schema",
			expectTitle: "Test Schema",
			expectDesc:  "A test schema",
		},
		{
			name:       "Invalid JSON",
			jsonSchema: `{"type":"object"`, // Missing closing brace
			expectErr:  true,
		},
		{
			name:       "Empty schema",
			jsonSchema: `{}`,
			expectID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(tt.jsonSchema)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectID, s.ID)
			assert.Equal(t, tt.expectTitle, s.Title)
			assert.Equal(t, tt.expectDesc, s.Description)
			assert.NotNil(t, s.Raw)
			assert.Equal(t, ModeStrict, s.Mode)
		})
	}
}

func TestCompile(t *testing.T) {
	tests := []struct {
		name      string
		schema    *Schema
		expectErr string
	}{
		{
			name: "Simple schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"name"},
				},
			},
		},
		{
			name: "Pattern properties",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
					"patternProperties": map[string]interface{}{
						"^foo.*": map[string]interface{}{"type": "string"},
					},
				},
			},
		},
		{
			name: "Dependencies",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
					"dependencies": map[string]interface{}{
						"field1": []interface{}{"field2"},
						"field3": map[string]interface{}{"type": "number"},
					},
				},
			},
		},
		{
			name: "Items tuple",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "array",
					"items": []interface{}{
						map[string]interface{}{"type": "string"},
						map[string]interface{}{"type": "number"},
					},
				},
			},
		},
		{
			name: "Additional properties schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
					"additionalProperties": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		{
			name: "Additional properties boolean",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type":                 "object",
					"additionalProperties": false,
				},
			},
		},
		{
			name: "Multiple types",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": []interface{}{"string", "null"},
				},
			},
		},
		{
			name: "Invalid property schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"properties": map[string]interface{}{
						"name": "string", // Should be an object
					},
				},
			},
			expectErr: "property 'name' must be an object",
		},
		{
			name: "Invalid pattern property",
			schema: &Schema{
				Raw: map[string]interface{}{
					"patternProperties": map[string]interface{}{
						"^foo": "string",
					},
				},
			},
			expectErr: "pattern property '^foo' must be an object",
		},
		{
			name: "Invalid required field",
			schema: &Schema{
				Raw: map[string]interface{}{
					"required": []interface{}{42},
				},
			},
			expectErr: "required[0] must be a string",
		},
		{
			name: "Invalid type array",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": []interface{}{"string", 42},
				},
			},
			expectErr: "type array contains non-string value",
		},
		{
			name: "Unknown keyword strict",
			schema: &Schema{
				Raw: map[string]interface{}{
					"unknown": "value",
				},
				Mode: ModeStrict,
			},
			expectErr: "unknown keyword 'unknown' in strict mode",
		},
		{
			name: "Unknown keyword loose",
			schema: &Schema{
				Raw: map[string]interface{}{
					"unknown": "value",
				},
				Mode: ModeLoose,
			},
		},
		{
			name: "Nil raw",
			schema: &Schema{
				Raw: nil,
			},
			expectErr: "schema raw data is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schema.Compile()
			if tt.expectErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tt.schema.Compiled)
				assert.NotNil(t, tt.schema.Compiled.Keywords)
				assert.NotNil(t, tt.schema.Compiled.SubSchemas)
			}
		})
	}
}

func TestSetMode(t *testing.T) {
	s := &Schema{}
	s.SetMode(ModeLoose)
	assert.Equal(t, ModeLoose, s.Mode)

	s.SetMode(ModeStrict)
	assert.Equal(t, ModeStrict, s.Mode)

	s.SetMode(ModeWarn)
	assert.Equal(t, ModeWarn, s.Mode)
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		schema   *Schema
		expected string
	}{
		{
			name: "Valid schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
					"name": "test",
				},
			},
		},
		{
			name:     "Nil raw",
			schema:   &Schema{Raw: nil},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.schema.String()
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				var expected map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(result), &expected))
				assert.Equal(t, tt.schema.Raw, expected)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		schema *Schema
	}{
		{
			name: "Valid schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
					"name": "test",
				},
			},
		},
		{
			name:   "Nil raw",
			schema: &Schema{Raw: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := tt.schema.MarshalJSON()
			assert.NoError(t, err)
			var result map[string]interface{}
			assert.NoError(t, json.Unmarshal(bytes, &result))
			if tt.schema.Raw == nil {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.schema.Raw, result)
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		expectID  string
		expectErr bool
	}{
		{
			name:     "Valid schema",
			data:     `{"$id":"test","type":"object"}`,
			expectID: "test",
		},
		{
			name:      "Invalid JSON",
			data:      `{`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Schema{}
			err := s.UnmarshalJSON([]byte(tt.data))
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectID, s.ID)
			if tt.data != "{" {
				var raw map[string]interface{}
				assert.NoError(t, json.Unmarshal([]byte(tt.data), &raw))
				assert.Equal(t, raw, s.Raw)
			}
		})
	}
}

func TestGetType(t *testing.T) {
	tests := []struct {
		name     string
		schema   *Schema
		expected interface{}
	}{
		{
			name: "Compiled schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "string",
				},
			},
			expected: "string",
		},
		{
			name: "Uncompiled schema",
			schema: &Schema{
				Raw: map[string]interface{}{
					"type": "object",
				},
			},
			expected: "object",
		},
		{
			name:     "No type",
			schema:   &Schema{Raw: map[string]interface{}{}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Compiled schema" {
				assert.NoError(t, tt.schema.Compile())
			}
			assert.Equal(t, tt.expected, tt.schema.GetType())
		})
	}
}

func TestHasKeyword(t *testing.T) {
	s := &Schema{
		Raw: map[string]interface{}{
			"type":     "string",
			"required": []interface{}{"name"},
		},
	}
	assert.True(t, s.HasKeyword("type"))
	assert.True(t, s.HasKeyword("required"))
	assert.False(t, s.HasKeyword("unknown"))
	assert.False(t, (&Schema{Raw: nil}).HasKeyword("type"))
}

func TestGetKeyword(t *testing.T) {
	s := &Schema{
		Raw: map[string]interface{}{
			"type":     "string",
			"required": []interface{}{"name"},
		},
	}
	assert.Equal(t, "string", s.GetKeyword("type"))
	assert.Equal(t, []interface{}{"name"}, s.GetKeyword("required"))
	assert.Nil(t, s.GetKeyword("unknown"))
	assert.Nil(t, (&Schema{Raw: nil}).GetKeyword("type"))
}
