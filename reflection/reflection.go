package reflection

import "reflect"

// GetElem is similar to reflect.Indirect but initialize a new empty type.
// It initializes when the kind of the type is reflect.Ptr and the pointer is nil.
func GetElem(v *reflect.Value) *reflect.Value {
	kind := v.Type().Kind()
	if kind != reflect.Ptr {
		return v
	}
	if v.CanAddr() && v.IsNil() {
		v.Set(reflect.New(v.Type().Elem()))
	}

	elem := reflect.Indirect(*v)
	return &elem
}
