package autowire

type OnCreate interface {
	OnCreate()
}
type OnInitialized interface { //初始化完成
	Initialized()
}
type Complete interface { // 注入完成
	OnComplete()
}
type PreComplete interface {
	OnPreComplete()
}

type PostComplete interface {
	OnPostComplete()
}

// Handler 用于在 Check 成功后，按用户自定义逻辑对 bean 执行事件。
// name 为 bean 名称，bean 为注入容器中的实例。
type Handler interface {
	Handle(name string, bean any)
}

// HandlerFunc 适配器，便于直接使用函数。
type HandlerFunc func(name string, bean any)

func (f HandlerFunc) Handle(name string, bean any) {
	f(name, bean)
}

func CreateHandler[T any](h func(T)) HandlerFunc {
	return func(name string, bean any) {

		if eh, ok := bean.(T); ok {
			h(eh)
		}
	}
}
