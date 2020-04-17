package xorm

import (
	"sync"

	"github.com/cespare/xxhash"
	"github.com/fairyhunter13/xorm/core"
	"github.com/fairyhunter13/xorm/lexer/hashkey"
)

func newStatementCache() *StatementCache {
	return &StatementCache{
		mapping: make(map[uint64]map[*core.DB]*core.Stmt),
		mutex:   new(sync.RWMutex),
	}
}

// StatementCache provides mechanism to map statement to db and query.
type StatementCache struct {
	mapping map[uint64]map[*core.DB]*core.Stmt
	mutex   *sync.RWMutex
}

func (sc *StatementCache) getDBMap(key uint64) (dbMap map[*core.DB]*core.Stmt) {
	var (
		ok bool
	)
	sc.mutex.RLock()
	dbMap, ok = sc.mapping[key]
	sc.mutex.RUnlock()
	if !ok {
		dbMap = make(map[*core.DB]*core.Stmt)
		sc.mutex.Lock()
		sc.mapping[key] = dbMap
		sc.mutex.Unlock()
	}
	return
}

// Get return the statement based on the hash key and db.
func (sc *StatementCache) Get(key uint64, db *core.DB) (stmt *core.Stmt, has bool) {
	dbMap := sc.getDBMap(key)
	sc.mutex.RLock()
	stmt, has = dbMap[db]
	sc.mutex.RUnlock()
	return
}

// Set sets the statement based on the hash key and the db.
func (sc *StatementCache) Set(key uint64, db *core.DB, stmt *core.Stmt) {
	dbMap := sc.getDBMap(key)
	sc.mutex.Lock()
	dbMap[db] = stmt
	sc.mutex.Unlock()
}

var (
	stmtCache = newStatementCache()
)

func (session *Session) doPrepare(db *core.DB, sqlStr string) (stmt *core.Stmt, err error) {
	xxh := xxhash.Sum64String(hashkey.Get(sqlStr))
	var has bool
	stmt, has = stmtCache.Get(xxh, db)
	if !has {
		stmt, err = db.PrepareContext(session.ctx, sqlStr)
		if err != nil {
			return nil, err
		}
		stmtCache.Set(xxh, db, stmt)
	}
	return
}
