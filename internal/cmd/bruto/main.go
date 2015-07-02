package main

import (
	"flag"
	"fmt"
	"os"

    "github.com/dullgiulio/bruto"
)

func main() {
	flag.Parse()
	host := flag.Arg(0)
	if host == "" {
		fmt.Fprintf(os.Stderr, "Usage: bruto HOST\n")
		os.Exit(1)
	}
	runner := bruto.NewRunner(host)
	runner.Run(os.Stdout, 1)
}
