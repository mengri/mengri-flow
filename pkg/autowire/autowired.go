package autowire

import (
	"log"
	"reflect"
)

// Autowired 自动注入
// 1. 如果传入值为nil, 则整体注入
// 2. 如果传入值不为nil, 则按内部字段注入
// 3. 如果传入值为nil, 且beanName为空, 则不支持基础类型匿名注入
// 4. 如果传入值为nil, 且beanName不为空, 则整体注入
func Autowired[T any](v T, name ...string) {

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Interface {
		log.Fatal("only autowired ptr but get:", rv.Kind().String())
	}
	if rv.IsNil() {
		log.Fatal("not allow autowired nil ptr for:", rv.String())
	}
	lock.Lock()
	defer lock.Unlock()
	ert := elemType(reflect.TypeOf(new(T)))
	beanName := reflectTypeName(ert)
	if ert.PkgPath() == "" {
		if len(name) == 0 { // 不支持基础类型匿名注入
			log.Fatalf("not support autowire without name  for %s", ert.Name())
			return
		}
		beanName = name[0]
	}
	rv = setEmpty(rv)
	if rv.IsNil() { // 如果传入值为nil, 则整体注入

		bv, has := beans[beanName]
		if has {
			// 存在则整体注入

			setValue(bv.v, rv, true)
		} else {

			// 不存在则缓存起来,时机合适再注入
			requireBy[beanName] = append(requireBy[beanName], &requireItem{
				value:      rv,
				path:       nil,
				root:       nil,
				isExported: true,
			})
		}

		return
	} else { // 如果传入值不为nil, 则按内部字段注入,并将其视为完整的bean
		if ert.PkgPath() == "" && len(name) == 0 { // 基础类型不支持内部字段注入
			log.Fatalf("not support autowire with name  for %s", ert.Name())
		}

		nb := newBean(rv, ert)

		if ert.PkgPath() != "" || len(name) > 0 { // 非基础类型或者有名字, 则按名字加入到bean中
			addBeans(beanName, nb) // 同名的, 第一个才会成为有效的bean
		}
		if isChecked {
			// 再次 doCheck 时不再传入 handler，只做依赖注入，不再重复执行事件
			doCheck()
		}

	}

}
func addBeans(name string, b *bean) bool {
	_, has := beans[name]
	if has {
		return false
	}
	beans[name] = b
	return true
}
