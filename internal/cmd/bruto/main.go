package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dullgiulio/bruto"
	"github.com/dullgiulio/bruto/backend"
	"github.com/dullgiulio/bruto/backend/typo3"
)

type beType string

const (
	beTypeTypo3 beType = "typo3"
	//beTypeGeneric    = "generic"
)

func (t *beType) Set(v string) error {
	switch beType(v) {
	case beTypeTypo3:
		*t = beTypeTypo3
	default:
		return fmt.Errorf("Invalid backend type %s", v)
	}
	return nil
}

func (t *beType) String() string {
	return string(*t)
}

func (t *beType) backend() bruto.Backend {
	switch *t {
	case beTypeTypo3:
		return typo3.New()
	}
	return nil // return generic.New()
}

var Usage = func() {
	fmt.Fprintln(os.Stderr, "Usage: bruto [OPTIONS...] HOST")
	flag.PrintDefaults()
}

func main() {
	var be beType
	flag.Var(&be, "type", "Type of backend to use")
	flag.DurationVar(&backend.Config.Timeout, "timeout", 10*time.Second, "Timeout when performing HTTP requests")
	flag.Parse()
	host := flag.Arg(0)
	if host == "" {
		Usage()
		os.Exit(1)
	}
	runner := bruto.NewRunner(be.backend, host)
	runner.Run(os.Stdout, 1)
}
