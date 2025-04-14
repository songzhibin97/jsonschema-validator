package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplexNestedSchema(t *testing.T) {
	// 构建一个复杂的多层嵌套 schema
	schema := `{
		"type": "object",
		"properties": {
			"level1": {
				"type": "object",
				"properties": {
					"level2": {
						"type": "object",
						"properties": {
							"level3": {
								"type": "object",
								"properties": {
									"level4": {
										"type": "object",
										"properties": {
											"level5": {
												"type": "string",
												"minLength": 2,
												"maxLength": 10,
												"pattern": "^[a-z]+$"
											},
											"array5": {
												"type": "array",
												"items": {
													"type": "object",
													"properties": {
														"key": {"type": "string"},
														"value": {"type": "number"}
													},
													"required": ["key", "value"]
												},
												"minItems": 1
											}
										},
										"required": ["level5"]
									},
									"arrays": {
										"type": "array",
										"items": {
											"type": "string"
										}
									}
								},
								"additionalProperties": false
							}
						}
					}
				}
			}
		}
	}`

	testCases := []struct {
		name        string
		data        string
		expectValid bool
		errorCount  int
		errContains string
	}{
		{
			name: "Valid deeply nested data",
			data: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": {
								"level5": "valid",
								"array5": [
									{"key": "item1", "value": 10},
									{"key": "item2", "value": 20}
								]
							},
							"arrays": ["string1", "string2"]
						}
					}
				}
			}`,
			expectValid: true,
		},
		{
			name: "Invalid - pattern violation at level5",
			data: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": {
								"level5": "INVALID",
								"array5": [
									{"key": "item1", "value": 10}
								]
							}
						}
					}
				}
			}`,
			expectValid: false,
			errContains: "pattern",
		},
		{
			name: "Invalid - missing required property",
			data: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": {
								"array5": [
									{"key": "item1", "value": 10}
								]
							}
						}
					}
				}
			}`,
			expectValid: false,
			errContains: "required",
		},
		{
			name: "Invalid - additionalProperties",
			data: `{
                "level1": {
                    "level2": {
                        "level3": {
                            "level4": {
                                "level5": "valid"
                            },
                            "extraProp": "should not be here"
                        }
                    }
                }
            }`,
			expectValid: false,
			// 根据当前实现，错误消息可能是"unknown field"
			errContains: "unknown field",
		},
		{
			name: "Invalid - deep array item type",
			data: `{
				"level1": {
					"level2": {
						"level3": {
							"level4": {
								"level5": "valid",
								"array5": [
									{"key": "item1", "value": "not a number"}
								]
							}
						}
					}
				}
			}`,
			expectValid: false,
			errContains: "number",
		},
		{
			name: "Invalid - empty required array",
			data: `{
                "level1": {
                    "level2": {
                        "level3": {
                            "level4": {
                                "level5": "valid",
                                "array5": []
                            }
                        }
                    }
                }
            }`,
			expectValid: false,
			// 根据当前实现，错误消息可能是"fewer items"
			errContains: "fewer items",
		},
	}

	v := New(WithValidationMode(0)) // Strict mode

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := v.ValidateJSON(tc.data, schema)
			assert.NoError(t, err, "Schema validation setup should not error")
			assert.Equal(t, tc.expectValid, result.Valid)

			if !tc.expectValid && tc.errContains != "" {
				foundError := false
				for _, e := range result.Errors {
					if assert.NotEmpty(t, e.Message) && assert.NotEmpty(t, e.Path) {
						if strings.Contains(e.Message, tc.errContains) {
							foundError = true
							break
						}
					}
				}
				assert.True(t, foundError, "Expected to find error containing %q", tc.errContains)
			}
		})
	}
}
