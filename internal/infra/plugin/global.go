package plugin

import "sync"

var (
	globalRegistry *Registry
	once           sync.Once
)

// GlobalRegistry 获取全局插件注册表
func GlobalRegistry() *Registry {
	once.Do(func() {
		globalRegistry = NewRegistry()
	})
	return globalRegistry
}