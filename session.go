package bruto

import (
	"errors"

	"github.com/dullgiulio/bruto/backend"
	"github.com/dullgiulio/bruto/gen"
)

type Backend interface {
	Setup(domain string, conn *backend.HTTP)
	Open(conn *backend.HTTP) error
	Try(conn *backend.HTTP, l gen.Login) (success bool, err error)
}

var errSessionOver = errors.New("Session has terminated")
var errSessionReady = errors.New("Session has started")
var errSessionAttempt = errors.New("Login attempt")

type sessionError struct {
	s   *Session
	err error
}

func newSessionError(s *Session, err error) *sessionError {
	return &sessionError{s: s, err: err}
}

func (s *sessionError) ready() bool {
	return s.err == errSessionReady
}

func (s *sessionError) attempt() bool {
	return s.err == errSessionAttempt
}

func (s *sessionError) finished() bool {
	return s.err == errSessionOver
}

func (s *sessionError) Error() string {
	return s.err.Error()
}

type Session struct {
	// Shared HTTP client
	conn *backend.HTTP
	// Actual implement  Backend
	be Backend
	// ready is closed when the session is ready, otherwise
	// an error is sent then the channel is closed.
	sessions chan<- error
	// password channel generates passwords to try
	logins <-chan gen.Login
	// broken is the channel where sucessful logins are sent
	broken chan<- gen.Login
}

func newSession(be Backend, domain string, sessions chan<- error, logins <-chan gen.Login, broken chan<- gen.Login) *Session {
	s := &Session{
		conn:     backend.NewHTTP(),
		be:       be,
		sessions: sessions,
		logins:   logins,
		broken:   broken,
	}
	s.be.Setup(domain, s.conn)
	return s
}

func (s *Session) ready() {
	s.sessions <- newSessionError(s, errSessionReady)
}

func (s *Session) fail(err error) error {
	s.sessions <- newSessionError(s, err)
	return err
}

func (s *Session) init() error {
	if err := s.conn.Init(); err != nil {
		return err
	}
	return s.be.Open(s.conn)
}

func (s *Session) run() error {
	if err := s.init(); err != nil {
		return s.fail(err)
	}
	// Signal that this session is ready to do real work
	s.ready()
	// Fetch login pairs and try them
	for l := range s.logins {
		success, err := s.be.Try(s.conn, l)
		if err != nil {
			s.fail(err)
			continue
		}
		// This login worked, send back on the broken logins channel
		if success {
			s.broken <- l
		}
		// Signal that this attempt is over
		s.fail(errSessionAttempt)
	}
	// Signal that this session has terminated
	return s.fail(errSessionOver)
}
