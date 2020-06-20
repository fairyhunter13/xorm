package xorm

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
	sp.sessionChan <- session
}

// Close closes the underlying channel.
func (sp *SessionPool) Close() {
	close(sp.sessionChan)
}
