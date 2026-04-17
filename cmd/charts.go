package cmd

import (
	"os"

	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chartsCmd)
	chartsCmd.AddCommand(categoryRankingsCmd)

	categoryRankingsCmd.Flags().String("country", "US", "Store country")
	categoryRankingsCmd.Flags().Int("category", 0, "Category ID")
	categoryRankingsCmd.Flags().String("date", "", "Chart date in YYYY-MM-DD")
	categoryRankingsCmd.Flags().String("device", "iphone", "Device type")
	categoryRankingsCmd.Flags().Int("limit", 25, "Page size")
	categoryRankingsCmd.Flags().Int("offset", 0, "Pagination offset")
	_ = categoryRankingsCmd.MarkFlagRequired("date")
}

var chartsCmd = &cobra.Command{
	Use:   "charts",
	Short: "Chart endpoints",
}

var categoryRankingsCmd = &cobra.Command{
	Use:   "category-rankings",
	Short: "Fetch iOS category rankings",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		country, _ := cmd.Flags().GetString("country")
		category, _ := cmd.Flags().GetInt("category")
		date, _ := cmd.Flags().GetString("date")
		device, _ := cmd.Flags().GetString("device")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		resp, err := client.CategoryRankings(commandContext(cmd), country, category, date, device, limit, offset)
		if err != nil {
			return err
		}
		if opts.Output == "json" {
			return output.RenderJSON(os.Stdout, resp)
		}
		return output.RenderCategoryRankingsTable(os.Stdout, resp)
	},
}
