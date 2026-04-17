package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ferdikt/sensortower-cli/internal/config"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	Output     string
	ConfigPath string
}

var opts globalOptions

var rootCmd = &cobra.Command{
	Use:   "sensortower",
	Short: "Sensor Tower CLI",
	Long:  "A command line interface for Sensor Tower market intelligence endpoints.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch opts.Output {
		case "table", "json":
			return nil
		default:
			return fmt.Errorf("invalid --output value %q (allowed: table, json)", opts.Output)
		}
	},
}

func init() {
	defaultConfigPath, err := config.DefaultPath()
	if err != nil {
		defaultConfigPath = "sensortower-config.json"
	}

	rootCmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "table", "Output format (table|json)")
	rootCmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", defaultConfigPath, "Config file path")
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
	if opts.Output == "table" && cfg.Output != "" {
		// Keep explicit CLI flag precedence. Cobra does not expose the raw default
		// cleanly here, so only honor config when user left the default unchanged.
		if !rootCmd.PersistentFlags().Changed("output") {
			opts.Output = cfg.Output
		}
	}
	return sensortower.NewClient(sensortower.Options{
		BaseURL:        cfg.BaseURL,
		TimeoutSeconds: cfg.TimeoutSeconds,
		Cookie:         cfg.Cookie,
		Headers:        cfg.Headers,
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
	os.Exit(1)
}
