package main

import (
	"os"

	"github.com/grobmeier/humblebee/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetBuildInfo(cli.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
