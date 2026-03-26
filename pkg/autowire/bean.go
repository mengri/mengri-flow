package autowire

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	AutowiredTag = "autowired"
)

type bean struct {
	v            reflect.Value
	name         string
	isComplete   bool
	requireCount int
}

func newBean(v reflect.Value, t reflect.Type) *bean {
	b := &bean{
		v:            v,
		name:         reflectTypeName(t),
		isComplete:   false,
		requireCount: 0,
	}
	b.init()
	return b
}

type requireItem struct {
	value      reflect.Value
	isExported bool
	root       *bean
	path       []string
}

func (b *bean) init() {

	v := elemValue(b.v)
	t := elemType(v.Type())
	if t.Kind() != reflect.Struct {
		// 非 struct 类型 不需要初始化
		return
		//panic(fmt.Sprintf("just support inject struct but get :%s of %s.%s", t.Kind().String(), t.PkgPath(), t.Name()))
	}
	num := t.NumField()
	vStruct := v.Elem()
	for i := 0; i < num; i++ {
		f := t.Field(i)
		b.initField(nil, vStruct.Field(i), f)
	}

}
func (b *bean) initField(path []string, v reflect.Value, field reflect.StructField) {

	path = append(path, field.Name)
	tag, has := field.Tag.Lookup(AutowiredTag)

	if !has {

		// 只对非指针类型的子字段执行内部注入
		if v.Kind() == reflect.Struct {
			t := elemType(v.Type())

			for i := 0; i < t.NumField(); i++ {
				b.initField(path, v.Field(i), t.Field(i))
			}
		}
		return
	}
	var beanName = strings.Split(tag, ",")[0]
	if tag == "" {
		t := elemType(field.Type)
		if t.PkgPath() == "" {
			panic(fmt.Sprintf("anonymous autowired not support field [%s:%s] type [%s] ", b.name, strings.Join(path, "."), t.Name()))
		}
		beanName = reflectTypeName(t)
	}

	b.requireCount++
	requireBy[beanName] = append(requireBy[beanName], &requireItem{
		value:      v,
		root:       b,
		path:       path,
		isExported: field.IsExported(),
	})

}
