package backend

import "time"

type Conf struct {
	Timeout time.Duration
}

var Config *Conf

func init() {
	// Relay on the flags to setup default variables.
	Config = &Conf{}
}
