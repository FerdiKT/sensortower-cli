package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ferdikt/sensortower-cli/internal/sensortower"
)

func TestRenderPublisherAppsTable(t *testing.T) {
	resp := &sensortower.PublisherAppsResponse{
		Data: []sensortower.PublisherApp{
			{
				AppID:                        1,
				HumanizedName:                "Playlist Transfer",
				PublisherName:                "Virals",
				UpdatedDate:                  "2026-03-09T00:00:00Z",
				WorldwideLast30DaysDownloads: sensortower.HumanizedMetric{String: "10k"},
				WorldwideLast30DaysRevenue:   sensortower.HumanizedMetric{String: "$60k"},
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderPublisherAppsTable(&buf, resp); err != nil {
		t.Fatalf("RenderPublisherAppsTable() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"APP ID", "Playlist Transfer", "Virals", "10k", "$60k"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %s", want, out)
		}
	}
}

func TestRenderCategoryRankingsTable(t *testing.T) {
	resp := &sensortower.CategoryRankingsResponse{
		Data: sensortower.CategoryRankingBuckets{
			Free: []sensortower.CategoryRankingEntry{
				{
					Rank:                        1,
					PreviousRank:                2,
					AppID:                       6448311069,
					Name:                        "ChatGPT",
					PublisherName:               "OpenAI OpCo, LLC",
					WorldwideLastMonthDownloads: sensortower.HumanizedMetric{String: "21m"},
					WorldwideLastMonthRevenue:   sensortower.HumanizedMetric{String: "$248m"},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderCategoryRankingsTable(&buf, resp); err != nil {
		t.Fatalf("RenderCategoryRankingsTable() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"FREE", "ChatGPT", "OpenAI OpCo, LLC", "21m", "$248m"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q: %s", want, out)
		}
	}
}
