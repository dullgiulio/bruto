package gen

import "fmt"

type Login struct {
	User string
	Pass string
}

func (l *Login) String() string {
	return fmt.Sprintf("%s %s", l.User, l.Pass)
}

type Logins struct {
	ch        chan Login
	usernames Filelines
	passwords Filelines
}

func NewLogins() *Logins {
	return &Logins{
		usernames: MakeFilelines(),
		passwords: MakeFilelines(),
		ch:        make(chan Login),
	}
}

func (l *Logins) Load(users, pass string) error {
	if err := l.usernames.Load(users); err != nil {
		return err
	}
	return l.passwords.Load(pass)
}

func (l *Logins) Chan() <-chan Login {
	return l.ch
}

func (l *Logins) Generate() {
	// Try each a new username as soon as possible as they might not exist
	// and they are probably less than the passwords.
	for p := range l.passwords {
		for u := range l.usernames {
			l.ch <- Login{
				User: l.usernames[u],
				Pass: l.passwords[p],
			}
		}
	}
	close(l.ch)
}
