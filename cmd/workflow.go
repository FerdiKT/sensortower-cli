package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(workflowCmd)
	workflowCmd.AddCommand(workflowCompetitorsCmd)
	workflowCompetitorsCmd.Flags().String("country", "US", "Store country")
	workflowCompetitorsCmd.Flags().String("categories", "", "Comma-separated category IDs")
	workflowCompetitorsCmd.Flags().Int("top", 200, "Maximum unique apps to enrich")
	workflowCompetitorsCmd.Flags().String("device", "iphone", "Device type")
	workflowCompetitorsCmd.Flags().String("date", time.Now().AddDate(0, 0, -1).Format("2006-01-02"), "Chart date in YYYY-MM-DD")
	workflowCompetitorsCmd.Flags().Int("concurrency", 4, "Concurrent app metadata enrich requests")
	_ = workflowCompetitorsCmd.MarkFlagRequired("categories")
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
