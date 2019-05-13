package main

import (
	"fmt"
	"os"

	"github.com/pierrec/cmdflag"
)

var cli = cmdflag.New(nil)

func main() {
	if err := cli.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
