package main

import (
	"os"

	"github.com/catfly/restic-helpers/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
