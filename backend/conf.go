package backend

import "time"

type Conf struct {
	Timeout time.Duration
}

var Config *Conf

func init() {
	Config = &Conf{
		Timeout: 10 * time.Second,
	}
}
