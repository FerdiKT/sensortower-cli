package cmd

import (
	"strings"
	"testing"

	"github.com/ferdikt/sensortower-cli/internal/sensortower"
)

func TestEnrichWarningFromHTTPError(t *testing.T) {
	warning := enrichWarningFromError(123, &sensortower.HTTPError{
		StatusCode:        429,
		Body:              "rate limited",
		RetryAfterSeconds: 15,
		URL:               "https://example.test/app",
	})

	if warning.AppID != 123 {
		t.Fatalf("app_id = %d, want 123", warning.AppID)
	}
	if warning.StatusCode != 429 {
		t.Fatalf("status_code = %d, want 429", warning.StatusCode)
	}
	if warning.RetryAfterSeconds != 15 {
		t.Fatalf("retry_after_seconds = %d, want 15", warning.RetryAfterSeconds)
	}
	if warning.URL == "" {
		t.Fatal("expected warning URL")
	}
}

func TestSampleEnrichWarnings(t *testing.T) {
	sample := sampleEnrichWarnings([]enrichWarning{
		{AppID: 123, StatusCode: 429, RetryAfterSeconds: 15},
		{AppID: 456, Error: "network error"},
	})

	for _, want := range []string{"123(status=429,retry_after=15s)", "456(network error)"} {
		if !strings.Contains(sample, want) {
			t.Fatalf("sample %q missing %q", sample, want)
		}
	}
}
