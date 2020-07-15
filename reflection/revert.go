package reflection

import (
	"reflect"

	"github.com/fairyhunter13/reflecthelper"
)

// RevertVal2Zero reverts the reflect.Value to the Zero value of it's type.
func RevertVal2Zero(val reflect.Value) {
	if !val.IsValid() {
		return
	}

	kind := reflecthelper.GetKind(val)
	elemKind := kind
	elemKind = reflecthelper.GetElemKind(val)

	for kind == reflect.Ptr && elemKind == reflect.Ptr {
		val = val.Elem()

		kind = reflecthelper.GetKind(val)
		elemKind = reflecthelper.GetElemKind(val)
	}

	if reflecthelper.IsValueZero(val) {
		reflecthelper.SetReflectZero(val)
	}
}
