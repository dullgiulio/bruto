package bruto

import (
	"errors"

	"github.com/dullgiulio/bruto/backend"
	"github.com/dullgiulio/bruto/gen"
)

// Backend is the implementation of a worker that can try
// username/passwords pair against a system.
type Backend interface {
	// Setup initializes a backend with an unconnected HTTP backend
	Setup(domain string, conn *backend.HTTP)
	// Open opens a connection to a server with a given connection
	Open(conn *backend.HTTP) error
	// Try tries a pair of username/password to the server connected with conn
	Try(conn *backend.HTTP, l gen.Login) (success bool, err error)
}

// errSessionOver is returned when the HTTP session is terminated
var errSessionOver = errors.New("Session has terminated")

// errSessionReady is returned when the HTTP session is ready to accept pairs
var errSessionReady = errors.New("Session has started")

// errSessionAttempt is returned when
var errSessionAttempt = errors.New("Login attempt")

// sessionError wraps an error with the session that generated it
type sessionError struct {
	s   *Session
	err error
}

// newSessionError makes a sessionError
func newSessionError(s *Session, err error) *sessionError {
	return &sessionError{s: s, err: err}
}

// ready is true if the error is a ready error
func (s *sessionError) ready() bool {
	return s.err == errSessionReady
}

// attempt is true if the error is an attempt error
func (s *sessionError) attempt() bool {
	return s.err == errSessionAttempt
}

// finished is true if the error is a finished error
func (s *sessionError) finished() bool {
	return s.err == errSessionOver
}

// Error returns the underlying error
func (s *sessionError) Error() string {
	return s.err.Error()
}

// Session represents a connected backend worker
type Session struct {
	// Shared HTTP client
	conn *backend.HTTP
	// Actual implement  Backend
	be Backend
	// ready is closed when the session is ready, otherwise
	// an error is sent then the channel is closed.
	sessions chan<- error
	// Generator of user agent strings
	agents <-chan string
	// password channel generates passwords to try
	logins <-chan gen.Login
	// broken is the channel where sucessful logins are sent
	broken chan<- gen.Login
}

// newSession allocates a session with some shared channels
func newSession(be Backend, domain string, sessions chan<- error, logins <-chan gen.Login, agents <-chan string, broken chan<- gen.Login) *Session {
	s := &Session{
		conn:     backend.NewHTTP(),
		be:       be,
		sessions: sessions,
		logins:   logins,
		agents:   agents,
		broken:   broken,
	}
	s.be.Setup(domain, s.conn)
	return s
}

// ready signals the sessions handler that this session is ready to start trying pairs
func (s *Session) ready() {
	s.sessions <- newSessionError(s, errSessionReady)
}

// fail signals the session handler that this session has a fatal error and must be terminated
func (s *Session) fail(err error) error {
	s.sessions <- newSessionError(s, err)
	return err
}

// init initializes a session and opens the backend
func (s *Session) init() error {
	if err := s.conn.Init(); err != nil {
		return err
	}
	// Set a random user-agent for this session
	s.conn.Header.Set("User-Agent", <-s.agents)
	return s.be.Open(s.conn)
}

// run initializes a session and handles it's tries until it exhaustes the pairs
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
