package main

import (
	"os"
)

func main() {
	cli := NewCLI(os.Args)
	os.Exit(cli.Run())
}
