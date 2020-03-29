// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"github.com/fairyhunter13/xorm/log"
	"github.com/fairyhunter13/xorm/names"
)

var (
	// DefaultCacheSize sets the default cache size
	DefaultCacheSize = 200
)

func MapToSlice(query string, mp interface{}) (string, []interface{}, error) {
	vv := reflect.ValueOf(mp)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Map {
		return "", []interface{}{}, ErrNoMapPointer
	}

	args := make([]interface{}, 0, len(vv.Elem().MapKeys()))
	var err error
	query = re.ReplaceAllStringFunc(query, func(src string) string {
		v := vv.Elem().MapIndex(reflect.ValueOf(src[1:]))
		if !v.IsValid() {
			err = fmt.Errorf("map key %s is missing", src[1:])
		} else {
			args = append(args, v.Interface())
		}
		return "?"
	})

	return query, args, err
}

func StructToSlice(query string, st interface{}) (string, []interface{}, error) {
	vv := reflect.ValueOf(st)
	if vv.Kind() != reflect.Ptr || vv.Elem().Kind() != reflect.Struct {
		return "", []interface{}{}, ErrNoStructPointer
	}

	args := make([]interface{}, 0)
	var err error
	query = re.ReplaceAllStringFunc(query, func(src string) string {
		fv := vv.Elem().FieldByName(src[1:]).Interface()
		if v, ok := fv.(driver.Valuer); ok {
			var value driver.Value
			value, err = v.Value()
			if err != nil {
				return "?"
			}
			args = append(args, value)
		} else {
			args = append(args, fv)
		}
		return "?"
	})
	if err != nil {
		return "", []interface{}{}, err
	}
	return query, args, nil
}

type cacheStruct struct {
	value reflect.Value
	idx   int
}

var (
	_ QueryExecuter = &DB{}
)

// DB is a wrap of sql.DB with extra contents
type DB struct {
	*sql.DB
	Mapper            names.Mapper
	reflectCache      map[reflect.Type]*cacheStruct
	reflectCacheMutex sync.RWMutex
	Logger            log.ContextLogger
	hooks             contexts.Hooks
}

var (
	// ErrInvalidTypeOfSource indicates that the source passed to the NewEngine function
	// is invalid.
	ErrInvalidTypeOfSource = errors.New("Source type of the engine must be valid string, *string, or *sql.DB")
	// ErrSourceMusntBeEmpty indicates that the passed data source is either empty or nil.
	ErrSourceMusntBeEmpty = errors.New("The source musn't be empty or nil")
)

// Open opens a database
func Open(driverName string, dataSourceName interface{}) (*DB, error) {
	var (
		db  *sql.DB
		err error
	)
	switch dsn := dataSourceName.(type) {
	case string:
		db, err = sql.Open(driverName, dsn)
		if err != nil {
			return nil, err
		}
	case *sql.DB:
		if dsn == nil {
			return nil, ErrSourceMusntBeEmpty
		}
		db = dsn
	default:
		return nil, ErrInvalidTypeOfSource
	}
	return &DB{
		DB:           db,
		Mapper:       names.NewCacheMapper(&names.SnakeMapper{}),
		reflectCache: make(map[reflect.Type]*cacheStruct),
	}, nil
}

// FromDB creates a DB from a sql.DB
func FromDB(db *sql.DB) *DB {
	return &DB{
		DB:           db,
		Mapper:       names.NewCacheMapper(&names.SnakeMapper{}),
		reflectCache: make(map[reflect.Type]*cacheStruct),
	}
}

// NeedLogSQL returns true if need to log SQL
func (db *DB) NeedLogSQL(ctx context.Context) bool {
	if db.Logger == nil {
		return false
	}

	v := ctx.Value(log.SessionShowSQLKey)
	if showSQL, ok := v.(bool); ok {
		return showSQL
	}
	return db.Logger.IsShowSQL()
}

func (db *DB) reflectNew(typ reflect.Type) reflect.Value {
	db.reflectCacheMutex.Lock()
	defer db.reflectCacheMutex.Unlock()
	cs, ok := db.reflectCache[typ]
	if !ok || cs.idx+1 > DefaultCacheSize-1 {
		cs = &cacheStruct{reflect.MakeSlice(reflect.SliceOf(typ), DefaultCacheSize, DefaultCacheSize), 0}
		db.reflectCache[typ] = cs
	} else {
		cs.idx = cs.idx + 1
	}
	return cs.value.Index(cs.idx).Addr()
}

