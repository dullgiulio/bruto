package bruto

type pool map[*session]struct{}

func newPool() pool {
	return pool(make(map[*session]struct{}))
}

func (p pool) add(ses *session) {
	p[ses] = struct{}{}
}

func (p pool) del(ses maybeSession) {
	delete(p, ses.s)
}

func (p pool) alive() bool {
	return len(p) > 0
}
