package xorm

import (
	"strings"
	"sync"

	"github.com/cespare/xxhash"
	"github.com/fairyhunter13/xorm/core"
)

var (
	stmtCache = make(map[uint64]*core.Stmt, 0) //key: xxhash of sanitized sqlstring
	mutex     = new(sync.RWMutex)
)

func getKey(sqlStr string) string {
	return strings.Join(strings.Fields(sqlStr), "")
}

func (session *Session) doPrepare(db *core.DB, sqlStr string) (stmt *core.Stmt, err error) {
	xxh := xxhash.Sum64String(getKey(sqlStr))
	var has bool
	mutex.RLock()
	stmt, has = stmtCache[xxh]
	mutex.RUnlock()
	if !has {
		stmt, err = db.PrepareContext(session.ctx, sqlStr)
		if err != nil {
			return nil, err
		}
		mutex.Lock()
		stmtCache[xxh] = stmt
		mutex.Unlock()
	}
	return
}
