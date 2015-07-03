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
	sessions chan error
	// Signal that the login pair generator has finished
	pwdOver chan struct{}
	// Login pair generator
	logins *logins
	// Receiver for successful login attempts
	broken broken
	// Pool of session workers
	pool pool
}

func NewRunner(host string) *Runner {
	return &Runner{
		domain:   urls(host),
		sessions: make(chan error),
		pwdOver:  make(chan struct{}),
		broken:   makeBroken(),
		logins:   makeLogins(),
		pool:     newPool(),
	}
}

// Utility to create a new session
func (r *Runner) makeSession() {
	s := newSession(r.domain, r.sessions, r.logins.ch, r.broken)
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
	if err := r.logins.usernames.load("usernames.txt"); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if err := r.logins.passwords.load("passwords.txt"); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	// Generate username/password pairs and signal when there are no more
	go r.generateLogins()
	// Print broken login pairs to stdout
	go r.broken.writeTo(w)
	// Start some workers
	r.startWorkers(workers)
	for {
		select {
		case s := <-r.sessions:
			if _, ok := s.(*sessionError); !ok {
				fmt.Printf("Error: %s\n", s)
				break
			}
			se := s.(*sessionError)
			if !se.fatal() {
				break
			}
			// TODO: Detect the error rate here. If high, don't start new workes, exit.
			if !se.finished() {
				fmt.Printf("Error: %s\n", s)
			}
			// Remove a worker from the pool if it had an error and it's dead
			r.pool.del(se)
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
