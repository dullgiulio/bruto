package bruto

import "time"

// time represents the ready() state
type pool map[*session]time.Time

func newPool() pool {
	return pool(make(map[*session]time.Time))
}

func (p pool) add(ses *session) {
	p[ses] = time.Now()
}

func (p pool) del(ses *sessionError) {
	delete(p, ses.s)
}

func (p pool) alive() bool {
	return len(p) > 0
}
