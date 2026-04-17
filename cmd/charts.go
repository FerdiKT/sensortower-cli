package cmd

import (
	"fmt"
	"os"

	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/spf13/cobra"
)

const categoryRankingsPageCap = 25

func init() {
	rootCmd.AddCommand(chartsCmd)
	chartsCmd.AddCommand(categoryRankingsCmd)

	categoryRankingsCmd.Flags().String("country", "US", "Store country")
	categoryRankingsCmd.Flags().Int("category", 0, "Category ID")
	categoryRankingsCmd.Flags().String("date", "", "Chart date in YYYY-MM-DD")
	categoryRankingsCmd.Flags().String("device", "iphone", "Device type")
	categoryRankingsCmd.Flags().Int("limit", 25, "Page size")
	categoryRankingsCmd.Flags().Int("offset", 0, "Pagination offset")
	categoryRankingsCmd.Flags().Bool("all-pages", false, "Collect all rows available from the endpoint (currently capped at 25 per bucket)")
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
		limit, offset, err = normalizeCategoryRankingsPage(limit, offset, allPages)
		if err != nil {
			return err
		}

		resp, meta, err := client.CategoryRankings(commandContext(cmd), country, category, date, device, limit, offset)
		if err != nil {
			return err
		}
		if allPages {
			if bucketCount(resp) == categoryRankingsPageCap {
				_, _ = fmt.Fprintln(os.Stderr, "note: category_rankings is currently capped at 25 rows per bucket by the upstream endpoint")
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

func normalizeCategoryRankingsPage(limit, offset int, allPages bool) (int, int, error) {
	if offset < 0 {
		return 0, 0, clierror.Wrap(11, "offset must be 0 or greater")
	}
	if offset >= categoryRankingsPageCap {
		return 0, 0, clierror.Wrap(11, "category_rankings supports offsets below 25 only")
	}
	if limit <= 0 || limit > categoryRankingsPageCap {
		limit = categoryRankingsPageCap
	}
	if allPages {
		offset = 0
		limit = categoryRankingsPageCap
	}
	return limit, offset, nil
}

func bucketCount(resp *sensortower.CategoryRankingsResponse) int {
	return max(len(resp.Data.Free), max(len(resp.Data.Grossing), len(resp.Data.Paid)))
}

func trimBuckets(resp *sensortower.CategoryRankingsResponse, maxItems int) {
	for _, bucket := range []*[]sensortower.CategoryRankingEntry{&resp.Data.Free, &resp.Data.Grossing, &resp.Data.Paid} {
		if len(*bucket) > maxItems {
			*bucket = (*bucket)[:maxItems]
		}
	}
}
