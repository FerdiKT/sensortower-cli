package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowCompetitorsCmd, workflowFreshEarnersCmd)
	workflowCompetitorsCmd.Flags().String("country", "US", "Store country")
	workflowCompetitorsCmd.Flags().String("categories", "", "Comma-separated category IDs")
	workflowCompetitorsCmd.Flags().Int("top", 200, "Maximum unique apps to enrich")
	workflowCompetitorsCmd.Flags().String("device", "iphone", "Device type")
	workflowCompetitorsCmd.Flags().String("date", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "Chart date in YYYY-MM-DD")
	workflowCompetitorsCmd.Flags().Int("concurrency", 4, "Concurrent app metadata enrich requests")
	_ = workflowCompetitorsCmd.MarkFlagRequired("categories")

	workflowFreshEarnersCmd.Flags().String("country", "US", "Store country")
	workflowFreshEarnersCmd.Flags().String("categories", "0", "Comma-separated category IDs (0 means all categories)")
	workflowFreshEarnersCmd.Flags().String("device", "iphone", "Device type")
	workflowFreshEarnersCmd.Flags().String("date", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "Chart date in YYYY-MM-DD")
	workflowFreshEarnersCmd.Flags().Int("months", 1, "Release recency window in months")
	workflowFreshEarnersCmd.Flags().Int64("min-revenue-usd", 10000, "Minimum last-month revenue in USD")
	workflowFreshEarnersCmd.Flags().Int("top", 200, "Maximum unique ranked apps to inspect before filtering")
	workflowFreshEarnersCmd.Flags().Int("concurrency", 4, "Concurrent app metadata enrich requests")
}

var workflowCmd = &cobra.Command{Use: "workflow", Short: "Higher-level workflows"}

var workflowCompetitorsCmd = &cobra.Command{
	Use:   "competitors",
	Short: "Fetch rankings, dedupe apps, and enrich competitor metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		country, _ := cmd.Flags().GetString("country")
		categoriesText, _ := cmd.Flags().GetString("categories")
		device, _ := cmd.Flags().GetString("device")
		top, _ := cmd.Flags().GetInt("top")
		date, _ := cmd.Flags().GetString("date")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		categories, err := parseInts(categoriesText)
		if err != nil {
			return clierror.Wrap(11, err.Error())
		}
		if top <= 0 {
			return clierror.Wrap(11, "top must be greater than 0")
		}
		seen := map[int64]*sensortower.CompetitorRecord{}
		for _, category := range categories {
			resp, _, err := client.CategoryRankings(commandContext(cmd), country, category, date, device, categoryRankingsPageCap, 0)
			if err != nil {
				return err
			}
			addCompetitorBucket(seen, category, "free", country, resp.Data.Free)
			addCompetitorBucket(seen, category, "grossing", country, resp.Data.Grossing)
			addCompetitorBucket(seen, category, "paid", country, resp.Data.Paid)
		}
		appIDs := make([]int64, 0, len(seen))
		for appID := range seen {
			appIDs = append(appIDs, appID)
		}
		sort.Slice(appIDs, func(i, j int) bool {
			left := seen[appIDs[i]]
			right := seen[appIDs[j]]
			leftRank := bestObservedRank(left)
			rightRank := bestObservedRank(right)
			if leftRank != rightRank {
				return leftRank < rightRank
			}
			return appIDs[i] < appIDs[j]
		})
		if len(appIDs) > top {
			appIDs = appIDs[:top]
		}
		if err := enrichCompetitors(commandContext(cmd), client, seen, appIDs, country, concurrency); err != nil {
			return err
		}
		rows := make([]map[string]any, 0, len(appIDs))
		for _, appID := range appIDs {
			rows = append(rows, structToMap(*seen[appID]))
		}
		return writeOutput(rows)
	},
}

