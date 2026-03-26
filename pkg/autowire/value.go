package autowire

import (
	"reflect"
	"unsafe"
)

func setFieldValue(field, value reflect.Value, exported bool) {
	if exported {

		field.Set(value)

	} else {
		//nolint:gosec
		reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
			Elem().
			Set(value)
	}

}
func setValue(target, source reflect.Value, exported bool) {
	target = setEmpty(target)
	tt := target.Type()
	ett := elemType(tt)
	switch tt.Kind() {
	case reflect.Ptr:
		if source.Kind() == reflect.Ptr {
			setFieldValue(target, source, exported)
			return
		}
		setFieldValue(target, reflect.New(ett), exported)
		target.Elem().Set(source)
		return
	case reflect.Interface:
		st := source.Type()
		//if st.Kind() == reflect.Ptr {
		/// bean 必须是指针
		for !st.Implements(tt) && st.Kind() == reflect.Ptr {
			st = st.Elem()
			source = source.Elem()
		}
		setFieldValue(target, source, exported)
		return
		//}
	default:

	}

	if tt.Kind() == reflect.Interface {
		st := source.Type()
		for !st.Implements(tt) {
			st = st.Elem()
			source = source.Elem()
		}
		setFieldValue(target, source, exported)
		return
	}
	for source.Kind() == reflect.Ptr {
		source = source.Elem()
	}
	setFieldValue(target, source, exported)

}
