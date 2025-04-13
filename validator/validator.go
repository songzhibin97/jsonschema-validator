package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/songzhibin97/jsonschema-validator/comparators"
	"github.com/songzhibin97/jsonschema-validator/errors"
	rules2 "github.com/songzhibin97/jsonschema-validator/rules"
	"github.com/songzhibin97/jsonschema-validator/schema"
)

// Validate 是验证函数的签名
type Validate func(ctx context.Context, value interface{}, schema interface{}, path string) (bool, error)

// Validator 表示验证器实例
type Validator struct {
	opts               *Options
	lock               sync.RWMutex
	validators         map[string]rules2.RuleFunc
	comparators        map[string]comparators.CompareFunc
	tagNameFunc        func(field reflect.StructField) string
	customTypeFunc     func(field reflect.Value) interface{}
	customValidateFunc func(ctx context.Context, value interface{}, path string) (bool, error)
	cache              *sync.Map
}

// New 创建一个新的验证器实例
func New(opts ...Option) *Validator {
	options := &Options{
		TagName:             "validate",
		ValidationMode:      schema.ModeStrict,
		ErrorFormattingMode: errors.FormattingModeDetailed,
	}
	for _, opt := range opts {
		opt(options)
	}

	v := &Validator{
		opts:        options,
		validators:  make(map[string]rules2.RuleFunc),
		comparators: make(map[string]comparators.CompareFunc),
		cache:       &sync.Map{},
	}

	// 注册内置规则和比较器
	rules2.RegisterBuiltInRules(v)
	comparators.RegisterBuiltInComparators(v)

	return v
}

// RegisterValidator 注册自定义验证器
func (v *Validator) RegisterValidator(name string, fn rules2.RuleFunc) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	if name == "" {
		return errors.New("validator name cannot be empty")
	}
	if fn == nil {
		return errors.New("validator function cannot be nil")
	}
	v.validators[name] = fn
	return nil
}

// RegisterValidatorMust 注册自定义验证器（出错时panic）
func (v *Validator) RegisterValidatorMust(name string, fn rules2.RuleFunc) {
	if err := v.RegisterValidator(name, fn); err != nil {
		panic(err)
	}
}

// RegisterComparator 注册自定义比较函数
func (v *Validator) RegisterComparator(name string, fn comparators.CompareFunc) error {
	v.lock.Lock()
	defer v.lock.Unlock()
	if name == "" {
		return errors.New("comparator name cannot be empty")
	}
	if fn == nil {
		return errors.New("comparator function cannot be nil")
	}
	v.comparators[name] = fn
	return nil
}

// RegisterComparatorMust 注册自定义比较函数（出错时panic）
func (v *Validator) RegisterComparatorMust(name string, fn comparators.CompareFunc) {
	if err := v.RegisterComparator(name, fn); err != nil {
		panic(err)
	}
}

// SetTagName 设置用于结构体标签的名称
func (v *Validator) SetTagName(name string) {
	v.opts.TagName = name
}

// SetValidationMode 设置验证模式
func (v *Validator) SetValidationMode(mode schema.ValidationMode) {
	v.opts.ValidationMode = mode
}

// SetErrorFormattingMode 设置错误格式化模式
func (v *Validator) SetErrorFormattingMode(mode errors.FormattingMode) {
	v.opts.ErrorFormattingMode = mode
}

// SetCustomTypeFunc 设置自定义类型转换函数
func (v *Validator) SetCustomTypeFunc(fn func(field reflect.Value) interface{}) {
	v.customTypeFunc = fn
}

// SetTagNameFunc 设置自定义标签名称获取函数
func (v *Validator) SetTagNameFunc(fn func(field reflect.StructField) string) {
	v.tagNameFunc = fn
}

func (v *Validator) SetCustomValidateFunc(fn func(ctx context.Context, value interface{}, path string) (bool, error)) {
	v.customValidateFunc = fn
}

// Struct 验证结构体
func (v *Validator) Struct(s interface{}) error {
	return v.StructCtx(context.Background(), s)
}

