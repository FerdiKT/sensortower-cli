package cmd

import (
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
	publishersAppsCmd.Flags().Bool("all-pages", false, "Automatically paginate through all pages")
	publishersAppsCmd.Flags().Int("max-items", 0, "Maximum number of items to return when paginating")
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
		allPages, _ := cmd.Flags().GetBool("all-pages")
		maxItems, _ := cmd.Flags().GetInt("max-items")

		resp, meta, err := client.PublisherApps(commandContext(cmd), publisherID, limit, offset, sortBy)
		if err != nil {
			return err
		}
		if allPages {
			for {
				if maxItems > 0 && len(resp.Data) >= maxItems {
					resp.Data = resp.Data[:maxItems]
					break
				}
				if len(resp.Data) < limit {
					break
				}
				offset += limit
				page, pageMeta, err := client.PublisherApps(commandContext(cmd), publisherID, limit, offset, sortBy)
				if err != nil {
					return err
				}
				if pageMeta != nil && meta != nil && pageMeta.Retried > meta.Retried {
					meta = pageMeta
				}
				if len(page.Data) == 0 {
					break
				}
				resp.Data = append(resp.Data, page.Data...)
			}
		}
		emitMeta(meta)
		if opts.Output != "table" {
			rows := make([]map[string]any, 0, len(resp.Data))
			for _, app := range resp.Data {
				rows = append(rows, structToMap(app))
			}
			return writeOutputWithMeta(rows, meta)
		}
		w, err := outputWriter()
		if err != nil {
			return err
		}
		defer w.Close()
		return output.RenderPublisherAppsTable(w, resp)
	},
}
