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
	if v := os.Getenv("SENSORTOWER_COOKIE"); strings.TrimSpace(v) != "" {
		c.Cookie = v
	}
	if v := strings.TrimSpace(os.Getenv("SENSORTOWER_HEADERS_JSON")); v != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(v), &headers); err == nil {
			c.Headers = headers
		}
	}
}
