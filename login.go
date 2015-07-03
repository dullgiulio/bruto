package bruto

import (
	"bufio"
	"fmt"
	"os"
)

type login struct {
	user string
	pass string
}

func (l *login) String() string {
	return fmt.Sprintf("%s %s", l.user, l.pass)
}

type filelines []string

func makeFilelines() filelines {
	return filelines(make([]string, 0))
}

func (f *filelines) add(s string) {
	*f = append(*f, s)
}

func (f *filelines) load(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		f.add(scanner.Text())
	}
	return scanner.Err()
}

type logins struct{
	ch chan login
	usernames filelines
	passwords filelines
}

func makeLogins() *logins {
	return &logins{
		usernames: makeFilelines(),
		passwords: makeFilelines(),
		ch: make(chan login),
	}
}

func (l *logins) generate() {
	tot := len(l.usernames) * len(l.passwords)
	for i := 0; i < tot; i++ {
		l.ch <- login{
			user: l.usernames[i%len(l.usernames)],
			pass: l.passwords[i%len(l.passwords)],
		}
	}
	close(l.ch)
}
