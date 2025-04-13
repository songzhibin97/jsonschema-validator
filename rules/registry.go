package rules

import (
	"sync"
)

// ValidatorRegistry 接口定义了验证器注册表的行为
type ValidatorRegistry interface {
	// RegisterValidator 注册验证器
	RegisterValidator(name string, fn RuleFunc) error

	// GetValidator 获取验证器
	GetValidator(name string) RuleFunc
}

// Registry 是规则注册表的实现
type Registry struct {
	rules map[string]RuleFunc
	mutex sync.RWMutex
}

// NewRegistry 创建一个新的规则注册表
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]RuleFunc),
	}
}

// Register 注册一个规则
func (r *Registry) Register(name string, rule Rule) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.rules[name] = rule.Validate
}

// RegisterFunc 注册一个规则函数
func (r *Registry) RegisterFunc(name string, fn RuleFunc) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.rules[name] = fn
}

// Get 获取一个规则函数
func (r *Registry) Get(name string) RuleFunc {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.rules[name]
}

// Has 检查是否存在指定名称的规则
func (r *Registry) Has(name string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	_, exists := r.rules[name]
	return exists
}

// Names 获取所有注册的规则名称
func (r *Registry) Names() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	names := make([]string, 0, len(r.rules))
	for name := range r.rules {
		names = append(names, name)
	}
	return names
}

// Count 获取注册的规则数量
func (r *Registry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.rules)
}

// Clear 清空所有注册的规则
func (r *Registry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.rules = make(map[string]RuleFunc)
}

// RegisterValidator 实现ValidatorRegistry接口
func (r *Registry) RegisterValidator(name string, fn RuleFunc) error {
	r.RegisterFunc(name, fn)
	return nil
}

// GetValidator 实现ValidatorRegistry接口
func (r *Registry) GetValidator(name string) RuleFunc {
	return r.Get(name)
}

// DefaultRegistry 是全局默认的规则注册表
var DefaultRegistry = NewRegistry()

// RegisterBuiltInRules 注册所有内置规则到指定的注册表
func RegisterBuiltInRules(registry ValidatorRegistry) {
	registerTypeRules(registry)
	registerNumberRules(registry)
	registerStringRules(registry)
	registerArrayRules(registry)
	registerObjectRules(registry)
	registerFormatRules(registry)
	registerLogicalRules(registry)
	registerConditionalRules(registry)

}

// RegisterAll 注册所有内置规则到默认注册表
func RegisterAll() {
	RegisterBuiltInRules(DefaultRegistry)
}

func ClearDefaultRegistry() {
	DefaultRegistry.Clear()
}

//// init 在包初始化时注册所有内置规则
//func init() {
//	RegisterAll()
//}