// StructCtx 带上下文的结构体验证
func (v *Validator) StructCtx(ctx context.Context, s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}

	result := &ValidationResult{Valid: true, Errors: []errors.ValidationError{}}
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// 获取标签
		tag := field.Tag.Get(v.opts.TagName)
		if v.tagNameFunc != nil {
			tag = v.tagNameFunc(field)
		}
		if tag == "" {
			continue
		}

		schemaMap := v.parseTag(tag)
		if len(schemaMap) == 0 {
			continue
		}

		path := field.Name
		fieldValue := value.Interface()
		if v.customTypeFunc != nil {
			fieldValue = v.customTypeFunc(value)
		}

		// 自定义验证
		if v.customValidateFunc != nil {
			isValid, err := v.customValidateFunc(ctx, fieldValue, path)
			if err != nil {
				return fmt.Errorf("custom validation failed for %s: %w", path, err)
			}
			if !isValid {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: "value must start with 'ADMIN_'",
					Tag:     "custom",
					Value:   fieldValue,
				})
				if v.opts.StopOnFirstError {
					return errors.ValidationErrors(result.Errors)
				}
				continue
			}
		}

		// 处理 required
		if _, isRequired := schemaMap["required"]; isRequired {
			if isZero(value) {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: "field is required",
					Tag:     "required",
				})
				if v.opts.StopOnFirstError {
					return errors.ValidationErrors(result.Errors)
				}
				continue
			}
			delete(schemaMap, "required")
		}

		// 递归验证嵌套结构体
		if v.opts.RecursiveValidation && value.Kind() == reflect.Struct {
			if err := v.StructCtx(ctx, fieldValue); err != nil {
				if ve, ok := err.(errors.ValidationErrors); ok {
					for _, e := range ve {
						e.Path = path + "." + e.Path
						result.Errors = append(result.Errors, e)
					}
					result.Valid = false
					if v.opts.StopOnFirstError {
						return errors.ValidationErrors(result.Errors)
					}
				}
			}
			continue
		}

		// 验证其他规则
		fieldResult, err := v.ValidateWithSchema(fieldValue, schemaMap, path)
		if err != nil {
			return err
		}
		if !fieldResult.Valid {
			result.Valid = false
			result.Errors = append(result.Errors, fieldResult.Errors...)
			if v.opts.StopOnFirstError {
				return errors.ValidationErrors(result.Errors)
			}
		}
	}

	if !result.Valid {
		return errors.ValidationErrors(result.Errors)
	}
	return nil
}

// Var 验证单个变量
func (v *Validator) Var(field interface{}, tag string) error {
	return v.VarCtx(context.Background(), field, tag)
}

// VarCtx 带上下文的单个变量验证
func (v *Validator) VarCtx(ctx context.Context, field interface{}, tag string) error {
	schemaMap := v.parseTag(tag)
	if len(schemaMap) == 0 {
		return nil
	}
	result, err := v.ValidateWithSchema(field, schemaMap, "var")
	if err != nil {
		return err
	}
	if !result.Valid {
		return errors.ValidationErrors(result.Errors)
	}
	return nil
}

// ValidateJSON 验证JSON字符串是否符合指定的schema
func (v *Validator) ValidateJSON(jsonData string, schemaJSON string) (*ValidationResult, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON data: %w", err)
	}

	// 检查缓存
	if v.opts.EnableCaching {
		if cached, ok := v.cache.Load(schemaJSON); ok {
			if s, ok := cached.(*schema.Schema); ok && s.Compiled != nil {
				return v.validateCompiledSchema(data, s, "$")
			}
		}
	}

	// 解析和编译 schema
	s, err := schema.Parse(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid schema JSON: %w", err)
	}
	if err := s.Compile(); err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}
	if v.opts.EnableCaching {
		v.cache.Store(schemaJSON, s)
	}

	return v.validateCompiledSchema(data, s, "$")
}

