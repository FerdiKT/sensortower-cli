package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentInstallSkillWritesBundledFiles(t *testing.T) {
	codexHome := t.TempDir()
	cmd := newAgentInstallSkillCommand()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--codex-home", codexHome})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "skill installed:") {
		t.Fatalf("stdout = %q, want install message", stdout.String())
	}

	skillDir := filepath.Join(codexHome, "skills", bundledSkillName)
	skillDoc, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("ReadFile(SKILL.md) error = %v", err)
	}
	if !strings.Contains(string(skillDoc), "Use this skill for repository-local Sensor Tower CLI work.") {
		t.Fatalf("unexpected SKILL.md content: %s", string(skillDoc))
	}

	agentYAML, err := os.ReadFile(filepath.Join(skillDir, "agents", "openai.yaml"))
	if err != nil {
		t.Fatalf("ReadFile(openai.yaml) error = %v", err)
	}
	if !strings.Contains(string(agentYAML), "display_name: Sensor Tower CLI") {
		t.Fatalf("unexpected openai.yaml content: %s", string(agentYAML))
	}
}

func TestAgentLinkSkillCreatesSymlink(t *testing.T) {
	sourceRoot := t.TempDir()
	sourceDir := filepath.Join(sourceRoot, bundledSkillName)
	if err := os.MkdirAll(filepath.Join(sourceDir, "agents"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "SKILL.md"), []byte("test skill"), 0o644); err != nil {
		t.Fatalf("WriteFile(SKILL.md) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "agents", "openai.yaml"), []byte("display_name: Test"), 0o644); err != nil {
		t.Fatalf("WriteFile(openai.yaml) error = %v", err)
	}

	codexHome := t.TempDir()
	cmd := newAgentLinkSkillCommand()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--codex-home", codexHome, "--source", sourceDir})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	targetDir := filepath.Join(codexHome, "skills", bundledSkillName)
	info, err := os.Lstat(targetDir)
	if err != nil {
		t.Fatalf("Lstat() error = %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("target is not a symlink: mode=%v", info.Mode())
	}

	resolved, err := os.Readlink(targetDir)
	if err != nil {
		t.Fatalf("Readlink() error = %v", err)
	}
	if resolved != sourceDir {
		t.Fatalf("symlink target = %q, want %q", resolved, sourceDir)
	}
}

func TestAgentShowSkillPath(t *testing.T) {
	codexHome := t.TempDir()
	cmd := newAgentShowSkillPathCommand()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--codex-home", codexHome})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := filepath.Join(codexHome, "skills", bundledSkillName)
	if strings.TrimSpace(stdout.String()) != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}
