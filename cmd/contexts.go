package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/ferdikt/sensortower-cli/internal/clierror"
	"github.com/ferdikt/sensortower-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(contextsCmd)
	contextsCmd.AddCommand(contextsAddCmd, contextsListCmd, contextsUseCmd)

	contextsAddCmd.Flags().String("name", "", "Context name")
	contextsAddCmd.Flags().String("base-url", "", "Base URL override")
	contextsAddCmd.Flags().String("cookie", "", "Cookie value")
	contextsAddCmd.Flags().String("headers-json", "", "Extra headers as JSON")
	contextsAddCmd.Flags().Int("timeout-seconds", 30, "Timeout in seconds")
	_ = contextsAddCmd.MarkFlagRequired("name")

	contextsUseCmd.Flags().String("name", "", "Context name")
	_ = contextsUseCmd.MarkFlagRequired("name")
}

var contextsCmd = &cobra.Command{Use: "contexts", Short: "Manage named contexts"}

var contextsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add or update a named context",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		baseURL, _ := cmd.Flags().GetString("base-url")
		cookie, _ := cmd.Flags().GetString("cookie")
		headersJSON, _ := cmd.Flags().GetString("headers-json")
		timeoutSeconds, _ := cmd.Flags().GetInt("timeout-seconds")
		headers, err := parseHeadersJSON(headersJSON)
		if err != nil {
			return clierror.Wrap(11, err.Error())
		}
		cfg.SetContext(name, config.Context{
			BaseURL:        baseURL,
			Cookie:         cookie,
			Headers:        headers,
			TimeoutSeconds: timeoutSeconds,
			Output:         cfg.Output,
		})
		if cfg.ActiveContext == "" {
			cfg.ActiveContext = name
		}
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "context saved: %s\n", name)
		return err
	},
}

var contextsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			return err
		}
		names := make([]string, 0, len(cfg.Contexts))
		for name := range cfg.Contexts {
			names = append(names, name)
		}
		sort.Strings(names)
		rows := make([]map[string]any, 0, len(names))
		for _, name := range names {
			ctx := cfg.Contexts[name]
			rows = append(rows, map[string]any{
				"name":       name,
				"active":     cfg.ActiveContext == name,
				"base_url":   ctx.BaseURL,
				"has_cookie": ctx.Cookie != "",
			})
		}
		return writeOutput(rows)
	},
}

var contextsUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Set the active context",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(opts.ConfigPath)
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if _, ok := cfg.Contexts[name]; !ok {
			return clierror.Wrap(11, fmt.Sprintf("context not found: %s", name))
		}
		cfg.ActiveContext = name
		if err := cfg.Save(opts.ConfigPath); err != nil {
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "active context: %s\n", name)
		return err
	},
}