// validateCompiledSchema 使用编译后的 schema 验证
// validator.go
func (v *Validator) validateCompiledSchema(value interface{}, s *schema.Schema, path string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true, Errors: []errors.ValidationError{}}
	ctx := context.WithValue(context.Background(), "validator", v)
	ctx = context.WithValue(ctx, "validationMode", int(s.Mode))

	// 验证顶层 required 关键字
	if required, ok := s.Compiled.Keywords["required"].([]string); ok {
		if obj, ok := value.(map[string]interface{}); ok {
			for _, req := range required {
				if _, exists := obj[req]; !exists {
					result.Valid = false
					result.Errors = append(result.Errors, errors.ValidationError{
						Path:    path + "." + req,
						Message: fmt.Sprintf("required property '%s' is missing", req),
						Tag:     "required",
					})
					if v.opts.StopOnFirstError {
						return result, nil
					}
				}
			}
		} else {
			result.Valid = false
			result.Errors = append(result.Errors, errors.ValidationError{
				Path:    path,
				Message: "value must be an object for required validation",
				Tag:     "required",
			})
			if v.opts.StopOnFirstError {
				return result, nil
			}
		}
	}

	// 处理其他关键字
	for keyword, schemaValue := range s.Compiled.Keywords {
		if keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" || keyword == "required" {
			continue
		}

		// 处理类型关键字
		if keyword == "type" {
			validator, exists := v.validators["type"]
			if exists {
				isValid, err := validator(ctx, value, schemaValue, path)
				if err != nil {
					validErr, ok := err.(*errors.ValidationError)
					if ok {
						result.Valid = false
						result.Errors = append(result.Errors, *validErr)
					} else {
						return nil, fmt.Errorf("validation error: %w", err)
					}
				} else if !isValid {
					result.Valid = false
				}
				if !result.Valid && v.opts.StopOnFirstError {
					return result, nil
				}
			}
			continue
		}

		// 处理属性关键字
		if keyword == "properties" {
			props, ok := schemaValue.(map[string]*schema.CompiledSchema)
			if !ok {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: fmt.Sprintf("properties must be a schema map, got %T", schemaValue),
					Tag:     "properties",
				})
				if v.opts.StopOnFirstError {
					return result, nil
				}
				continue
			}
			if obj, ok := value.(map[string]interface{}); ok {
				for propName, propSchema := range props {
					propPath := path + "." + propName
					if propValue, exists := obj[propName]; exists {
						propResult, err := v.validateCompiledSchema(propValue, &schema.Schema{Compiled: propSchema, Mode: s.Mode}, propPath)
						if err != nil {
							return nil, err
						}
						if !propResult.Valid {
							result.Valid = false
							result.Errors = append(result.Errors, propResult.Errors...)
							if v.opts.StopOnFirstError {
								return result, nil
							}
						}
					}
				}
			} else if s.Compiled.Keywords["type"] == "object" {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: "value must be an object",
					Tag:     "properties",
				})
				if v.opts.StopOnFirstError {
					return result, nil
				}
			}
			continue
		}

		// 处理数组元素
		if keyword == "items" {
			itemsSchema, ok := schemaValue.(*schema.CompiledSchema)
			if !ok {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: fmt.Sprintf("items must be a schema, got %T", schemaValue),
					Tag:     "items",
				})
				if v.opts.StopOnFirstError {
					return result, nil
				}
				continue
			}
			if arr, ok := value.([]interface{}); ok {
				for i, item := range arr {
					itemPath := fmt.Sprintf("%s[%d]", path, i)
					itemResult, err := v.validateCompiledSchema(item, &schema.Schema{Compiled: itemsSchema, Mode: s.Mode}, itemPath)
					if err != nil {
						return nil, err
					}
					if !itemResult.Valid {
						result.Valid = false
						result.Errors = append(result.Errors, itemResult.Errors...)
						if v.opts.StopOnFirstError {
							return result, nil
						}
					}
				}
			} else if s.Compiled.Keywords["type"] == "array" {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: "value must be an array",
					Tag:     "items",
				})
				if v.opts.StopOnFirstError {
					return result, nil
				}
			}
			continue
		}

		// 处理 additionalProperties
		if keyword == "additionalProperties" {
			if additionalProps, ok := schemaValue.(bool); ok && !additionalProps && !v.opts.AllowUnknownFields {
				if obj, ok := value.(map[string]interface{}); ok {
					props, _ := s.Compiled.Keywords["properties"].(map[string]*schema.CompiledSchema)
					for key := range obj {
						if _, exists := props[key]; !exists {
							result.Valid = false
							result.Errors = append(result.Errors, errors.ValidationError{
								Path:    path + "." + key,
								Message: "unknown field",
								Tag:     "additionalProperties",
								Value:   obj[key],
							})
							if v.opts.StopOnFirstError {
								return result, nil
							}
						}
					}
				}
			}
			continue
		}

		// 处理其他验证器
		validator, exists := v.validators[keyword]
		if !exists {
			if s.Mode == schema.ModeStrict && !isMetadataKey(keyword) {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: fmt.Sprintf("unknown validation keyword: %s", keyword),
					Tag:     keyword,
				})
			}
			continue
		}

		isValid, err := validator(ctx, value, schemaValue, path)
		if err != nil {
			validErr, ok := err.(*errors.ValidationError)
			if ok {
				result.Valid = false
				result.Errors = append(result.Errors, *validErr)
			} else {
				return nil, fmt.Errorf("validation error: %w", err)
			}
		} else if !isValid {
			result.Valid = false
			result.Errors = append(result.Errors, errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed for keyword %s", keyword),
				Tag:     keyword,
				Value:   value,
			})
		}

		if !result.Valid && v.opts.StopOnFirstError {
			return result, nil
		}
	}

	return result, nil
}

