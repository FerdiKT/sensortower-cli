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
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	cookie     string
	headers    map[string]string
	userAgent  string
}

type Options struct {
	BaseURL        string
	TimeoutSeconds int
	Cookie         string
	Headers        map[string]string
}

type HTTPError struct {
	StatusCode int
	Body       string
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
	}
}

func (c *Client) PublisherApps(ctx context.Context, publisherID int64, limit, offset int, sortBy string) (*PublisherAppsResponse, error) {
	path := fmt.Sprintf("/api/ios/publishers/%d/apps", publisherID)
	var out PublisherAppsResponse
	err := c.getJSON(ctx, path, map[string]string{
		"limit":   strconv.Itoa(limit),
		"offset":  strconv.Itoa(offset),
		"sort_by": sortBy,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) AppDetails(ctx context.Context, appID int64, country string) (*AppDetails, error) {
	path := fmt.Sprintf("/api/ios/apps/%d", appID)
	var out AppDetails
	err := c.getJSON(ctx, path, map[string]string{
		"country": country,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CategoryRankings(ctx context.Context, country string, category int, date, device string, limit, offset int) (*CategoryRankingsResponse, error) {
	var out CategoryRankingsResponse
	err := c.getJSON(ctx, "/api/ios/category_rankings", map[string]string{
		"country":  country,
		"category": strconv.Itoa(category),
		"date":     date,
		"device":   device,
		"limit":    strconv.Itoa(limit),
		"offset":   strconv.Itoa(offset),
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) getJSON(ctx context.Context, path string, query map[string]string, dst any) error {
	endpoint, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("build url: %w", err)
	}

	values := endpoint.Query()
	for k, v := range query {
		if strings.TrimSpace(v) != "" {
			values.Set(k, v)
		}
	}
	endpoint.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
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
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return &HTTPError{StatusCode: resp.StatusCode, Body: msg}
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
