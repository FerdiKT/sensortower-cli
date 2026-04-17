package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "missing.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != defaultBaseURL {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, defaultBaseURL)
	}
	if cfg.TimeoutSeconds != defaultTimeoutSeconds {
		t.Fatalf("TimeoutSeconds = %d, want %d", cfg.TimeoutSeconds, defaultTimeoutSeconds)
	}
	if cfg.Output != defaultOutput {
		t.Fatalf("Output = %q, want %q", cfg.Output, defaultOutput)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"base_url":"https://example.com","timeout_seconds":10,"output":"json","cookie":"file-cookie","headers":{"X-From":"file"}}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv("SENSORTOWER_BASE_URL", "https://override.example.com")
	t.Setenv("SENSORTOWER_TIMEOUT_SECONDS", "45")
	t.Setenv("SENSORTOWER_OUTPUT", "table")
	t.Setenv("SENSORTOWER_COOKIE", "session=abc")
	t.Setenv("SENSORTOWER_HEADERS_JSON", `{"X-Test":"1","Authorization":"Bearer token"}`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "https://override.example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.TimeoutSeconds != 45 {
		t.Fatalf("TimeoutSeconds = %d", cfg.TimeoutSeconds)
	}
	if cfg.Output != "table" {
		t.Fatalf("Output = %q", cfg.Output)
	}
	if cfg.Cookie != "session=abc" {
		t.Fatalf("Cookie = %q", cfg.Cookie)
	}
	if cfg.Headers["Authorization"] != "Bearer token" {
		t.Fatalf("Headers[Authorization] = %q", cfg.Headers["Authorization"])
	}
}
