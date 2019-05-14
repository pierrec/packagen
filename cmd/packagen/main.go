package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pierrec/cmdflag"
)

var cli = cmdflag.New(nil)

var verbose bool

func newLogger() *log.Logger {
	return log.New(os.Stderr, "packagen", 0)
}

func main() {
	flag.BoolVar(&verbose, "v", false, "verbose mode")

	if err := cli.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
