package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecursiveSchema(t *testing.T) {
	// 带有自引用的 schema
	schemaJSON := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"children": {
				"type": "array",
				"items": {"$ref": "#"}
			}
		}
	}`

	// 解析 schema
	s, err := Parse(schemaJSON)
	assert.NoError(t, err)

	// 编译 schema - 目前实现中这会抛出错误，因为不支持自引用
	// 这个测试会失败，表明需要实现 $ref 支持
	err = s.Compile()

	// 当前实现应该会失败
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "$ref")
}
