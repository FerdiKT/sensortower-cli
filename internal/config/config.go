package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultBaseURL        = "https://app.sensortower.com"
	defaultTimeoutSeconds = 30
	defaultOutput         = "table"
)

type Config struct {
	BaseURL         string             `json:"base_url,omitempty"`
	TimeoutSeconds  int                `json:"timeout_seconds,omitempty"`
	Output          string             `json:"output,omitempty"`
	Cookie          string             `json:"cookie,omitempty"`
	Headers         map[string]string  `json:"headers,omitempty"`
	CacheTTLSeconds int                `json:"cache_ttl_seconds,omitempty"`
	ActiveContext   string             `json:"active_context,omitempty"`
	Contexts        map[string]Context `json:"contexts,omitempty"`
}

type Context struct {
	BaseURL        string            `json:"base_url,omitempty"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty"`
	Output         string            `json:"output,omitempty"`
	Cookie         string            `json:"cookie,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, "sensortower", "config.json"), nil
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is empty")
	}

	cfg := &Config{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg.applyEnv()
	cfg.normalize()
	return cfg, nil
}

func (c *Config) Save(path string) error {
	if path == "" {
		return errors.New("config path is empty")
	}
	c.normalize()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func (c *Config) normalize() {
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	c.BaseURL = strings.TrimRight(c.BaseURL, "/")
	if c.TimeoutSeconds <= 0 {
		c.TimeoutSeconds = defaultTimeoutSeconds
	}
	if c.Output == "" {
		c.Output = defaultOutput
	}
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	if c.CacheTTLSeconds < 0 {
		c.CacheTTLSeconds = 0
	}
	if c.Contexts == nil {
		c.Contexts = map[string]Context{}
	}
	for name, ctx := range c.Contexts {
		if ctx.BaseURL == "" {
			ctx.BaseURL = c.BaseURL
		}
		ctx.BaseURL = strings.TrimRight(ctx.BaseURL, "/")
		if ctx.TimeoutSeconds <= 0 {
			ctx.TimeoutSeconds = c.TimeoutSeconds
		}
		if ctx.Output == "" {
			ctx.Output = c.Output
		}
		if ctx.Headers == nil {
			ctx.Headers = map[string]string{}
		}
		c.Contexts[name] = ctx
	}
}

func (c *Config) applyEnv() {
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_BASE_URL")); v != "" {
		c.BaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_TIMEOUT_SECONDS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.TimeoutSeconds = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_OUTPUT")); v != "" {
		c.Output = v
	}
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_CACHE_TTL_SECONDS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.CacheTTLSeconds = n
		}
	}
	if v := os.Getenv("SENSORTOWER_COOKIE"); strings.TrimSpace(v) != "" {
		c.Cookie = v
	}
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_HEADERS_JSON")); v != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(v), &headers); err == nil {
			c.Headers = headers
		}
	}
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_CONTEXT")); v != "" {
		c.ActiveContext = v
	}
}

func (c *Config) EffectiveContext(name string) Context {
	if name == "" {
		name = c.ActiveContext
	}
	if ctx, ok := c.Contexts[name]; ok {
		return ctx
	}
	return Context{
		BaseURL:        c.BaseURL,
		TimeoutSeconds: c.TimeoutSeconds,
		Output:         c.Output,
		Cookie:         c.Cookie,
		Headers:        c.Headers,
	}
}

func (c *Config) SetContext(name string, ctx Context) {
	if c.Contexts == nil {
		c.Contexts = map[string]Context{}
	}
	if ctx.Headers == nil {
		ctx.Headers = map[string]string{}
	}
	c.Contexts[name] = ctx
}
