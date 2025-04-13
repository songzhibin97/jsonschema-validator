package errors

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormattingMode 定义错误格式化方式
type FormattingMode int

const (
	// FormattingModeSimple 简单格式
	FormattingModeSimple FormattingMode = iota

	// FormattingModeDetailed 详细格式
	FormattingModeDetailed

	// FormattingModeJSON JSON格式
	FormattingModeJSON
)

// ValidationError 表示验证错误
type ValidationError struct {
	// Path 指向错误发生的位置
	Path string `json:"path"`

	// Message 错误消息
	Message string `json:"message"`

	// Value 导致错误的值
	Value interface{} `json:"value,omitempty"`

	// Tag 相关的验证标签
	Tag string `json:"tag,omitempty"`

	// Param 相关的参数
	Param string `json:"param,omitempty"`
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s (path: %s)", e.Message, e.Path)
}

// ValidationErrors 表示多个验证错误
type ValidationErrors []ValidationError

// Error 实现error接口
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("validation failed with the following errors:\n")

	for i, err := range ve {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, err.Error()))
	}

	return sb.String()
}

// FormatWithMode 根据指定模式格式化错误信息
func (ve ValidationErrors) FormatWithMode(mode FormattingMode) string {
	switch mode {
	case FormattingModeSimple:
		return ve.formatSimple()
	case FormattingModeDetailed:
		return ve.formatDetailed()
	case FormattingModeJSON:
		return ve.formatJSON()
	default:
		return ve.Error()
	}
}

// formatSimple 简单格式化
func (ve ValidationErrors) formatSimple() string {
	if len(ve) == 0 {
		return ""
	}

	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}

	return strings.Join(messages, "; ")
}

// formatDetailed 详细格式化
func (ve ValidationErrors) formatDetailed() string {
	return ve.Error()
}

// formatJSON JSON格式化
func (ve ValidationErrors) formatJSON() string {
	if len(ve) == 0 {
		return "[]"
	}
	bytes, err := json.Marshal(ve)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal errors: %v"}`, err)
	}
	return string(bytes)
}

// New 创建一个新的错误
func New(text string) error {
	return fmt.Errorf(text)
}

// ValidationErrorMap 对应不同字段的验证错误
type ValidationErrorMap map[string]ValidationErrors

// Error 实现error接口
func (m ValidationErrorMap) Error() string {
	if len(m) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("validation failed for the following fields:\n")

	for field, errs := range m {
		sb.WriteString(fmt.Sprintf("Field '%s':\n", field))
		for i, err := range errs {
			sb.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, err.Message))
		}
	}

	return sb.String()
}
