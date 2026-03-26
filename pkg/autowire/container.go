// Package autowire 提供轻量级的依赖注入容器，基于 Go 泛型实现类型安全的注册与解析。
//
// 设计原则：
//   - 零反射：通过泛型在编译期保证类型安全
//   - 延迟初始化：Provider 函数在首次 Resolve 时才执行
//   - 单例模式：每个类型只创建一次实例，后续调用返回缓存
//   - 显式依赖：Provider 函数通过 Resolve 声明依赖，形成清晰的依赖链
//
// 使用示例：
//
//	c := autowire.New()
//	autowire.Provide(c, func(c *autowire.Container) (UserRepository, error) {
//	    return mysql.NewUserRepository(db), nil
//	})
//	autowire.Provide(c, func(c *autowire.Container) (UserService, error) {
//	    repo := autowire.MustResolve[UserRepository](c)
//	    return service.NewUserService(repo), nil
//	})
//	svc := autowire.MustResolve[UserService](c)
package autowire

import (
	"fmt"
	"reflect"
	"sync"
)

// Container 是依赖注入容器，存储所有 Provider 和已解析的单例实例。
type Container struct {
	mu        sync.RWMutex
	providers map[reflect.Type]provider
	instances map[reflect.Type]any
	resolving map[reflect.Type]bool // 循环依赖检测
}

// provider 包装了一个延迟初始化函数。
type provider struct {
	fn func(*Container) (any, error)
}

// New 创建一个新的依赖注入容器。
func New() *Container {
	return &Container{
		providers: make(map[reflect.Type]provider),
		instances: make(map[reflect.Type]any),
		resolving: make(map[reflect.Type]bool),
	}
}

// Provide 注册一个类型的 Provider 函数。
// Provider 在首次 Resolve 时执行，结果会被缓存（单例）。
// 如果同一类型重复注册，后注册的会覆盖前者。
func Provide[T any](c *Container, fn func(*Container) (T, error)) {
	t := typeOf[T]()
	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[t] = provider{
		fn: func(c *Container) (any, error) {
			return fn(c)
		},
	}
}

// Resolve 解析指定类型的实例。
// 如果实例已缓存则直接返回；否则调用 Provider 创建并缓存。
// 返回错误：类型未注册、循环依赖、Provider 执行失败。
func Resolve[T any](c *Container) (T, error) {
	t := typeOf[T]()
	var zero T

	// 快速路径：已有缓存实例
	c.mu.RLock()
	if instance, ok := c.instances[t]; ok {
		c.mu.RUnlock()
		return instance.(T), nil
	}
	c.mu.RUnlock()

	// 慢路径：需要创建
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if instance, ok := c.instances[t]; ok {
		return instance.(T), nil
	}

	// 循环依赖检测
	if c.resolving[t] {
		return zero, fmt.Errorf("autowire: circular dependency detected for type %s", t)
	}

	p, ok := c.providers[t]
	if !ok {
		return zero, fmt.Errorf("autowire: no provider registered for type %s", t)
	}

	c.resolving[t] = true
	defer delete(c.resolving, t)

	// 临时释放锁让 Provider 内部可以递归 Resolve
	c.mu.Unlock()
	instance, err := p.fn(c)
	c.mu.Lock()

	if err != nil {
		return zero, fmt.Errorf("autowire: failed to create %s: %w", t, err)
	}

	c.instances[t] = instance
	return instance.(T), nil
}

// MustResolve 解析指定类型的实例，失败时 panic。
// 适用于应用启动阶段，此时依赖缺失应立即终止。
func MustResolve[T any](c *Container) T {
	instance, err := Resolve[T](c)
	if err != nil {
		panic(err)
	}
	return instance
}

// typeOf 获取泛型类型 T 的 reflect.Type。
func typeOf[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}
