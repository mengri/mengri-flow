package autowire

import "reflect"

func setEmpty(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}
	t := v.Type()
	t = t.Elem()

	if t.Kind() != reflect.Ptr && t.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		if v.Kind() == reflect.Ptr {
			v.Set(reflect.New(v.Type()))
		}
	}
	return setEmpty(v.Elem())
}
