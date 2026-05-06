package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	bundledSkillName = "sensortower-cli"
	bundledSkillDoc  = `---
name: sensortower-cli
description: Use this skill when working with the local ` + "`sensortower`" + ` CLI for Sensor Tower iOS market data, category rankings, app details, publisher app lookups, JSON-first workflows, and repo-local release or Homebrew tap maintenance.
---

# SensorTower CLI

Use this skill for repository-local Sensor Tower CLI work.

## Workflow

1. Prefer public endpoint reads first.
2. Preserve upstream JSON fields unless a change is clearly needed.
3. Prefer JSON output for agent workflows.
4. Keep table output compact and operationally useful.
5. Validate with repo smoke checks after CLI changes.

## Read Pattern

- Use ` + "`sensortower publishers apps`" + ` for publisher-level app lookups.
- Use ` + "`sensortower apps get`" + ` for a full app detail payload.
- Use ` + "`sensortower charts category-rankings`" + ` for free, grossing, and paid rankings.
- Use ` + "`sensortower workflow fresh-earners`" + ` to find newly released apps above a revenue threshold (defaults: last 1 month and >= $10k).
- Add ` + "`--output json`" + ` for agent consumption.

## Config Pattern

- Default config is loaded from the user config directory.
- Override with ` + "`--config`" + ` when testing alternate setups.
- Use env vars like ` + "`SENSORTOWER_COOKIE`" + ` and ` + "`SENSORTOWER_HEADERS_JSON`" + ` when session-backed requests are needed.

## Release Pattern

- Run ` + "`go test ./...`" + ` before release work.
- Build archives with ` + "`make brew-dist VERSION=<version>`" + `.
- Push tags like ` + "`v0.1.2`" + ` to trigger GitHub releases.
- Keep the Homebrew tap formula in sync with the latest release artifacts and checksums.
`
	bundledAgentYAML = `display_name: Sensor Tower CLI
short_description: Work with the local Sensor Tower CLI for iOS market data, JSON reads, and release maintenance.
default_prompt: Use the local sensortower CLI. Prefer public endpoint reads first, prefer JSON output for agent workflows, preserve upstream payload shapes, and validate with repo smoke checks after changes.
`
)

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(newAgentInstallSkillCommand())
	agentCmd.AddCommand(newAgentLinkSkillCommand())
	agentCmd.AddCommand(newAgentShowSkillPathCommand())
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install or link agent assets such as the bundled Codex skill",
}

func newAgentInstallSkillCommand() *cobra.Command {
	var codexHome string
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install the bundled sensortower Codex skill into CODEX_HOME",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			if err := ensureReplaceableTarget(targetDir, force); err != nil {
				return err
			}
			if err := writeBundledSkill(targetDir); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "skill installed: %s\n", targetDir)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing installed skill")
	return cmd
}

func newAgentLinkSkillCommand() *cobra.Command {
	var codexHome string
	var source string
	var force bool

	cmd := &cobra.Command{
		Use:   "link-skill",
		Short: "Symlink the local sensortower skill into CODEX_HOME",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			if source == "" {
				source = filepath.Join(".", "skills", bundledSkillName)
			}
			source, err = filepath.Abs(source)
			if err != nil {
				return fmt.Errorf("resolve source path: %w", err)
			}
			if err := validateSkillSource(source); err != nil {
				return err
			}
			if err := ensureReplaceableTarget(targetDir, force); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
				return fmt.Errorf("create destination parent: %w", err)
			}
			if err := os.Symlink(source, targetDir); err != nil {
				return fmt.Errorf("create symlink: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "skill linked: %s -> %s\n", targetDir, source)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	cmd.Flags().StringVar(&source, "source", "", "Source skill directory to symlink; defaults to ./skills/sensortower-cli")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing installed skill or symlink")
	return cmd
}

func newAgentShowSkillPathCommand() *cobra.Command {
	var codexHome string

	cmd := &cobra.Command{
		Use:   "show-skill-path",
		Short: "Print the destination path for the sensortower Codex skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), targetDir)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	return cmd
}

func resolveSkillInstallDir(codexHome string) (string, error) {
	home := codexHome
	if home == "" {
		home = os.Getenv("CODEX_HOME")
	}
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve user home: %w", err)
		}
		home = filepath.Join(userHome, ".codex")
	}
	return filepath.Join(home, "skills", bundledSkillName), nil
}

func ensureReplaceableTarget(target string, force bool) error {
	info, err := os.Lstat(target)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("inspect target: %w", err)
	}
	if !force {
		return fmt.Errorf("target already exists: %s (use --force to replace)", target)
	}
	if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
		return os.RemoveAll(target)
	}
	return os.Remove(target)
}

func writeBundledSkill(target string) error {
	if err := os.MkdirAll(filepath.Join(target, "agents"), 0o755); err != nil {
		return fmt.Errorf("create skill directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte(bundledSkillDoc), 0o644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "agents", "openai.yaml"), []byte(bundledAgentYAML), 0o644); err != nil {
		return fmt.Errorf("write agents/openai.yaml: %w", err)
	}
	return nil
}

func validateSkillSource(source string) error {
	info, err := os.Stat(source)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("skill source does not exist: %s", source)
		}
		return fmt.Errorf("inspect source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("skill source is not a directory: %s", source)
	}
	if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
		return fmt.Errorf("skill source is missing SKILL.md: %s", source)
	}
	return nil
}
