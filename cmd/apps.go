package cmd

import (
	"os"

	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsGetCmd)

	appsGetCmd.Flags().Int64("app-id", 0, "App ID")
	appsGetCmd.Flags().String("country", "US", "Store country")
	_ = appsGetCmd.MarkFlagRequired("app-id")
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "App endpoints",
}

var appsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetch app details",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		appID, _ := cmd.Flags().GetInt64("app-id")
		country, _ := cmd.Flags().GetString("country")

		resp, err := client.AppDetails(commandContext(cmd), appID, country)
		if err != nil {
			return err
		}
		if opts.Output == "json" {
			return output.RenderJSON(os.Stdout, resp)
		}
		return output.RenderAppDetailsTable(os.Stdout, resp)
	},
}
