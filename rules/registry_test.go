package rules

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryOperations(t *testing.T) {
	registry := NewRegistry()

	// 测试注册
	registry.RegisterValidator("test", func(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
		return true, nil
	})

	assert.True(t, registry.Has("test"))
	assert.NotNil(t, registry.GetValidator("test"))
	assert.Equal(t, 1, registry.Count())
	assert.Equal(t, []string{"test"}, registry.Names())

	// 测试获取不存在的验证器
	assert.Nil(t, registry.GetValidator("unknown"))

	// 测试清空
	registry.Clear()
	assert.False(t, registry.Has("test"))
	assert.Equal(t, 0, registry.Count())
	assert.Empty(t, registry.Names())
}

func TestRegistryConcurrency(t *testing.T) {
	registry := NewRegistry()
	var wg sync.WaitGroup

	// 并发注册
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("test%d", id)
			registry.RegisterValidator(name, func(ctx context.Context, value interface{}, schemaValue interface{}, path string) (bool, error) {
				return true, nil
			})
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 10, registry.Count())

	// 并发获取
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("test%d", id)
			assert.NotNil(t, registry.GetValidator(name))
		}(i)
	}

	wg.Wait()
}
