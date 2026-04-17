package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	SavedAt time.Time `json:"saved_at"`
	Body    []byte    `json:"body"`
}

type Cache struct {
	dir string
	ttl time.Duration
}

func New(dir string, ttl time.Duration) *Cache {
	return &Cache{dir: dir, ttl: ttl}
}

func DefaultDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "sensortower"), nil
}

func (c *Cache) Get(key string) ([]byte, bool, error) {
	if c == nil || c.ttl <= 0 || c.dir == "" {
		return nil, false, nil
	}
	path := filepath.Join(c.dir, hash(key)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, err
	}
	if time.Since(entry.SavedAt) > c.ttl {
		return nil, false, nil
	}
	return entry.Body, true, nil
}

func (c *Cache) Put(key string, body []byte) error {
	if c == nil || c.ttl <= 0 || c.dir == "" {
		return nil
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	entry := Entry{SavedAt: time.Now().UTC(), Body: body}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, hash(key)+".json"), data, 0o644)
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func Key(parts ...string) string {
	return fmt.Sprint(parts)
}
