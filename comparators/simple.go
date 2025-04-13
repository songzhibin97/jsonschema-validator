package comparators

import (
	"fmt"
	"sync"
)

// SimpleComparatorRegistry 是一个简单的比较器注册表实现
type SimpleComparatorRegistry struct {
	comparators map[string]CompareFunc
	mu          sync.RWMutex
}

// NewSimpleComparatorRegistry 创建一个新的比较器注册表
func NewSimpleComparatorRegistry() *SimpleComparatorRegistry {
	return &SimpleComparatorRegistry{
		comparators: make(map[string]CompareFunc),
	}
}

// RegisterComparator 注册比较器
func (r *SimpleComparatorRegistry) RegisterComparator(name string, fn CompareFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.comparators[name]; exists {
		return fmt.Errorf("comparator %s already registered", name)
	}
	if fn == nil {
		return fmt.Errorf("comparator function cannot be nil")
	}
	r.comparators[name] = fn
	return nil
}

// GetComparator 获取比较器
func (r *SimpleComparatorRegistry) GetComparator(name string) CompareFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.comparators[name]
}
