package main

import (
	"os"

	"github.com/grobmeier/humblebee/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

