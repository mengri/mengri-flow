package plugin

// Registry 插件注册表
type Registry struct {
	resourcePlugins map[string]ResourcePlugin
	triggerPlugins  map[string]TriggerPlugin
	enabledPlugins  map[string]bool // 启用的插件名称列表
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		resourcePlugins: make(map[string]ResourcePlugin),
		triggerPlugins:  make(map[string]TriggerPlugin),
		enabledPlugins:  make(map[string]bool),
	}
}

// SetEnabledPlugins 设置启用的插件列表
func (r *Registry) SetEnabledPlugins(enabledPlugins []string) {
	r.enabledPlugins = make(map[string]bool)
	for _, name := range enabledPlugins {
		r.enabledPlugins[name] = true
	}
}

// isPluginEnabled 检查插件是否启用
func (r *Registry) isPluginEnabled(name string) bool {
	// 如果没有设置 enabledPlugins，则启用所有插件（向后兼容）
	if len(r.enabledPlugins) == 0 {
		return true
	}
	return r.enabledPlugins[name]
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

// GetResource 获取资源插件（只返回启用的插件）
func (r *Registry) GetResource(name string) (ResourcePlugin, bool) {
	p, ok := r.resourcePlugins[name]
	if !ok || !r.isPluginEnabled(name) {
		return nil, false
	}
	return p, true
}

// GetTrigger 获取触发器插件（只返回启用的插件）
func (r *Registry) GetTrigger(name string) (TriggerPlugin, bool) {
	p, ok := r.triggerPlugins[name]
	if !ok || !r.isPluginEnabled(name) {
		return nil, false
	}
	return p, true
}

// ListResources 列出所有启用的资源插件
func (r *Registry) ListResources() []PluginMeta {
	metas := make([]PluginMeta, 0)
	for name, p := range r.resourcePlugins {
		if r.isPluginEnabled(name) {
			metas = append(metas, p.PluginMeta())
		}
	}
	return metas
}

// ListTriggers 列出所有启用的触发器插件
func (r *Registry) ListTriggers() []PluginMeta {
	metas := make([]PluginMeta, 0)
	for name, p := range r.triggerPlugins {
		if r.isPluginEnabled(name) {
			metas = append(metas, p.PluginMeta())
		}
	}
	return metas
}