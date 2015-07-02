package main

type login struct {
	user string
	pass string
}

type logins chan login

var usernames = []string{"giulio.iotti", "giotti", "iotti.giulio"}
var passwords = []string{"password", "12345", "test123", "hello321"}

func (l logins) generate() {
	tot := len(usernames) * len(passwords)
	for i := 0; i < tot; i++ {
		l <- login{
			user: usernames[i%len(usernames)],
			pass: passwords[i%len(passwords)],
		}
	}
	close(l)
}
