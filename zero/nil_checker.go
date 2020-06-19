package zero

import "reflect"

// IsNil checks if the payload i is nil.
// IsNil use reflect if it's needed.
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Ptr, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
