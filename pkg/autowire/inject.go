package autowire

import (
	"fmt"
	"maps"
	"strings"
)

func Check() {
	lock.Lock()
	defer lock.Unlock()
	if isChecked {
		// 已经进行过一次完整检查，后续再次调用 Check 即使传入 handler 也不再生效
		return
	}

	doCheck(CreateHandler(func(v PreComplete) { v.OnPreComplete() }), CreateHandler(func(v PostComplete) { v.OnPostComplete() }))

}
func CheckWithEvent(handlers ...Handler) {
	lock.Lock()
	defer lock.Unlock()
	if isChecked {
		// 已经进行过一次完整检查，后续再次调用 Check 即使传入 handler 也不再生效
		return
	}

	doCheck(handlers...)

}
func doCheck(handlers ...Handler) {

	if len(requireBy) == 0 {
		return
	}
	for {
		temp := maps.Clone(requireBy)
		autoCreate(temp)
		if len(temp) == len(requireBy) { // 意味着没有新增依赖
			break
		}
	}
	// 在依赖注入前调用 OnInitialized
	beanOnInitialized()
	lack := make(map[string][]string)
	for beanName, reqs := range requireBy {
		b, has := beans[beanName]
		if !has {
			for _, req := range reqs {
				if req.root != nil {
					lack[beanName] = append(lack[beanName], fmt.Sprintf("%s.%s", req.root.name, strings.Join(req.path, ".")))
				} else {
					lack[beanName] = append(lack[beanName], "autowired")
				}
			}
			continue
		}
		for _, req := range reqs {

			setValue(req.value, b.v, req.isExported)
			if req.root != nil {
				req.root.requireCount--
			}
		}
	}

	// 检查是否有未找到的bean
	if len(lack) > 0 {
		panicLack(lack)
		return
	}

	// 依赖注入完成后，先收集当前本次可完成的 bean
	type beanItem struct {
		name string
		b    *bean
	}
	items := make([]beanItem, 0, len(beans))
	for name, b := range beans {
		if b.requireCount == 0 && !b.isComplete {
			items = append(items, beanItem{name: name, b: b})
		}
	}

	// 先对所有 bean 依次执行所有 handler
	for _, h := range handlers {
		for _, item := range items {
			h.Handle(item.name, item.b.v.Interface())
		}
	}

	// 最后对所有 bean 执行 OnComplete
	for _, item := range items {
		if ch, ok := item.b.v.Interface().(Complete); ok {
			ch.OnComplete()
		}
		item.b.isComplete = true
	}

	isChecked = true
	//清空已经注入的依赖关系表
	requireBy = make(map[string][]*requireItem)
}

func autoCreate(reqs map[string][]*requireItem) {
	for beanName := range reqs {
		if !tryAuto(beanName) {
			panic("can not find bean:" + beanName)
		}
	}
}

func beanOnInitialized() {
	for _, bv := range beans {
		if ih, ok := bv.v.Interface().(OnInitialized); ok {
			ih.Initialized()
		}
	}
}
func panicLack(lack map[string][]string) {
	b := &strings.Builder{}
	for beanName, reqs := range lack {
		_, _ = fmt.Fprintf(b, "not found bean [%s], require by [%s] ", beanName, strings.Join(reqs, ","))
	}
	panic(b.String())
}
