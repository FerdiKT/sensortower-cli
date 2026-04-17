package cmd

import (
	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
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
	categoryRankingsCmd.Flags().Bool("all-pages", false, "Automatically paginate through all pages")
	categoryRankingsCmd.Flags().Int("max-items", 0, "Maximum items per bucket when paginating")
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
		allPages, _ := cmd.Flags().GetBool("all-pages")
		maxItems, _ := cmd.Flags().GetInt("max-items")

		resp, meta, err := client.CategoryRankings(commandContext(cmd), country, category, date, device, limit, offset)
		if err != nil {
			return err
		}
		if allPages {
			for {
				if stopAtBucketMax(resp, maxItems) || bucketCount(resp) < limit {
					break
				}
				offset += limit
				page, pageMeta, err := client.CategoryRankings(commandContext(cmd), country, category, date, device, limit, offset)
				if err != nil {
					return err
				}
				if pageMeta != nil && meta != nil && pageMeta.Retried > meta.Retried {
					meta = pageMeta
				}
				if bucketCount(page) == 0 {
					break
				}
				resp.Data.Free = append(resp.Data.Free, page.Data.Free...)
				resp.Data.Grossing = append(resp.Data.Grossing, page.Data.Grossing...)
				resp.Data.Paid = append(resp.Data.Paid, page.Data.Paid...)
			}
			if maxItems > 0 {
				trimBuckets(resp, maxItems)
			}
		}
		emitMeta(meta)
		if opts.Output != "table" {
			var rows []map[string]any
			for _, bucket := range []struct {
				name  string
				items []sensortower.CategoryRankingEntry
			}{
				{"free", resp.Data.Free},
				{"grossing", resp.Data.Grossing},
				{"paid", resp.Data.Paid},
			} {
				for _, item := range bucket.items {
					row := structToMap(item)
					row["bucket"] = bucket.name
					rows = append(rows, row)
				}
			}
			return writeOutputWithMeta(rows, meta)
		}
		w, err := outputWriter()
		if err != nil {
			return err
		}
		defer w.Close()
		return output.RenderCategoryRankingsTable(w, resp)
	},
}

func bucketCount(resp *sensortower.CategoryRankingsResponse) int {
	return max(len(resp.Data.Free), max(len(resp.Data.Grossing), len(resp.Data.Paid)))
}

func stopAtBucketMax(resp *sensortower.CategoryRankingsResponse, maxItems int) bool {
	return maxItems > 0 && len(resp.Data.Free) >= maxItems && len(resp.Data.Grossing) >= maxItems && len(resp.Data.Paid) >= maxItems
}

func trimBuckets(resp *sensortower.CategoryRankingsResponse, maxItems int) {
	for _, bucket := range []*[]sensortower.CategoryRankingEntry{&resp.Data.Free, &resp.Data.Grossing, &resp.Data.Paid} {
		if len(*bucket) > maxItems {
			*bucket = (*bucket)[:maxItems]
		}
	}
}
