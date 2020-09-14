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
	elemKind := reflecthelper.GetElemKind(val)
	originVal := val

	for isValueElemable(kind) && isValueElemable(elemKind) {
		val = val.Elem()

		kind = reflecthelper.GetKind(val)
		elemKind = reflecthelper.GetElemKind(val)
	}

	if reflecthelper.IsValueZero(val) {
		if !originVal.CanSet() && isValueElemable(reflecthelper.GetKind(originVal)) {
			originVal = originVal.Elem()
		}
		reflecthelper.SetReflectZero(originVal)
	}
}

func isValueElemable(kind reflect.Kind) bool {
	return kind == reflect.Ptr || kind == reflect.Interface
}
