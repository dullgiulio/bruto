package bruto

import "time"

// pool is a collection of session; time represents the ready() state
type pool map[*Session]time.Time

// newPool allocates a pool
func newPool() pool {
	return pool(make(map[*Session]time.Time))
}

// add adds a session and marks it as ready
func (p pool) add(ses *Session) {
	p[ses] = time.Now()
}

// del removes a session rom the pool
func (p pool) del(ses *sessionError) {
	delete(p, ses.s)
}

// alive returns true if there still are running sessions
func (p pool) alive() bool {
	return len(p) > 0
}
