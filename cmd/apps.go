package cmd

import (
	"context"
	"sync"
	"time"

	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/fields"
	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(appsCmd)
	appsCmd.AddCommand(appsGetCmd)

	appsGetCmd.Flags().Int64("app-id", 0, "App ID")
	appsGetCmd.Flags().String("country", "US", "Store country")
	appsGetCmd.Flags().String("app-ids-file", "", "Read app IDs from a file, one per line")
	appsGetCmd.Flags().String("fields", "", "Comma-separated top-level or dotted fields to return")
	appsGetCmd.Flags().Int("concurrency", 4, "Batch worker count")
	appsGetCmd.Flags().Int("throttle-ms", 250, "Delay between batch requests in milliseconds")
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
		idsFile, _ := cmd.Flags().GetString("app-ids-file")
		fieldList, _ := cmd.Flags().GetString("fields")
		selectedFields := fields.Parse(fieldList)

		if idsFile != "" {
			ids, err := readInt64Lines(idsFile)
			if err != nil {
				return clierror.Wrap(11, err.Error())
			}
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			throttleMS, _ := cmd.Flags().GetInt("throttle-ms")
			result, err := fetchAppsBatch(commandContext(cmd), client, ids, country, selectedFields, concurrency, throttleMS)
			if err != nil {
				return err
			}
			return writeOutput(result)
		}
		if appID == 0 {
			return clierror.Wrap(11, "either --app-id or --app-ids-file is required")
		}

		resp, meta, err := client.AppDetails(commandContext(cmd), appID, country)
		if err != nil {
			return err
		}
		emitMeta(meta)
		if opts.Output != "table" {
			raw := resp.Raw
			if len(selectedFields) > 0 {
				raw = fields.FilterMap(resp.Raw, selectedFields)
			}
			return writeOutputWithMeta(raw, meta)
		}
		w, err := outputWriter()
		if err != nil {
			return err
		}
		defer w.Close()
		return output.RenderAppDetailsTable(w, resp)
	},
}

func fetchAppsBatch(ctx context.Context, client *sensortower.Client, ids []int64, country string, selectedFields []string, concurrency, throttleMS int) (*sensortower.AppsBatchResult, error) {
	if concurrency <= 0 {
		concurrency = 1
	}
	result := &sensortower.AppsBatchResult{}
	type job struct{ appID int64 }
	jobs := make(chan job)
	var mu sync.Mutex
	var wg sync.WaitGroup
	ticker := time.NewTicker(time.Duration(max(throttleMS, 1)) * time.Millisecond)
	defer ticker.Stop()

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			if throttleMS > 0 {
				<-ticker.C
			}
			resp, meta, err := client.AppDetails(ctx, j.appID, country)
			mu.Lock()
			if err != nil {
				result.Failed = append(result.Failed, sensortower.AppsBatchFailure{AppID: j.appID, Error: err.Error()})
				mu.Unlock()
				continue
			}
			row := resp.Raw
			if len(selectedFields) > 0 {
				row = fields.FilterMap(resp.Raw, selectedFields)
			}
			result.OK = append(result.OK, row)
			if meta != nil && meta.Retried > result.Meta.Retried {
				result.Meta = *meta
			}
			mu.Unlock()
		}
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker()
	}
	for _, appID := range ids {
		jobs <- job{appID: appID}
	}
	close(jobs)
	wg.Wait()
	return result, nil
}
