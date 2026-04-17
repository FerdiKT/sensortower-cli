package main

import (
	"fmt"
	"os"

	"github.com/ferdikt/sensortower-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
