package xorm

import "github.com/panjf2000/ants/v2"

var sessionPool = newSessionPool()

// SessionPool is a cutomized pooling for session.
type SessionPool struct {
	sessionChan chan *Session
}

func newSessionPool() *SessionPool {
	return &SessionPool{
		sessionChan: make(chan *Session),
	}
}

// Get get the session from the session pool.
func (sp *SessionPool) Get() (session *Session) {
	select {
	case session = <-sp.sessionChan:
	default:
		session = new(Session)
	}
	return
}

// Put puts the session into the pool.
func (sp *SessionPool) Put(session *Session) {
	ants.Submit(func() {
		sp.sessionChan <- session
	})
}
