package cmd

import (
	"os"

	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(publishersCmd)
	publishersCmd.AddCommand(publishersAppsCmd)

	publishersAppsCmd.Flags().Int64("publisher-id", 0, "Publisher ID")
	publishersAppsCmd.Flags().Int("limit", 25, "Page size")
	publishersAppsCmd.Flags().Int("offset", 0, "Pagination offset")
	publishersAppsCmd.Flags().String("sort-by", "downloads", "Sort field")
	_ = publishersAppsCmd.MarkFlagRequired("publisher-id")
}

var publishersCmd = &cobra.Command{
	Use:   "publishers",
	Short: "Publisher endpoints",
}

var publishersAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List apps for a publisher",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		publisherID, _ := cmd.Flags().GetInt64("publisher-id")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		sortBy, _ := cmd.Flags().GetString("sort-by")

		resp, err := client.PublisherApps(commandContext(cmd), publisherID, limit, offset, sortBy)
		if err != nil {
			return err
		}
		if opts.Output == "json" {
			return output.RenderJSON(os.Stdout, resp)
		}
		return output.RenderPublisherAppsTable(os.Stdout, resp)
	},
}
