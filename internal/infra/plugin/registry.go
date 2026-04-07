package plugin

// Registry 插件注册表
type Registry struct {
	resourcePlugins map[string]ResourcePlugin
	triggerPlugins  map[string]TriggerPlugin
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		resourcePlugins: make(map[string]ResourcePlugin),
		triggerPlugins:  make(map[string]TriggerPlugin),
	}
}

// RegisterResource 注册资源插件
func (r *Registry) RegisterResource(p ResourcePlugin) {
	meta := p.PluginMeta()
	r.resourcePlugins[meta.Name] = p
}

// RegisterTrigger 注册触发器插件
func (r *Registry) RegisterTrigger(p TriggerPlugin) {
	meta := p.PluginMeta()
	r.triggerPlugins[meta.Name] = p
}

// GetResource 获取资源插件
func (r *Registry) GetResource(name string) (ResourcePlugin, bool) {
	p, ok := r.resourcePlugins[name]
	return p, ok
}

// GetTrigger 获取触发器插件
func (r *Registry) GetTrigger(name string) (TriggerPlugin, bool) {
	p, ok := r.triggerPlugins[name]
	return p, ok
}

// ListResources 列出所有资源插件
func (r *Registry) ListResources() []PluginMeta {
	metas := make([]PluginMeta, 0, len(r.resourcePlugins))
	for _, p := range r.resourcePlugins {
		metas = append(metas, p.PluginMeta())
	}
	return metas
}

// ListTriggers 列出所有触发器插件
func (r *Registry) ListTriggers() []PluginMeta {
	metas := make([]PluginMeta, 0, len(r.triggerPlugins))
	for _, p := range r.triggerPlugins {
		metas = append(metas, p.PluginMeta())
	}
	return metas
}