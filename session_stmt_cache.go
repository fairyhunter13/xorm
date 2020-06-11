package xorm

import (
	"sync"

	"github.com/cespare/xxhash"
	"github.com/fairyhunter13/xorm/core"
	"github.com/fairyhunter13/xorm/lexer/hashkey"
)

var (
	stmtCache = make(map[uint64]map[*core.DB]*core.Stmt, 0)
	mutex     = new(sync.RWMutex)
)

func (session *Session) doPrepare(db *core.DB, sqlStr string) (stmt *core.Stmt, err error) {
	xxh := xxhash.Sum64String(hashkey.Get(sqlStr))
	var has bool
	mutex.RLock()
	stmt, has = stmtCache[xxh][db]
	mutex.RUnlock()
	if !has {
		stmt, err = db.PrepareContext(session.ctx, sqlStr)
		if err != nil {
			return nil, err
		}
		mutex.Lock()
		stmtCache[xxh][db] = stmt
		mutex.Unlock()
	}
	return
}
