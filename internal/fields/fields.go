package fields

import (
	"strings"
)

func FilterMap(raw map[string]any, fields []string) map[string]any {
	if len(fields) == 0 {
		return raw
	}
	out := map[string]any{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if value, ok := getPath(raw, field); ok {
			out[field] = value
		}
	}
	return out
}

func Parse(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getPath(raw map[string]any, path string) (any, bool) {
	cur := any(raw)
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[part]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
