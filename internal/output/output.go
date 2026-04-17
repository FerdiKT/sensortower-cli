package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ferdikt/sensortower-cli/internal/sensortower"
)

func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func RenderPublisherAppsTable(w io.Writer, resp *sensortower.PublisherAppsResponse) error {
	tw := newWriter(w)
	_, _ = fmt.Fprintln(tw, "APP ID\tNAME\tPUBLISHER\tDOWNLOADS(30D)\tREVENUE(30D)\tUPDATED")
	for _, app := range resp.Data {
		_, _ = fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n",
			app.AppID,
			app.HumanizedNameOrName(),
			app.PublisherName,
			metricString(app.WorldwideLast30DaysDownloads),
			metricString(app.WorldwideLast30DaysRevenue),
			app.UpdatedDate,
		)
	}
	return tw.Flush()
}

func RenderAppDetailsTable(w io.Writer, app *sensortower.AppDetails) error {
	tw := newWriter(w)
	rows := [][2]string{
		{"App ID", strconv.FormatInt(app.AppID, 10)},
		{"Name", app.Name},
		{"Publisher", app.PublisherName},
		{"Country", app.Country},
		{"Version", app.CurrentVersion},
		{"Rating", fmt.Sprintf("%.2f (%d)", app.Rating, app.RatingCount)},
		{"Downloads (last month)", strconv.FormatInt(app.WorldwideLastMonthDownloads.Value, 10)},
		{"Revenue (last month)", fmt.Sprintf("%s %d", strings.TrimSpace(app.WorldwideLastMonthRevenue.Currency), app.WorldwideLastMonthRevenue.Value)},
		{"Top countries", strings.Join(app.TopCountries, ", ")},
		{"Website", app.WebsiteURL},
		{"Support", app.SupportURL},
	}
	for _, row := range rows {
		if strings.TrimSpace(row[1]) == "" {
			continue
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\n", row[0], row[1])
	}
	return tw.Flush()
}

func RenderCategoryRankingsTable(w io.Writer, resp *sensortower.CategoryRankingsResponse) error {
	tw := newWriter(w)
	writeBucket := func(name string, entries []sensortower.CategoryRankingEntry) {
		if len(entries) == 0 {
			return
		}
		_, _ = fmt.Fprintf(tw, "%s\t\t\t\t\t\t\n", strings.ToUpper(name))
		_, _ = fmt.Fprintln(tw, "RANK\tPREV\tAPP ID\tNAME\tPUBLISHER\tDOWNLOADS\tREVENUE")
		for _, entry := range entries {
			_, _ = fmt.Fprintf(tw, "%d\t%d\t%d\t%s\t%s\t%s\t%s\n",
				entry.Rank,
				entry.PreviousRank,
				entry.AppID,
				entry.Name,
				entry.PublisherName,
				metricString(entry.WorldwideLastMonthDownloads),
				metricString(entry.WorldwideLastMonthRevenue),
			)
		}
		_, _ = fmt.Fprintln(tw)
	}
	writeBucket("free", resp.Data.Free)
	writeBucket("grossing", resp.Data.Grossing)
	writeBucket("paid", resp.Data.Paid)
	return tw.Flush()
}

func newWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}

func metricString(metric sensortower.HumanizedMetric) string {
	if metric.String != "" {
		return metric.String
	}
	if metric.Downloads != 0 {
		return strconv.FormatInt(metric.Downloads, 10)
	}
	if metric.Revenue != 0 {
		return strconv.FormatInt(metric.Revenue, 10)
	}
	return "-"
}
