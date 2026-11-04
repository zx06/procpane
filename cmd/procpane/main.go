package main

import (
	"os"

	"github.com/zx06/procpane/cmd/procpane/cmdroot"
)

func main() {
	if err := cmdroot.Execute(); err != nil {
		os.Exit(1)
	}
}
