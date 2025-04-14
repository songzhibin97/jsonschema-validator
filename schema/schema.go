package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// ValidationMode 定义验证模式
type ValidationMode int

const (
	ModeStrict ValidationMode = iota
	ModeLoose
	ModeWarn
)

// Schema 表示JSON Schema
type Schema struct {
	Raw         map[string]interface{}
	Compiled    *CompiledSchema
	ID          string
	Title       string
	Description string
	Mode        ValidationMode
}

// CompiledSchema 表示编译后的Schema
type CompiledSchema struct {
	Keywords   map[string]interface{}
	TypeRules  map[string][]string
	SubSchemas map[string]*CompiledSchema
}

// Parse 解析JSON字符串为Schema
func Parse(jsonSchema string) (*Schema, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonSchema), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	schema := &Schema{
		Raw:  raw,
		Mode: ModeStrict,
	}

	if id, ok := raw["$id"].(string); ok {
		schema.ID = id
	}
	if title, ok := raw["title"].(string); ok {
		schema.Title = title
	}
	if desc, ok := raw["description"].(string); ok {
		schema.Description = desc
	}

	return schema, nil
}

// Compile 编译Schema以提高性能
func (s *Schema) Compile() error {
	if s.Raw == nil {
		return fmt.Errorf("schema raw data is nil")
	}

	compiled := &CompiledSchema{
		Keywords:   make(map[string]interface{}),
		TypeRules:  make(map[string][]string),
		SubSchemas: make(map[string]*CompiledSchema),
	}

	// 处理类型关键字
	if typeVal, ok := s.Raw["type"]; ok {
		switch v := typeVal.(type) {
		case string:
			compiled.Keywords["type"] = v
			compiled.TypeRules["primary"] = []string{v}
		case []interface{}:
			types := make([]string, 0, len(v))
			for _, t := range v {
				if ts, ok := t.(string); ok {
					types = append(types, ts)
				} else {
					return fmt.Errorf("type array contains non-string value: %v", t)
				}
			}
			compiled.Keywords["type"] = types
			compiled.TypeRules["alternatives"] = types
		default:
			return fmt.Errorf("invalid type value: %v", v)
		}
	}

	// 处理数值约束关键字
	for _, key := range []string{"minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf"} {
		if val, ok := s.Raw[key]; ok {
			if num, ok := val.(float64); ok {
				compiled.Keywords[key] = num
			} else {
				return fmt.Errorf("invalid %s value: expected number, got %T", key, val)
			}
		}
	}

	// 处理字符串约束关键字
	for _, key := range []string{"minLength", "maxLength"} {
		if val, ok := s.Raw[key]; ok {
			if num, ok := val.(float64); ok {
				compiled.Keywords[key] = int(num)
			} else {
				return fmt.Errorf("invalid %s value: expected integer, got %T", key, val)
			}
		}
	}

	if pattern, ok := s.Raw["pattern"]; ok {
		if str, ok := pattern.(string); ok {
			compiled.Keywords["pattern"] = str
		} else {
			return fmt.Errorf("invalid pattern value: expected string, got %T", pattern)
		}
	}

	// 处理数组约束关键字
	for _, key := range []string{"minItems", "maxItems"} {
		if val, ok := s.Raw[key]; ok {
			if num, ok := val.(float64); ok {
				compiled.Keywords[key] = int(num)
			} else {
				return fmt.Errorf("invalid %s value: expected integer, got %T", key, val)
			}
		}
	}

	// 处理属性关键字
	if props, ok := s.Raw["properties"].(map[string]interface{}); ok {
		propSchemas := make(map[string]*CompiledSchema)
		for propName, propSchema := range props {
			ps, ok := propSchema.(map[string]interface{})
			if !ok {
				return fmt.Errorf("property '%s' must be an object, got %T", propName, propSchema)
			}
			subSchema := &Schema{
				Raw:  ps,
				Mode: s.Mode,
			}
			if err := subSchema.Compile(); err != nil {
				return fmt.Errorf("failed to compile property '%s': %w", propName, err)
			}
			propSchemas[propName] = subSchema.Compiled
		}
		compiled.Keywords["properties"] = propSchemas
	}

	// 处理模式属性
	if patternProps, ok := s.Raw["patternProperties"].(map[string]interface{}); ok {
		patternSchemas := make(map[string]*CompiledSchema)
		for pattern, propSchema := range patternProps {
			_, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern in patternProperties: %s - %w", pattern, err)
			}

			ps, ok := propSchema.(map[string]interface{})
			if !ok {
				return fmt.Errorf("pattern property '%s' must be an object, got %T", pattern, propSchema)
			}
			subSchema := &Schema{
				Raw:  ps,
				Mode: s.Mode,
			}
			if err := subSchema.Compile(); err != nil {
				return fmt.Errorf("failed to compile pattern '%s': %w", pattern, err)
			}
			patternSchemas[pattern] = subSchema.Compiled
		}
		compiled.Keywords["patternProperties"] = patternSchemas
	}

	// 处理依赖
	if deps, ok := s.Raw["dependencies"].(map[string]interface{}); ok {
		depSchemas := make(map[string]interface{})
		for depName, depSchema := range deps {
			switch v := depSchema.(type) {
			case []interface{}:
				var fields []string
				for _, f := range v {
					if fs, ok := f.(string); ok {
						fields = append(fields, fs)
					} else {
						return fmt.Errorf("dependency '%s' contains non-string field: %v", depName, f)
					}
				}
				depSchemas[depName] = fields
			case map[string]interface{}:
				subSchema := &Schema{
					Raw:  v,
					Mode: s.Mode,
				}
				if err := subSchema.Compile(); err != nil {
					return fmt.Errorf("failed to compile dependency '%s': %w", depName, err)
				}
				depSchemas[depName] = subSchema.Compiled
			default:
				return fmt.Errorf("invalid dependency '%s': %v", depName, v)
			}
		}
		compiled.Keywords["dependencies"] = depSchemas
	}

	// 处理数组元素
	if items, ok := s.Raw["items"]; ok {
		switch v := items.(type) {
		case map[string]interface{}:
			subSchema := &Schema{
				Raw:  v,
				Mode: s.Mode,
			}
			if err := subSchema.Compile(); err != nil {
				return fmt.Errorf("failed to compile items: %w", err)
			}
			compiled.Keywords["items"] = subSchema.Compiled
		case []interface{}:
			itemSchemas := make([]*CompiledSchema, 0, len(v))
			for i, item := range v {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("items[%d] must be an object, got %T", i, item)
				}
				subSchema := &Schema{
					Raw:  itemMap,
					Mode: s.Mode,
				}
				if err := subSchema.Compile(); err != nil {
					return fmt.Errorf("failed to compile items[%d]: %w", i, err)
				}
				itemSchemas = append(itemSchemas, subSchema.Compiled)
			}
			compiled.Keywords["items"] = itemSchemas
		default:
			return fmt.Errorf("invalid items value: %T", v)
		}
	}

	// 处理额外属性
	if additionalProps, ok := s.Raw["additionalProperties"]; ok {
		if schemaMap, ok := additionalProps.(map[string]interface{}); ok {
			subSchema := &Schema{
				Raw:  schemaMap,
				Mode: s.Mode,
			}
			if err := subSchema.Compile(); err != nil {
				return fmt.Errorf("failed to compile additionalProperties: %w", err)
			}
			compiled.Keywords["additionalProperties"] = subSchema.Compiled
		} else if _, ok := additionalProps.(bool); ok {
			compiled.Keywords["additionalProperties"] = additionalProps
		} else {
			return fmt.Errorf("invalid additionalProperties value: %T", additionalProps)
		}
	}

	// 处理必需字段关键字
	if required, ok := s.Raw["required"].([]interface{}); ok {
		var requiredFields []string
		for i, field := range required {
			f, ok := field.(string)
			if !ok {
				return fmt.Errorf("required[%d] must be a string, got %T", i, field)
			}
			requiredFields = append(requiredFields, f)
		}
		compiled.Keywords["required"] = requiredFields
	}

	// 显式检查 $ref
	for key := range s.Raw {
		if key == "$ref" && s.Mode == ModeStrict {
			return fmt.Errorf("unsupported keyword '$ref' in strict mode")
		}
	}

	// 处理其他关键字
	for key, value := range s.Raw {
		if _, exists := compiled.Keywords[key]; !exists {
			if s.Mode == ModeStrict {
				if !isMetadataKey(key) && !isKnownValidationKey(key) {
					return fmt.Errorf("unknown keyword '%s' in strict mode", key)
				}
			}
			compiled.Keywords[key] = value
		}
	}

	s.Compiled = compiled
	return nil
}

