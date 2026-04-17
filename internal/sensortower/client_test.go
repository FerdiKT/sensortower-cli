package sensortower

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ferdikt/sensortower-cli/internal/cache"
)

func TestPublisherAppsURLConstruction(t *testing.T) {
	var gotPath string
	var gotQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"meta":{"count":0},"data":[]}`))
	}))
	defer server.Close()

	client := NewClient(Options{BaseURL: server.URL, TimeoutSeconds: 5})
	if _, _, err := client.PublisherApps(context.Background(), 1619264551, 25, 10, "downloads"); err != nil {
		t.Fatalf("PublisherApps() error = %v", err)
	}

	if gotPath != "/api/ios/publishers/1619264551/apps" {
		t.Fatalf("path = %q", gotPath)
	}
	for _, part := range []string{"limit=25", "offset=10", "sort_by=downloads"} {
		if !strings.Contains(gotQuery, part) {
			t.Fatalf("query %q missing %q", gotQuery, part)
		}
	}
}

func TestAppDetailsURLConstruction(t *testing.T) {
	var gotPath string
	var gotCountry string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotCountry = r.URL.Query().Get("country")
		_, _ = w.Write([]byte(`{"app_id":6478631467,"name":"Playlist Transfer","publisher_name":"Virals"}`))
	}))
	defer server.Close()

	client := NewClient(Options{BaseURL: server.URL, TimeoutSeconds: 5})
	if _, _, err := client.AppDetails(context.Background(), 6478631467, "US"); err != nil {
		t.Fatalf("AppDetails() error = %v", err)
	}

	if gotPath != "/api/ios/apps/6478631467" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotCountry != "US" {
		t.Fatalf("country = %q", gotCountry)
	}
}

func TestCategoryRankingsURLConstruction(t *testing.T) {
	var gotQuery = map[string]string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range r.URL.Query() {
			gotQuery[key] = values[0]
		}
		_, _ = w.Write([]byte(`{"data":{"free":[],"grossing":[],"paid":[]},"date":"2026-04-16","total_count":0,"offset":0,"limit":25}`))
	}))
	defer server.Close()

	client := NewClient(Options{BaseURL: server.URL, TimeoutSeconds: 5})
	if _, _, err := client.CategoryRankings(context.Background(), "US", 0, "2026-04-16", "iphone", 25, 0); err != nil {
		t.Fatalf("CategoryRankings() error = %v", err)
	}

	assertQuery(t, gotQuery, "country", "US")
	assertQuery(t, gotQuery, "category", "0")
	assertQuery(t, gotQuery, "date", "2026-04-16")
	assertQuery(t, gotQuery, "device", "iphone")
	assertQuery(t, gotQuery, "limit", "25")
	assertQuery(t, gotQuery, "offset", "0")
}

func TestClientInjectsCookieAndHeaders(t *testing.T) {
	var gotCookie string
	var gotHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		gotHeader = r.Header.Get("X-Test")
		_, _ = w.Write([]byte(`{"meta":{"count":0},"data":[]}`))
	}))
	defer server.Close()

	client := NewClient(Options{
		BaseURL:        server.URL,
		TimeoutSeconds: 5,
		Cookie:         "session=abc",
		Headers:        map[string]string{"X-Test": "1"},
	})
	if _, _, err := client.PublisherApps(context.Background(), 1, 25, 0, "downloads"); err != nil {
		t.Fatalf("PublisherApps() error = %v", err)
	}

	if gotCookie != "session=abc" {
		t.Fatalf("cookie = %q", gotCookie)
	}
	if gotHeader != "1" {
		t.Fatalf("header = %q", gotHeader)
	}
}

func TestFixtureDecoding(t *testing.T) {
	t.Run("publisher apps", func(t *testing.T) {
		data := mustReadFixture(t, "publisher_apps.json")
		var resp PublisherAppsResponse
		if err := jsonUnmarshal(data, &resp); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if len(resp.Data) == 0 {
			t.Fatal("expected publisher apps data")
		}
	})

	t.Run("app details", func(t *testing.T) {
		data := mustReadFixture(t, "app_details.json")
		var resp AppDetails
		if err := jsonUnmarshal(data, &resp); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if resp.AppID == 0 || resp.Name == "" {
			t.Fatalf("unexpected app details: %+v", resp)
		}
	})

	t.Run("category rankings", func(t *testing.T) {
		data := mustReadFixture(t, "category_rankings.json")
		var resp CategoryRankingsResponse
		if err := jsonUnmarshal(data, &resp); err != nil {
			t.Fatalf("unmarshal error = %v", err)
		}
		if len(resp.Data.Free) == 0 {
			t.Fatal("expected free rankings data")
		}
	})
}

func TestAppDetailsRetries429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"slow down"}`))
			return
		}
		_, _ = w.Write([]byte(`{"app_id":6478631467,"name":"Playlist Transfer","publisher_name":"Virals"}`))
	}))
	defer server.Close()

	client := NewClient(Options{
		BaseURL:        server.URL,
		TimeoutSeconds: 5,
		Retry429:       true,
		RetryMax:       2,
		RetryWait:      1,
	})
	resp, meta, err := client.AppDetails(context.Background(), 6478631467, "US")
	if err != nil {
		t.Fatalf("AppDetails() error = %v", err)
	}
	if resp.AppID != 6478631467 {
		t.Fatalf("app id = %d", resp.AppID)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if meta == nil || meta.Retried != 1 {
		t.Fatalf("meta retried = %+v", meta)
	}
	if meta.RateLimitHeaders["x_ratelimit_remaining"] != "0" {
		t.Fatalf("rate limit headers = %+v", meta.RateLimitHeaders)
	}
}

func TestAppDetailsUsesCache(t *testing.T) {
	cacheDir := t.TempDir()
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = w.Write([]byte(`{"app_id":6478631467,"name":"Playlist Transfer","publisher_name":"Virals"}`))
	}))
	defer server.Close()

	client := NewClient(Options{
		BaseURL:        server.URL,
		TimeoutSeconds: 5,
		Cache:          cache.New(cacheDir, 5*time.Second),
	})
	if _, meta, err := client.AppDetails(context.Background(), 6478631467, "US"); err != nil {
		t.Fatalf("first AppDetails() error = %v", err)
	} else if meta == nil || meta.Cached {
		t.Fatalf("first meta = %+v", meta)
	}
	if _, meta, err := client.AppDetails(context.Background(), 6478631467, "US"); err != nil {
		t.Fatalf("second AppDetails() error = %v", err)
	} else if meta == nil || !meta.Cached {
		t.Fatalf("second meta = %+v", meta)
	}
	if hits != 1 {
		t.Fatalf("hits = %d, want 1", hits)
	}
}

func assertQuery(t *testing.T, query map[string]string, key, want string) {
	t.Helper()
	if got := query[key]; got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return data
}

func jsonUnmarshal(data []byte, dst any) error {
	return json.Unmarshal(data, dst)
}
