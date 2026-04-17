package cmd

import (
	"fmt"

	"github.com/ferdikt/sensortower-cli/internal/buildinfo"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.Version = buildinfo.Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print sensortower version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildinfo.Version)
	},
}
