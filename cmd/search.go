package cmd

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.AddCommand(searchAppsCmd, searchPublishersCmd)

	searchAppsCmd.Flags().String("term", "", "Search term")
	searchAppsCmd.Flags().String("os", "both_stores", "OS scope")
	searchAppsCmd.Flags().Int("limit", 20, "Result limit")
	searchAppsCmd.Flags().Bool("expand-entities", true, "Expand app entities")
	searchAppsCmd.Flags().Bool("flags", false, "Include flags payload")
	searchAppsCmd.Flags().Bool("mark-usage-disabled-apps", false, "Mark usage-disabled apps")
	_ = searchAppsCmd.MarkFlagRequired("term")

	searchPublishersCmd.Flags().String("term", "", "Search term")
	searchPublishersCmd.Flags().String("os", "both_stores", "OS scope")
	searchPublishersCmd.Flags().Int("limit", 20, "Result limit")
	searchPublishersCmd.Flags().Bool("include-extended-info", false, "Include extended publisher info")
	_ = searchPublishersCmd.MarkFlagRequired("term")
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Autocomplete search endpoints",
}

var searchAppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Search apps via autocomplete",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		term, _ := cmd.Flags().GetString("term")
		osValue, _ := cmd.Flags().GetString("os")
		limit, _ := cmd.Flags().GetInt("limit")
		expandEntities, _ := cmd.Flags().GetBool("expand-entities")
		flagsValue, _ := cmd.Flags().GetBool("flags")
		markUsageDisabledApps, _ := cmd.Flags().GetBool("mark-usage-disabled-apps")
		if strings.TrimSpace(term) == "" {
			return clierror.Wrap(11, "term is required")
		}
		resp, meta, err := client.AppAutocomplete(commandContext(cmd), term, osValue, limit, expandEntities, flagsValue, markUsageDisabledApps)
		if err != nil {
			return err
		}
		emitMeta(meta)
		if opts.Output != "table" {
			return writeOutputWithMeta(resp, meta)
		}
		w, err := outputWriter()
		if err != nil {
			return err
		}
		defer w.Close()
		return renderSearchAppsTable(w, resp)
	},
}

var searchPublishersCmd = &cobra.Command{
	Use:   "publishers",
	Short: "Search publishers via autocomplete",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		term, _ := cmd.Flags().GetString("term")
		osValue, _ := cmd.Flags().GetString("os")
		limit, _ := cmd.Flags().GetInt("limit")
		includeExtendedInfo, _ := cmd.Flags().GetBool("include-extended-info")
		if strings.TrimSpace(term) == "" {
			return clierror.Wrap(11, "term is required")
		}
		resp, meta, err := client.PublisherAutocomplete(commandContext(cmd), term, osValue, limit, includeExtendedInfo)
		if err != nil {
			return err
		}
		emitMeta(meta)
		if opts.Output != "table" {
			return writeOutputWithMeta(resp, meta)
		}
		w, err := outputWriter()
		if err != nil {
			return err
		}
		defer w.Close()
		return renderSearchPublishersTable(w, resp)
	},
}

func renderSearchAppsTable(w io.Writer, rows []map[string]any) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "APP ID\tNAME\tPUBLISHER\tOS\tCOUNTRY")
	for _, row := range rows {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			stringValue(row["app_id"], row["id"]),
			stringValue(row["name"], row["humanized_name"], row["title"]),
			stringValue(row["publisher_name"], row["developer_name"]),
			stringValue(row["os"]),
			stringValue(row["country"], row["canonical_country"]),
		)
	}
	return tw.Flush()
}

func renderSearchPublishersTable(w io.Writer, rows []map[string]any) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "PUBLISHER ID\tNAME\tCOUNTRY\tOS")
	for _, row := range rows {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			stringValue(row["publisher_id"], row["id"]),
			stringValue(row["publisher_name"], row["name"]),
			stringValue(row["country"], row["canonical_country"]),
			stringValue(row["os"]),
		)
	}
	return tw.Flush()
}

func stringValue(values ...any) string {
	for _, value := range values {
		switch x := value.(type) {
		case nil:
			continue
		case string:
			if strings.TrimSpace(x) != "" {
				return x
			}
		default:
			text := fmt.Sprint(x)
			if strings.TrimSpace(text) != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}
