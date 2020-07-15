package statements

import (
	"reflect"

	"github.com/fairyhunter13/reflecthelper"
	"github.com/fairyhunter13/xorm/bits"
	"github.com/fairyhunter13/xorm/convert"
)

// GetConversion is a function to assign val from the convert.To interface to extract the data.
// The data then consumed for the database's input and processing.
// This function also handles the reduplication of extracting the data from the statement and the update.
func GetConversion(fieldValue reflect.Value, requiredField bool, val *interface{}) (isAppend bool, isContinue bool, err error) {
	if val == nil {
		return
	}

	typeField := fieldValue.Type()
	if typeField.Kind() == reflect.Ptr {
		typeField = typeField.Elem()
	}
	if _, valid := reflect.New(typeField).Interface().(convert.To); !valid {
		return
	}

	var copyVal reflect.Value
	if fieldValue.IsZero() {
		copyVal = reflect.New(typeField)
	} else {
		copyVal = reflect.ValueOf(fieldValue.Interface())
	}

	copyVal = reflecthelper.GetInitElem(copyVal)
	if copyVal.CanAddr() {
		copyVal = copyVal.Addr()
	}

	isAppend, err = getDataConversion(copyVal, requiredField, val)
	if err != nil {
		return
	}
	if !isAppend {
		isContinue = true
	}
	return
}

func getDataConversion(fieldValue reflect.Value, requiredField bool, val *interface{}) (isAppend bool, err error) {
	if structConvert, ok := fieldValue.Interface().(convert.To); ok {
		if reflecthelper.IsZero(structConvert) {
			if requiredField {
				isAppend = true
			}
			return
		}
		var data []byte
		data, err = structConvert.ToDB()
		if err != nil {
			return
		}

		if requiredField && bits.IsEmpty(data) {
			isAppend = true
			return
		}

		*val = data
		isAppend = true
	}
	return
}
