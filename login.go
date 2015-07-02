package bruto

import "fmt"

type login struct {
	user string
	pass string
}

func (l *login) String() string {
	return fmt.Sprintf("%s %s", l.user, l.pass)
}

type logins chan login

// TODO: Get from files according to conf
var usernames = []string{"giulio.iotti", "giotti", "iotti.giulio"}
var passwords = []string{"password", "12345", "test123", "hello321"}

func makeLogins() logins {
	return logins(make(chan login))
}

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