var workflowFreshEarnersCmd = &cobra.Command{
	Use:   "fresh-earners",
	Short: "Find recently released apps above a monthly revenue threshold",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		country, _ := cmd.Flags().GetString("country")
		categoriesText, _ := cmd.Flags().GetString("categories")
		device, _ := cmd.Flags().GetString("device")
		date, _ := cmd.Flags().GetString("date")
		months, _ := cmd.Flags().GetInt("months")
		minRevenueUSD, _ := cmd.Flags().GetInt64("min-revenue-usd")
		top, _ := cmd.Flags().GetInt("top")
		concurrency, _ := cmd.Flags().GetInt("concurrency")
		if months <= 0 {
			return clierror.Wrap(11, "months must be greater than 0")
		}
		if minRevenueUSD < 0 {
			return clierror.Wrap(11, "min-revenue-usd must be 0 or greater")
		}
		if top <= 0 {
			return clierror.Wrap(11, "top must be greater than 0")
		}
		categories, err := parseInts(categoriesText)
		if err != nil {
			return clierror.Wrap(11, err.Error())
		}
		if len(categories) == 0 {
			categories = []int{0}
		}

		seen := map[int64]*sensortower.CompetitorRecord{}
		for _, category := range categories {
			resp, _, err := client.CategoryRankings(commandContext(cmd), country, category, date, device, categoryRankingsPageCap, 0)
			if err != nil {
				return err
			}
			addCompetitorBucket(seen, category, "free", country, resp.Data.Free)
			addCompetitorBucket(seen, category, "grossing", country, resp.Data.Grossing)
			addCompetitorBucket(seen, category, "paid", country, resp.Data.Paid)
		}

		appIDs := make([]int64, 0, len(seen))
		for appID := range seen {
			appIDs = append(appIDs, appID)
		}
		sort.Slice(appIDs, func(i, j int) bool {
			left := seen[appIDs[i]]
			right := seen[appIDs[j]]
			leftRank := bestObservedRank(left)
			rightRank := bestObservedRank(right)
			if leftRank != rightRank {
				return leftRank < rightRank
			}
			return appIDs[i] < appIDs[j]
		})
		if len(appIDs) > top {
			appIDs = appIDs[:top]
		}
		if err := enrichCompetitors(commandContext(cmd), client, seen, appIDs, country, concurrency); err != nil {
			return err
		}

		cutoff := time.Now().AddDate(0, -months, 0)
		minRevenueCents := minRevenueUSD * 100
		rows := make([]map[string]any, 0, len(appIDs))
		for _, appID := range appIDs {
			record := seen[appID]
			if record == nil || record.Enriched == nil {
				continue
			}
			enriched := record.Enriched
			revenueCents := int64FromAny(enriched["worldwide_last_month_revenue"], "value")
			releaseAt := firstTimestamp(enriched["release_date"], enriched["worldwide_release_date"], enriched["country_release_date"])
			if releaseAt.IsZero() || releaseAt.Before(cutoff) || revenueCents < minRevenueCents {
				continue
			}

			row := map[string]any{
				"app_id":                  record.AppID,
				"name":                    firstNonEmptyString(record.Name, stringFromAny(enriched["name"])),
				"publisher_name":          firstNonEmptyString(record.PublisherName, stringFromAny(enriched["publisher_name"])),
				"country":                 country,
				"release_date":            releaseAt.UTC().Format(time.RFC3339),
				"months_window":           months,
				"monthly_revenue_usd":     float64(revenueCents) / 100.0,
				"monthly_revenue_cents":   revenueCents,
				"min_revenue_usd":         minRevenueUSD,
				"matched_categories":      record.Categories,
				"matched_ranking_buckets": record.Buckets,
				"observed_ranks":          record.ObservedRanks,
				"enriched":                enriched,
			}
			rows = append(rows, row)
		}

		sort.Slice(rows, func(i, j int) bool {
			left := numericFromAny(rows[i]["monthly_revenue_cents"])
			right := numericFromAny(rows[j]["monthly_revenue_cents"])
			if left != right {
				return left > right
			}
			return int64FromAny(rows[i]["app_id"]) < int64FromAny(rows[j]["app_id"])
		})
		return writeOutput(rows)
	},
}

func addCompetitorBucket(seen map[int64]*sensortower.CompetitorRecord, category int, bucket, country string, entries []sensortower.CategoryRankingEntry) {
	for _, entry := range entries {
		record := seen[entry.AppID]
		if record == nil {
			record = &sensortower.CompetitorRecord{AppID: entry.AppID, Name: entry.Name, PublisherName: entry.PublisherName, Country: country}
			seen[entry.AppID] = record
		}
		record.Categories = appendUniqueInt(record.Categories, category)
		record.Buckets = appendUnique(record.Buckets, bucket)
		record.ObservedRanks = append(record.ObservedRanks, map[string]any{"category": category, "bucket": bucket, "rank": entry.Rank, "previous_rank": entry.PreviousRank})
	}
}

func bestObservedRank(record *sensortower.CompetitorRecord) int {
	best := 1 << 30
	for _, observed := range record.ObservedRanks {
		value, ok := observed["rank"].(int)
		if ok && value > 0 && value < best {
			best = value
		}
	}
	return best
}

func enrichCompetitors(ctx context.Context, client *sensortower.Client, seen map[int64]*sensortower.CompetitorRecord, appIDs []int64, country string, concurrency int) error {
	if len(appIDs) == 0 {
		return nil
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	type result struct {
		appID int64
		resp  *sensortower.AppDetails
		err   error
	}
	jobs := make(chan int64)
	results := make(chan result, len(appIDs))
	for i := 0; i < concurrency; i++ {
		go func() {
			for appID := range jobs {
				resp, _, err := client.AppDetails(ctx, appID, country)
				results <- result{appID: appID, resp: resp, err: err}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, appID := range appIDs {
			jobs <- appID
		}
	}()

	var completed int32
	for range appIDs {
		result := <-results
		done := atomic.AddInt32(&completed, 1)
		if result.err == nil && result.resp != nil {
			seen[result.appID].Enriched = result.resp.Raw
			seen[result.appID].MetadataFetchedAt = time.Now().UTC()
		}
		if len(appIDs) >= 10 && (done == 1 || done%10 == 0 || int(done) == len(appIDs)) {
			_, _ = fmt.Fprintf(os.Stderr, "workflow competitors: enriched %d/%d apps\n", done, len(appIDs))
		}
	}
	return nil
}

func int64FromAny(v any, path ...string) int64 {
	current := v
	for _, key := range path {
		node, ok := current.(map[string]any)
		if !ok {
			return 0
		}
		current = node[key]
	}
	switch n := current.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	}
	return 0
}

func stringFromAny(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstTimestamp(values ...any) time.Time {
	for _, value := range values {
		ts := int64FromAny(value)
		if ts <= 0 {
			continue
		}
		// Sensor Tower payloads can return epoch in seconds or milliseconds.
		if ts >= 1_000_000_000_000 {
			return time.UnixMilli(ts)
		}
		return time.Unix(ts, 0)
	}
	return time.Time{}
}

func numericFromAny(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	}
	return 0
}
