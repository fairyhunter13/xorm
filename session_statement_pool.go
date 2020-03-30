package xorm

import "sync"

var sessionPool = &sync.Pool{
	New: func() interface{} {
		return new(Session)
	},
}

// Reset all the session to the initial states
func (session *Session) Reset() {
	session.engine = nil
	session.tx = nil
	session.statement = nil

	session.isAutoCommit = false
	session.isCommitedOrRollbacked = false
	session.isAutoClose = false
	session.isClosed = false
	session.prepareStmt = false
	session.autoResetStatement = false

	session.afterInsertBeans = nil
	session.afterUpdateBeans = nil
	session.afterDeleteBeans = nil

	session.beforeClosures = session.beforeClosures[:0]
	session.afterClosures = session.afterClosures[:0]
	session.afterProcessors = session.afterProcessors[:0]

	session.lastSQL = session.lastSQL[:0]
	session.lastSQLArgs = session.lastSQLArgs[:0]

	session.ctx = nil
	session.sessionType = engineSession
}