// QueryContext overwrites sql.DB.QueryContext
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*Rows, error) {
	hookCtx := contexts.NewContextHook(ctx, query, args)
	ctx, err := db.beforeProcess(hookCtx)
	if err != nil {
		return nil, err
	}
	rows, err := db.DB.QueryContext(ctx, query, args...)
	hookCtx.End(ctx, nil, err)
	if err := db.afterProcess(hookCtx); err != nil {
		if rows != nil {
			rows.Close()
		}
		return nil, err
	}
	return &Rows{rows, db}, nil
}

// Query overwrites sql.DB.Query
func (db *DB) Query(query string, args ...interface{}) (*Rows, error) {
	return db.QueryContext(context.Background(), query, args...)
}

// QueryMapContext executes query with parameters via map and context
func (db *DB) QueryMapContext(ctx context.Context, query string, mp interface{}) (*Rows, error) {
	query, args, err := MapToSlice(query, mp)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

// QueryMap executes query with parameters via map
func (db *DB) QueryMap(query string, mp interface{}) (*Rows, error) {
	return db.QueryMapContext(context.Background(), query, mp)
}

func (db *DB) QueryStructContext(ctx context.Context, query string, st interface{}) (*Rows, error) {
	query, args, err := StructToSlice(query, st)
	if err != nil {
		return nil, err
	}
	return db.QueryContext(ctx, query, args...)
}

func (db *DB) QueryStruct(query string, st interface{}) (*Rows, error) {
	return db.QueryStructContext(context.Background(), query, st)
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *Row {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return &Row{nil, err}
	}
	return &Row{rows, nil}
}

func (db *DB) QueryRow(query string, args ...interface{}) *Row {
	return db.QueryRowContext(context.Background(), query, args...)
}

func (db *DB) QueryRowMapContext(ctx context.Context, query string, mp interface{}) *Row {
	query, args, err := MapToSlice(query, mp)
	if err != nil {
		return &Row{nil, err}
	}
	return db.QueryRowContext(ctx, query, args...)
}

func (db *DB) QueryRowMap(query string, mp interface{}) *Row {
	return db.QueryRowMapContext(context.Background(), query, mp)
}

func (db *DB) QueryRowStructContext(ctx context.Context, query string, st interface{}) *Row {
	query, args, err := StructToSlice(query, st)
	if err != nil {
		return &Row{nil, err}
	}
	return db.QueryRowContext(ctx, query, args...)
}

func (db *DB) QueryRowStruct(query string, st interface{}) *Row {
	return db.QueryRowStructContext(context.Background(), query, st)
}

var (
	re = regexp.MustCompile(`[?](\w+)`)
)

// ExecMapContext exec map with context.ContextHook
// insert into (name) values (?)
// insert into (name) values (?name)
func (db *DB) ExecMapContext(ctx context.Context, query string, mp interface{}) (sql.Result, error) {
	query, args, err := MapToSlice(query, mp)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

func (db *DB) ExecMap(query string, mp interface{}) (sql.Result, error) {
	return db.ExecMapContext(context.Background(), query, mp)
}

func (db *DB) ExecStructContext(ctx context.Context, query string, st interface{}) (sql.Result, error) {
	query, args, err := StructToSlice(query, st)
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, query, args...)
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	hookCtx := contexts.NewContextHook(ctx, query, args)
	ctx, err := db.beforeProcess(hookCtx)
	if err != nil {
		return nil, err
	}
	res, err := db.DB.ExecContext(ctx, query, args...)
	hookCtx.End(ctx, res, err)
	if err := db.afterProcess(hookCtx); err != nil {
		return nil, err
	}
	return res, nil
}

func (db *DB) ExecStruct(query string, st interface{}) (sql.Result, error) {
	return db.ExecStructContext(context.Background(), query, st)
}

func (db *DB) beforeProcess(c *contexts.ContextHook) (context.Context, error) {
	if db.NeedLogSQL(c.Ctx) {
		db.Logger.BeforeSQL(log.LogContext(*c))
	}
	ctx, err := db.hooks.BeforeProcess(c)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (db *DB) afterProcess(c *contexts.ContextHook) error {
	err := db.hooks.AfterProcess(c)
	if db.NeedLogSQL(c.Ctx) {
		db.Logger.AfterSQL(log.LogContext(*c))
	}
	return err
}

func (db *DB) AddHook(h ...contexts.Hook) {
	db.hooks.AddHook(h...)
}
