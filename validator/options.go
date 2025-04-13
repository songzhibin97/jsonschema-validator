package validator

import (
	"github.com/songzhibin97/jsonschema-validator/errors"
	"github.com/songzhibin97/jsonschema-validator/schema"
)

// Options 包含验证器的配置选项
type Options struct {
	// TagName 是用于结构体验证的标签名
	TagName string

	// ValidationMode 控制验证器的严格程度
	ValidationMode schema.ValidationMode

	// ErrorFormattingMode 控制错误消息的格式化方式
	ErrorFormattingMode errors.FormattingMode

	// EnableCaching 是否启用Schema缓存
	EnableCaching bool

	// RecursiveValidation 是否递归验证嵌套结构
	RecursiveValidation bool

	// StopOnFirstError 是否在第一个错误时停止验证
	StopOnFirstError bool

	// AllowUnknownFields 是否允许数据中包含schema中未定义的字段
	AllowUnknownFields bool
}

// Option 是用于配置验证器的函数选项
type Option func(*Options)

// WithTagName 设置结构体验证的标签名
func WithTagName(name string) Option {
	return func(o *Options) {
		o.TagName = name
	}
}

// WithValidationMode 设置验证模式
func WithValidationMode(mode schema.ValidationMode) Option {
	return func(o *Options) {
		o.ValidationMode = mode
	}
}

// WithErrorFormattingMode 设置错误格式化模式
func WithErrorFormattingMode(mode errors.FormattingMode) Option {
	return func(o *Options) {
		o.ErrorFormattingMode = mode
	}
}

// WithCaching 设置是否启用Schema缓存
func WithCaching(enable bool) Option {
	return func(o *Options) {
		o.EnableCaching = enable
	}
}

// WithRecursiveValidation 设置是否递归验证嵌套结构
func WithRecursiveValidation(enable bool) Option {
	return func(o *Options) {
		o.RecursiveValidation = enable
	}
}

// WithStopOnFirstError 设置是否在第一个错误时停止验证
func WithStopOnFirstError(enable bool) Option {
	return func(o *Options) {
		o.StopOnFirstError = enable
	}
}

// WithAllowUnknownFields 设置是否允许未知字段
func WithAllowUnknownFields(allow bool) Option {
	return func(o *Options) {
		o.AllowUnknownFields = allow
	}
}
