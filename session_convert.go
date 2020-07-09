// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fairyhunter13/xorm/convert"
	"github.com/fairyhunter13/xorm/internal/json"
	"github.com/fairyhunter13/xorm/internal/utils"
	"github.com/fairyhunter13/xorm/reflection"
	"github.com/fairyhunter13/xorm/schemas"
)

func (session *Session) str2Time(col *schemas.Column, data string) (outTime time.Time, outErr error) {
	sdata := strings.TrimSpace(data)
	var x time.Time
	var err error

	var parseLoc = session.engine.DatabaseTZ
	if col.TimeZone != nil {
		parseLoc = col.TimeZone
	}

	if sdata == utils.ZeroTime0 || sdata == utils.ZeroTime1 {
	} else if !strings.ContainsAny(sdata, "- :") { // !nashtsai! has only found that mymysql driver is using this for time type column
		// time stamp
		sd, err := strconv.ParseInt(sdata, 10, 64)
		if err == nil {
			x = time.Unix(sd, 0)
			//session.engine.logger.Debugf("time(0) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
		} else {
			//session.engine.logger.Debugf("time(0) err key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
		}
	} else if len(sdata) > 19 && strings.Contains(sdata, "-") {
		x, err = time.ParseInLocation(time.RFC3339Nano, sdata, parseLoc)
		session.engine.logger.Debugf("time(1) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
		if err != nil {
			x, err = time.ParseInLocation("2006-01-02 15:04:05.999999999", sdata, parseLoc)
			//session.engine.logger.Debugf("time(2) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
		}
		if err != nil {
			x, err = time.ParseInLocation("2006-01-02 15:04:05.9999999 Z07:00", sdata, parseLoc)
			//session.engine.logger.Debugf("time(3) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
		}
	} else if len(sdata) == 19 && strings.Contains(sdata, "-") {
		x, err = time.ParseInLocation("2006-01-02 15:04:05", sdata, parseLoc)
		//session.engine.logger.Debugf("time(4) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
	} else if len(sdata) == 10 && sdata[4] == '-' && sdata[7] == '-' {
		x, err = time.ParseInLocation("2006-01-02", sdata, parseLoc)
		//session.engine.logger.Debugf("time(5) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
	} else if col.SQLType.Name == schemas.Time {
		if strings.Contains(sdata, " ") {
			ssd := strings.Split(sdata, " ")
			sdata = ssd[1]
		}

		sdata = strings.TrimSpace(sdata)
		if session.engine.dialect.URI().DBType == schemas.MYSQL && len(sdata) > 8 {
			sdata = sdata[len(sdata)-8:]
		}

		st := fmt.Sprintf("2006-01-02 %v", sdata)
		x, err = time.ParseInLocation("2006-01-02 15:04:05", st, parseLoc)
		//session.engine.logger.Debugf("time(6) key[%v]: %+v | sdata: [%v]\n", col.FieldName, x, sdata)
	} else {
		outErr = fmt.Errorf("unsupported time format %v", sdata)
		return
	}
	if err != nil {
		outErr = fmt.Errorf("unsupported time format %v: %v", sdata, err)
		return
	}
	outTime = x.In(session.engine.TZLocation)
	return
}

func (session *Session) byte2Time(col *schemas.Column, data []byte) (outTime time.Time, outErr error) {
	return session.str2Time(col, string(data))
}

// convert a db data([]byte) to a field value
func (session *Session) bytes2Value(col *schemas.Column, fieldValue *reflect.Value, data []byte) error {
	if fieldValue.CanAddr() {
		if structConvert, ok := fieldValue.Addr().Interface().(convert.From); ok {
			if utils.IsZero(structConvert) {
				return nil
			}
			return structConvert.FromDB(data)
		}
	}

	if structConvert, ok := fieldValue.Interface().(convert.From); ok {
		if utils.IsZero(structConvert) {
			return nil
		}
		return structConvert.FromDB(data)
	}

	fieldValue = reflection.GetElem(fieldValue)
	var v interface{}
	key := col.Name
	fieldType := fieldValue.Type()

	switch fieldType.Kind() {
	case reflect.Complex64, reflect.Complex128:
		x := reflect.New(fieldType)
		if len(data) > 0 {
			err := json.DefaultJSONHandler.Unmarshal(data, x.Interface())
			if err != nil {
				return err
			}
			fieldValue.Set(x.Elem())
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		v = data
		t := fieldType.Elem()
		k := t.Kind()
		if col.SQLType.IsText() {
			x := reflect.New(fieldType)
			if len(data) > 0 {
				err := json.DefaultJSONHandler.Unmarshal(data, x.Interface())
				if err != nil {
					return err
				}
				fieldValue.Set(x.Elem())
			}
		} else if col.SQLType.IsBlob() {
			if k == reflect.Uint8 {
				fieldValue.Set(reflect.ValueOf(v))
			} else {
				x := reflect.New(fieldType)
				if len(data) > 0 {
					err := json.DefaultJSONHandler.Unmarshal(data, x.Interface())
					if err != nil {
						return err
					}
					fieldValue.Set(x.Elem())
				}
			}
		} else {
			return ErrUnSupportedType
		}
	case reflect.String:
		fieldValue.SetString(string(data))
	case reflect.Bool:
		v, err := asBool(data)
		if err != nil {
			return fmt.Errorf("arg %v as bool: %s", key, err.Error())
		}
		fieldValue.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sdata := string(data)
		var x int64
		var err error
		// for mysql, when use bit, it returned \x01
		if col.SQLType.Name == schemas.Bit &&
			session.engine.dialect.URI().DBType == schemas.MYSQL { // !nashtsai! TODO dialect needs to provide conversion interface API
			if len(data) == 1 {
				x = int64(data[0])
			} else {
				x = 0
			}
		} else if strings.HasPrefix(sdata, "0x") {
			x, err = strconv.ParseInt(sdata, 16, 64)
		} else if strings.HasPrefix(sdata, "0") {
			x, err = strconv.ParseInt(sdata, 8, 64)
		} else if strings.EqualFold(sdata, "true") {
			x = 1
		} else if strings.EqualFold(sdata, "false") {
			x = 0
		} else {
			x, err = strconv.ParseInt(sdata, 10, 64)
		}
		if err != nil {
			return fmt.Errorf("arg %v as int: %s", key, err.Error())
		}
		fieldValue.SetInt(x)
	case reflect.Float32, reflect.Float64:
		x, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return fmt.Errorf("arg %v as float64: %s", key, err.Error())
		}
		fieldValue.SetFloat(x)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		x, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil {
			return fmt.Errorf("arg %v as int: %s", key, err.Error())
		}
		fieldValue.SetUint(x)
	//Currently only support Time type
	case reflect.Struct:
		// !<winxxp>! 增加支持sql.Scanner接口的结构，如sql.NullString
		if nulVal, ok := fieldValue.Addr().Interface().(sql.Scanner); ok {
			if err := nulVal.Scan(data); err != nil {
				return fmt.Errorf("sql.Scan(%v) failed: %s ", data, err.Error())
			}
		} else {
			if fieldType.ConvertibleTo(schemas.TimeType) {
				x, err := session.byte2Time(col, data)
				if err != nil {
					return err
				}
				v = x
				fieldValue.Set(reflect.ValueOf(v).Convert(fieldType))
			} else if session.statement.UseCascade {
				table, err := session.engine.tagParser.ParseWithCache(*fieldValue)
				if err != nil {
					return err
				}

				// TODO: current only support 1 primary key
				if len(table.PrimaryKeys) > 1 {
					return errors.New("unsupported composited primary key cascade")
				}

				var pk = make(schemas.PK, len(table.PrimaryKeys))
				rawValueType := table.ColumnType(table.PKColumns()[0].FieldName)
				pk[0], err = str2PK(string(data), rawValueType)
				if err != nil {
					return err
				}

				if !pk.IsZero() {
					// !nashtsai! TODO for hasOne relationship, it's preferred to use join query for eager fetch
					// however, also need to consider adding a 'lazy' attribute to xorm tag which allow hasOne
					// property to be fetched lazily
					structInter := reflect.New(fieldValue.Type())
					has, err := session.ID(pk).NoCascade().get(structInter.Interface())
					if err != nil {
						return err
					}
					if has {
						v = structInter.Elem().Interface()
						fieldValue.Set(reflect.ValueOf(v))
					} else {
						return errors.New("cascade obj is not exist")
					}
				}
			}
		}
	default:
		return fmt.Errorf("unsupported type in Scan: %s", fieldValue.Type().String())
	}

	return nil
}
