package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ferdikt/sensortower-cli/internal/cache"
	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/config"
	"github.com/ferdikt/sensortower-cli/internal/exitcode"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	Output     string
	ConfigPath string
	Context    string
	OutputFile string
	OutputFmt  string
	Retry429   bool
	RetryMax   int
	RetryWait  int
	NoCache    bool
	CacheTTL   int
}

var opts globalOptions

var rootCmd = &cobra.Command{
	Use:   "sensortower",
	Short: "Sensor Tower CLI",
	Long:  "A command line interface for Sensor Tower market intelligence endpoints.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch opts.Output {
		case "table", "json", "jsonl", "csv":
			return nil
		default:
			return clierror.Wrap(exitcode.Usage, fmt.Sprintf("invalid --output value %q (allowed: table, json, jsonl, csv)", opts.Output))
		}
	},
}

func init() {
	defaultConfigPath, err := config.DefaultPath()
	if err != nil {
		defaultConfigPath = "sensortower-config.json"
	}

	rootCmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "table", "Output format (table|json|jsonl|csv)")
	rootCmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", defaultConfigPath, "Config file path")
	rootCmd.PersistentFlags().StringVar(&opts.Context, "context", "", "Named context to use")
	rootCmd.PersistentFlags().StringVar(&opts.OutputFile, "output-file", "", "Write output to a file instead of stdout")
	rootCmd.PersistentFlags().StringVar(&opts.OutputFmt, "output-format", "", "Explicit output format (table|json|jsonl|csv)")
	rootCmd.PersistentFlags().BoolVar(&opts.Retry429, "retry-429", false, "Retry HTTP 429 responses with backoff")
	rootCmd.PersistentFlags().IntVar(&opts.RetryMax, "retry-max", 8, "Maximum number of 429 retries")
	rootCmd.PersistentFlags().IntVar(&opts.RetryWait, "retry-wait", 60, "Fallback wait seconds for 429 retries")
	rootCmd.PersistentFlags().BoolVar(&opts.NoCache, "no-cache", false, "Disable built-in cache")
	rootCmd.PersistentFlags().IntVar(&opts.CacheTTL, "cache-ttl", 300, "Cache TTL in seconds")
}

func Execute() error {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	return rootCmd.Execute()
}

func newClient() (*sensortower.Client, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	ctxCfg := cfg.EffectiveContext(opts.Context)
	if strings.TrimSpace(opts.OutputFmt) != "" {
		opts.Output = opts.OutputFmt
	}
	if opts.Output == "table" && ctxCfg.Output != "" {
		// Keep explicit CLI flag precedence. Cobra does not expose the raw default
		// cleanly here, so only honor config when user left the default unchanged.
		if !rootCmd.PersistentFlags().Changed("output") {
			opts.Output = ctxCfg.Output
		}
	}
	ttlSeconds := opts.CacheTTL
	if ttlSeconds == 300 && cfg.CacheTTLSeconds > 0 && !rootCmd.PersistentFlags().Changed("cache-ttl") {
		ttlSeconds = cfg.CacheTTLSeconds
	}
	var clientCache *cache.Cache
	if !opts.NoCache && ttlSeconds > 0 {
		cacheDir, err := cache.DefaultDir()
		if err == nil {
			clientCache = cache.New(cacheDir, time.Duration(ttlSeconds)*time.Second)
		}
	}
	return sensortower.NewClient(sensortower.Options{
		BaseURL:        ctxCfg.BaseURL,
		TimeoutSeconds: ctxCfg.TimeoutSeconds,
		Cookie:         ctxCfg.Cookie,
		Headers:        ctxCfg.Headers,
		Retry429:       opts.Retry429,
		RetryMax:       opts.RetryMax,
		RetryWait:      opts.RetryWait,
		Cache:          clientCache,
	}), nil
}

func commandContext(cmd *cobra.Command) context.Context {
	if cmd.Context() != nil {
		return cmd.Context()
	}
	return context.Background()
}

func Exitf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(exitcode.Internal)
}
