// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"fmt"
	"reflect"
	"strings"

	"xorm.io/core"
)

// tbNameWithSchema will automatically add schema prefix on table name
func (engine *Engine) tbNameWithSchema(v string) string {
	// Add schema name as prefix of table name.
	// Only for postgres database.
	if engine.dialect.DBType() == core.POSTGRES &&
		engine.dialect.URI().Schema != "" &&
		engine.dialect.URI().Schema != postgresPublicSchema &&
		strings.Index(v, ".") == -1 {
		return engine.dialect.URI().Schema + "." + v
	}
	return v
}

// TableName returns table name with schema prefix if has
func (engine *Engine) TableName(bean interface{}, includeSchema ...bool) string {
	tbName, _ := newTableName(engine.TableMapper, bean)
	if len(includeSchema) > 0 && includeSchema[0] {
		tbName.schema = engine.dialect.URI().Schema
		return tbName.withSchema()
	}

	return tbName.withNoSchema()
}

// tbName get some table's table name
func (session *Session) tbNameNoSchema(table *core.Table) string {
	if len(session.statement.altTableName) > 0 {
		return session.statement.altTableName
	}

	return table.Name
}

func tbNameForMap(mapper core.IMapper, v reflect.Value) string {
	if t, ok := v.Interface().(TableName); ok {
		return t.TableName()
	}
	if v.Type().Implements(tpTableName) {
		return v.Interface().(TableName).TableName()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		if t, ok := v.Interface().(TableName); ok {
			return t.TableName()
		}
		if v.Type().Implements(tpTableName) {
			return v.Interface().(TableName).TableName()
		}
	}

	return mapper.Obj2Table(v.Type().Name())
}

type tableName struct {
	name          string
	schema        string
	alias         string
	aliasSplitter string
}

func newTableName(mapper core.IMapper, tablename interface{}) (tableName, error) {
	switch tablename.(type) {
	case []string:
		t := tablename.([]string)
		if len(t) > 1 {
			return tableName{name: t[0], alias: t[1]}, nil
		} else if len(t) == 1 {
			return tableName{name: t[0]}, nil
		}
		return tableName{}, ErrTableNotFound
	case []interface{}:
		t := tablename.([]interface{})
		l := len(t)
		var table string
		if l > 0 {
			f := t[0]
			switch f.(type) {
			case string:
				table = f.(string)
			case TableName:
				table = f.(TableName).TableName()
			default:
				v := rValue(f)
				t := v.Type()
				if t.Kind() == reflect.Struct {
					table = tbNameForMap(mapper, v)
				} else {
					table = fmt.Sprintf("%v", f)
				}
			}
		}
		if l > 1 {
			return tableName{name: table, alias: fmt.Sprintf("%v", t[1])}, nil
		} else if l == 1 {
			return tableName{name: table}, nil
		}
	case TableName:
		fmt.Println("+++++++++++++++++++++++++", tablename.(TableName).TableName())
		return tableName{name: tablename.(TableName).TableName()}, nil
	case string:
		return tableName{name: tablename.(string)}, nil
	case reflect.Value:
		v := tablename.(reflect.Value)
		return tableName{name: tbNameForMap(mapper, v)}, nil
	default:
		v := rValue(tablename)
		t := v.Type()
		if t.Kind() == reflect.Struct {
			return tableName{name: tbNameForMap(mapper, v)}, nil
		}
		return tableName{name: fmt.Sprintf("%v", tablename)}, nil
	}
	return tableName{}, ErrTableNotFound
}

func (t tableName) withSchema() string {
	if t.schema == "" {
		return t.withNoSchema()
	}

	if t.alias != "" {
		if t.aliasSplitter != "" {
			return fmt.Sprintf("%s.%s %s %s", t.schema, t.name, t.aliasSplitter, t.alias)
		}
		return fmt.Sprintf("%s.%s %s", t.schema, t.name, t.alias)
	}
	return fmt.Sprintf("%s.%s", t.schema, t.name)
}

func (t tableName) withNoSchema() string {
	if t.alias != "" {
		if t.aliasSplitter != "" {
			return fmt.Sprintf("%s %s %s", t.name, t.aliasSplitter, t.alias)
		}
		return fmt.Sprintf("%s %s", t.name, t.alias)
	}
	return t.name
}
