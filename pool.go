package main

type spool map[*session]struct{}

func newPool() spool {
	return spool(make(map[*session]struct{}))
}

func (s spool) add(ses *session) {
	s[ses] = struct{}{}
}

func (s spool) del(ses *session) {
	delete(s, ses)
}

func (s spool) alive() bool {
	return len(s) > 0
}
