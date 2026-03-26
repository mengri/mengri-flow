package autowire

import (
	"log"
	"reflect"
)

type factoryItem struct {
	fh          func() reflect.Value
	et          reflect.Type
	factoryName string
	names       []string
}

// Auto   auto register a factory
// handler: a function that returns a new instance of the type T
// otherNames: other names that can be used to refer to the factory
// if otherNames is empty, the type name of T will be used
// if otherNames is not empty, the type name of T will not be used
func Auto[T any](handler func() T, otherNames ...string) {
	lock.Lock()
	defer lock.Unlock()
	
	if isChecked {
		log.Fatalf("autowire.Auto can not be called after autowire.Check")
	}
	var names []string
	if len(otherNames) == 0 {
		names = []string{typeName[T]()}
	} else {
		names = otherNames
	}
	ft := reflect.TypeOf(handler)
	it := &factoryItem{
		fh: func() reflect.Value {
			return reflect.ValueOf(handler())
		},
		et:          elemType(reflect.TypeOf(new(T))),
		factoryName: reflectTypeName(ft),
		names:       names,
	}

	for _, n := range names {
		if !addFactory(n, it) {
			log.Fatalf("duplicate factory name %s", n)
			return
		}
	}

}

func addFactory(name string, it *factoryItem) bool {

	if _, ok := factories[name]; ok {
		return false
	}
	factories[name] = it
	return true
}
func tryAuto(beanName string) bool {
	if _, ok := beans[beanName]; ok {
		return true
	}

	if factory, ok := factories[beanName]; ok {
		bv := factory.fh()

		if c, y := bv.Interface().(OnCreate); y {
			c.OnCreate()
		}

		bi := newBean(bv, factory.et)

		for _, n := range factory.names {
			if !addBeans(n, bi) {
				if len(factory.names) > 1 {
					continue
				}
				log.Fatalf("duplicate bean name %s by factory %s ", n, beanName)
			}
		}

		return true
	}
	return false
}