// isMetadataKey 检查关键字是否为元数据
func isMetadataKey(key string) bool {
	return key == "$id" || key == "title" || key == "description" || key == "$schema" || key == "$comment"
}

// ValidationResult 包含验证结果
type ValidationResult struct {
	Valid  bool                     `json:"valid"`
	Errors []errors.ValidationError `json:"errors,omitempty"`
}

// GetValidator 获取已注册的验证器
func (v *Validator) GetValidator(name string) rules2.RuleFunc {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.validators[name]
}

// GetComparator 获取已注册的比较函数
func (v *Validator) GetComparator(name string) comparators.CompareFunc {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.comparators[name]
}

// parseTag 解析验证标签
func (v *Validator) parseTag(tag string) map[string]interface{} {
	result := make(map[string]interface{})
	if tag == "" {
		return result
	}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if part == "required" {
			result["required"] = true
		} else if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			switch key {
			case "min", "max", "minLength", "maxLength", "minimum", "maximum":
				if num, err := strconv.Atoi(value); err == nil {
					result[key] = num
				} else if num, err := strconv.ParseFloat(value, 64); err == nil {
					result[key] = num
				} else {
					result[key] = value // 保留原始值，留给验证器处理
				}
			case "type", "pattern", "format":
				result[key] = value
			case "enum":
				result[key] = strings.Split(value, "|")
			default:
				result[key] = value
			}
		} else {
			result[part] = true
		}
	}
	return result
}

func isZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	default:
		return false
	}
}

// CompileSchema 编译Schema以提高重复使用的性能
func (v *Validator) CompileSchema(schemaJSON string) (*schema.Schema, error) {
	if v.opts.EnableCaching {
		if cached, ok := v.cache.Load(schemaJSON); ok {
			if s, ok := cached.(*schema.Schema); ok {
				return s, nil
			}
		}
	}

	s, err := schema.Parse(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}
	if err := s.Compile(); err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}
	if v.opts.EnableCaching {
		v.cache.Store(schemaJSON, s)
	}
	return s, nil
}

