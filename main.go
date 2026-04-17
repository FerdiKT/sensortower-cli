package main

import (
	"fmt"
	"os"

	"github.com/ferdikt/sensortower-cli/cmd"
	"github.com/ferdikt/sensortower-cli/internal/clierror"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(clierror.Code(err))
	}
}
