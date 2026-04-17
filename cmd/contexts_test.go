package cmd

import (
	"testing"

	"github.com/ferdikt/sensortower-cli/internal/config"
)

func TestBuildContextRowsIncludesDefault(t *testing.T) {
	rows := buildContextRows(&config.Config{
		BaseURL:        "https://app.sensortower.com",
		TimeoutSeconds: 30,
		Output:         "json",
	})
	if len(rows) != 1 {
		t.Fatalf("rows len = %d, want 1", len(rows))
	}
	if rows[0]["name"] != "default" {
		t.Fatalf("default row = %+v", rows[0])
	}
	if rows[0]["kind"] != "default" {
		t.Fatalf("kind = %v", rows[0]["kind"])
	}
}
