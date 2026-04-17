package sensortower

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ferdikt/sensortower-cli/internal/cache"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cookie     string
	headers    map[string]string
	userAgent  string
	retry429   bool
	retryMax   int
	retryWait  int
	cache      *cache.Cache
}

type Options struct {
	BaseURL        string
	TimeoutSeconds int
	Cookie         string
	Headers        map[string]string
	Retry429       bool
	RetryMax       int
	RetryWait      int
	Cache          *cache.Cache
}

type HTTPError struct {
	StatusCode        int
	Body              string
	RetryAfterSeconds int
	RateLimitHeaders  map[string]string
	URL               string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("request failed with status %d: %s", e.StatusCode, e.Body)
}

func NewClient(opts Options) *Client {
	baseURL := strings.TrimRight(opts.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://app.sensortower.com"
	}

	timeout := time.Duration(opts.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	headers := map[string]string{}
	for k, v := range opts.Headers {
		headers[k] = v
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		cookie:    opts.Cookie,
		headers:   headers,
		userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
		retry429:  opts.Retry429,
		retryMax:  opts.RetryMax,
		retryWait: opts.RetryWait,
		cache:     opts.Cache,
	}
}

func (c *Client) PublisherApps(ctx context.Context, publisherID int64, limit, offset int, sortBy string) (*PublisherAppsResponse, *ResponseMeta, error) {
	path := fmt.Sprintf("/api/ios/publishers/%d/apps", publisherID)
	var out PublisherAppsResponse
	meta, err := c.getJSON(ctx, path, map[string]string{
		"limit":   strconv.Itoa(limit),
		"offset":  strconv.Itoa(offset),
		"sort_by": sortBy,
	}, &out)
	if err != nil {
		return nil, nil, err
	}
	return &out, meta, nil
}

func (c *Client) AppDetails(ctx context.Context, appID int64, country string) (*AppDetails, *ResponseMeta, error) {
	path := fmt.Sprintf("/api/ios/apps/%d", appID)
	var out AppDetails
	meta, err := c.getJSON(ctx, path, map[string]string{
		"country": country,
	}, &out)
	if err != nil {
		return nil, nil, err
	}
	return &out, meta, nil
}

func (c *Client) CategoryRankings(ctx context.Context, country string, category int, date, device string, limit, offset int) (*CategoryRankingsResponse, *ResponseMeta, error) {
	var out CategoryRankingsResponse
	meta, err := c.getJSON(ctx, "/api/ios/category_rankings", map[string]string{
		"country":  country,
		"category": strconv.Itoa(category),
		"date":     date,
		"device":   device,
		"limit":    strconv.Itoa(limit),
		"offset":   strconv.Itoa(offset),
	}, &out)
	if err != nil {
		return nil, nil, err
	}
	return &out, meta, nil
}

func (c *Client) AppAutocomplete(ctx context.Context, term, os string, limit int, expandEntities, flags, markUsageDisabledApps bool) ([]map[string]any, *ResponseMeta, error) {
	var out autocompleteSearchResponse
	meta, err := c.getJSON(ctx, "/api/autocomplete_search", map[string]string{
		"entity_type":              "app",
		"term":                     term,
		"os":                       os,
		"limit":                    strconv.Itoa(limit),
		"expand_entities":          strconv.FormatBool(expandEntities),
		"flags":                    strconv.FormatBool(flags),
		"mark_usage_disabled_apps": strconv.FormatBool(markUsageDisabledApps),
	}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out.Data.Entities, meta, nil
}

func (c *Client) PublisherAutocomplete(ctx context.Context, term, os string, limit int, includeExtendedInfo bool) ([]map[string]any, *ResponseMeta, error) {
	var out autocompleteSearchResponse
	meta, err := c.getJSON(ctx, "/api/autocomplete_search", map[string]string{
		"entity_type":           "publisher",
		"term":                  term,
		"os":                    os,
		"limit":                 strconv.Itoa(limit),
		"include_extended_info": strconv.FormatBool(includeExtendedInfo),
	}, &out)
	if err != nil {
		return nil, nil, err
	}
	return out.Data.Entities, meta, nil
}

type autocompleteSearchResponse struct {
	Data struct {
		Entities []map[string]any `json:"entities"`
	} `json:"data"`
}

func (c *Client) getJSON(ctx context.Context, path string, query map[string]string, dst any) (*ResponseMeta, error) {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("build url: %w", err)
	}

	values := endpoint.Query()
	for k, v := range query {
		if strings.TrimSpace(v) != "" {
			values.Set(k, v)
		}
	}
	endpoint.RawQuery = values.Encode()
	meta := &ResponseMeta{RequestURL: endpoint.String()}
	cacheKey := cache.Key(endpoint.String(), c.cookie)
	if c.cache != nil {
		if body, ok, err := c.cache.Get(cacheKey); err == nil && ok {
			meta.Cached = true
			if err := json.Unmarshal(body, dst); err != nil {
				return nil, fmt.Errorf("decode cached response: %w", err)
			}
			return meta, nil
		}
	}

	var body []byte
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		if c.cookie != "" {
			req.Header.Set("Cookie", c.cookie)
		}
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("send request: %w", err)
		}

		body, err = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && c.retry429 && attempt < c.retryMax {
			waitSeconds := retryAfterSeconds(resp.Header.Get("Retry-After"), c.retryWait, attempt)
			meta.Retried++
			meta.RetryAfterSeconds = waitSeconds
			meta.RateLimitHeaders = mergeRateLimitHeaders(meta.RateLimitHeaders, rateLimitHeaders(resp.Header))
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := strings.TrimSpace(string(body))
			if len(msg) > 300 {
				msg = msg[:300]
			}
			return nil, &HTTPError{
				StatusCode:        resp.StatusCode,
				Body:              msg,
				RetryAfterSeconds: retryAfterSeconds(resp.Header.Get("Retry-After"), 0, 0),
				RateLimitHeaders:  rateLimitHeaders(resp.Header),
				URL:               endpoint.String(),
			}
		}
		meta.RateLimitHeaders = mergeRateLimitHeaders(meta.RateLimitHeaders, rateLimitHeaders(resp.Header))
		break
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if c.cache != nil {
		_ = c.cache.Put(cacheKey, body)
	}
	return meta, nil
}

func retryAfterSeconds(header string, fallback, attempt int) int {
	if header != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(header)); err == nil && n > 0 {
			return n
		}
	}
	if fallback > 0 {
		return fallback
	}
	seconds := 1 << attempt
	if seconds > 60 {
		seconds = 60
	}
	if seconds < 1 {
		seconds = 1
	}
	return seconds
}

func rateLimitHeaders(h http.Header) map[string]string {
	out := map[string]string{}
	for _, key := range []string{"Retry-After", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"} {
		if value := h.Get(key); value != "" {
			out[strings.ToLower(strings.ReplaceAll(key, "-", "_"))] = value
		}
	}
	return out
}

func mergeRateLimitHeaders(existing, incoming map[string]string) map[string]string {
	if len(existing) == 0 {
		return incoming
	}
	if len(incoming) == 0 {
		return existing
	}
	out := map[string]string{}
	for k, v := range existing {
		out[k] = v
	}
	for k, v := range incoming {
		out[k] = v
	}
	return out
}