// isMetadataKey 检查关键字是否为元数据
func isMetadataKey(key string) bool {
	return key == "$id" || key == "title" || key == "description" || key == "$schema" || key == "$comment"
}

// isKnownValidationKey 检查是否为已知的验证关键字
func isKnownValidationKey(key string) bool {
	knownKeys := map[string]bool{
		"minimum":          true,
		"maximum":          true,
		"exclusiveMinimum": true,
		"exclusiveMaximum": true,
		"multipleOf":       true,
		"minLength":        true,
		"maxLength":        true,
		"pattern":          true,
		"format":           true,
		"minItems":         true,
		"maxItems":         true,
		"uniqueItems":      true,
		"enum":             true,
	}
	return knownKeys[key]
}

// SetMode 设置Schema的验证模式
func (s *Schema) SetMode(mode ValidationMode) {
	s.Mode = mode
}

// String 返回Schema的字符串表示
func (s *Schema) String() string {
	if s.Raw == nil {
		return "{}"
	}
	bytes, err := json.MarshalIndent(s.Raw, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling schema: %v", err)
	}
	return string(bytes)
}

// MarshalJSON 实现json.Marshaler接口
func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.Raw == nil {
		return json.Marshal(map[string]interface{}{})
	}
	return json.Marshal(s.Raw)
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (s *Schema) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Raw = raw
	if id, ok := raw["$id"].(string); ok {
		s.ID = id
	}
	if title, ok := raw["title"].(string); ok {
		s.Title = title
	}
	if desc, ok := raw["description"].(string); ok {
		s.Description = desc
	}

	return nil
}

// GetType 获取Schema定义的类型
func (s *Schema) GetType() interface{} {
	if s.Compiled != nil {
		return s.Compiled.Keywords["type"]
	}
	return s.Raw["type"]
}

// HasKeyword 检查Schema是否包含指定关键字
func (s *Schema) HasKeyword(keyword string) bool {
	if s.Raw == nil {
		return false
	}
	_, exists := s.Raw[keyword]
	return exists
}

// GetKeyword 获取指定关键字的值
func (s *Schema) GetKeyword(keyword string) interface{} {
	if s.Raw == nil {
		return nil
	}
	return s.Raw[keyword]
}