// ValidateWithSchema 使用指定的schema验证值
func (v *Validator) ValidateWithSchema(value interface{}, schemaMap map[string]interface{}, path string) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true, Errors: []errors.ValidationError{}}
	ctx := context.WithValue(context.Background(), "validator", v)

	// 处理类型关键字
	if typeVal, ok := schemaMap["type"]; ok {
		validator, exists := v.validators["type"]
		if !exists {
			return nil, fmt.Errorf("type validator not found")
		}
		isValid, err := validator(ctx, value, typeVal, path)
		if err != nil {
			if ve, ok := err.(*errors.ValidationError); ok {
				result.Valid = false
				result.Errors = append(result.Errors, *ve)
			} else {
				return nil, err
			}
		} else if !isValid {
			result.Valid = false
		}
		if !result.Valid && v.opts.StopOnFirstError {
			return result, nil
		}
	}

	// 处理必需字段
	if requiredVal, ok := schemaMap["required"]; ok {
		requiredFields, ok := requiredVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("required must be an array")
		}
		obj, ok := value.(map[string]interface{})
		if !ok {
			result.Valid = false
			result.Errors = append(result.Errors, errors.ValidationError{
				Path:    path,
				Message: "value must be an object",
				Tag:     "required",
			})
			if v.opts.StopOnFirstError {
				return result, nil
			}
		}
		for _, field := range requiredFields {
			fieldStr, ok := field.(string)
			if !ok {
				return nil, fmt.Errorf("required field must be a string")
			}
			if _, exists := obj[fieldStr]; !exists {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path + "." + fieldStr,
					Message: fmt.Sprintf("required property '%s' is missing", fieldStr),
					Tag:     "required",
				})
				if v.opts.StopOnFirstError {
					return result, nil
				}
			}
		}
	}

	// 处理对象属性
	if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
		obj, ok := value.(map[string]interface{})
		if !ok && schemaMap["type"] == "object" {
			result.Valid = false
			result.Errors = append(result.Errors, errors.ValidationError{
				Path:    path,
				Message: "value must be an object",
				Tag:     "properties",
			})
			if v.opts.StopOnFirstError {
				return result, nil
			}
		}
		for propName, propSchema := range props {
			propMap, ok := propSchema.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("property '%s' schema must be an object", propName)
			}
			propPath := path + "." + propName
			if propVal, exists := obj[propName]; exists {
				propResult, err := v.ValidateWithSchema(propVal, propMap, propPath)
				if err != nil {
					return nil, err
				}
				if !propResult.Valid {
					result.Valid = false
					result.Errors = append(result.Errors, propResult.Errors...)
					if v.opts.StopOnFirstError {
						return result, nil
					}
				}
			}
		}
	}

	// 处理其他关键字
	for keyword, schemaValue := range schemaMap {
		if keyword == "type" || keyword == "properties" || keyword == "required" || keyword == "title" || keyword == "description" || keyword == "default" || keyword == "examples" {
			continue
		}
		validator, exists := v.validators[keyword]
		if !exists {
			if v.opts.ValidationMode == schema.ModeStrict {
				result.Valid = false
				result.Errors = append(result.Errors, errors.ValidationError{
					Path:    path,
					Message: fmt.Sprintf("unknown validation keyword: %s", keyword),
					Tag:     keyword,
				})
			}
			continue
		}
		isValid, err := validator(ctx, value, schemaValue, path)
		if err != nil {
			if ve, ok := err.(*errors.ValidationError); ok {
				result.Valid = false
				result.Errors = append(result.Errors, *ve)
			} else {
				return nil, fmt.Errorf("validation error: %w", err)
			}
		} else if !isValid {
			result.Valid = false
			result.Errors = append(result.Errors, errors.ValidationError{
				Path:    path,
				Message: fmt.Sprintf("validation failed for keyword %s", keyword),
				Tag:     keyword,
				Value:   value,
			})
		}
		if !result.Valid && v.opts.StopOnFirstError {
			return result, nil
		}
	}

	return result, nil
}

// ClearCache 清理 schema 缓存
func (v *Validator) ClearCache() {
	v.cache.Range(func(key, _ interface{}) bool {
		v.cache.Delete(key)
		return true
	})
}

// Instance 返回一个新的验证器实例
func Instance() *Validator {
	return New()
}

// 全局默认实例
var defaultValidator = Instance()

// RegisterValidator 在默认实例上注册验证器
func RegisterValidator(name string, fn rules2.RuleFunc) error {
	return defaultValidator.RegisterValidator(name, fn)
}

// RegisterValidatorMust 在默认实例上注册验证器（出错时panic）
func RegisterValidatorMust(name string, fn rules2.RuleFunc) {
	defaultValidator.RegisterValidatorMust(name, fn)
}

// RegisterComparator 在默认实例上注册比较函数
func RegisterComparator(name string, fn comparators.CompareFunc) error {
	return defaultValidator.RegisterComparator(name, fn)
}

// RegisterComparatorMust 在默认实例上注册比较函数（出错时panic）
func RegisterComparatorMust(name string, fn comparators.CompareFunc) {
	defaultValidator.RegisterComparatorMust(name, fn)
}

// Struct 使用默认实例验证结构体
func Struct(s interface{}) error {
	return defaultValidator.Struct(s)
}

// StructCtx 使用默认实例验证结构体（带上下文）
func StructCtx(ctx context.Context, s interface{}) error {
	return defaultValidator.StructCtx(ctx, s)
}

// Var 使用默认实例验证变量
func Var(field interface{}, tag string) error {
	return defaultValidator.Var(field, tag)
}

// VarCtx 使用默认实例验证变量（带上下文）
func VarCtx(ctx context.Context, field interface{}, tag string) error {
	return defaultValidator.VarCtx(ctx, field, tag)
}
