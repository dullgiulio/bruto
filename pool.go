package bruto

import "time"

// time represents the ready() state
type pool map[*Session]time.Time

func newPool() pool {
	return pool(make(map[*Session]time.Time))
}

func (p pool) add(ses *Session) {
	p[ses] = time.Now()
}

func (p pool) del(ses *sessionError) {
	delete(p, ses.s)
}

func (p pool) alive() bool {
	return len(p) > 0
}
