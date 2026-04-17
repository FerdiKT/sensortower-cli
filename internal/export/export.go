package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func Writer(path string) (io.WriteCloser, error) {
	if path == "" {
		return nopCloser{os.Stdout}, nil
	}
	return os.Create(path)
}

type nopCloser struct{ io.Writer }

func (n nopCloser) Close() error { return nil }

func WriteJSONL(w io.Writer, rows []map[string]any) error {
	enc := json.NewEncoder(w)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	return nil
}

func WriteCSV(w io.Writer, rows []map[string]any) error {
	cw := csv.NewWriter(w)
	headers := orderedHeaders(rows)
	if err := cw.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, header := range headers {
			record[i] = stringify(row[header])
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

func Flatten(v any) []map[string]any {
	switch x := v.(type) {
	case []map[string]any:
		return x
	case map[string]any:
		return []map[string]any{flattenMap("", x)}
	case []any:
		rows := make([]map[string]any, 0, len(x))
		for _, item := range x {
			if m, ok := item.(map[string]any); ok {
				rows = append(rows, flattenMap("", m))
			}
		}
		return rows
	default:
		return []map[string]any{{"value": x}}
	}
}

func flattenMap(prefix string, in map[string]any) map[string]any {
	out := map[string]any{}
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := in[key]
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		if nested, ok := value.(map[string]any); ok {
			for k, v := range flattenMap(fullKey, nested) {
				out[k] = v
			}
			continue
		}
		out[fullKey] = value
	}
	return out
}

func orderedHeaders(rows []map[string]any) []string {
	set := map[string]struct{}{}
	for _, row := range rows {
		for key := range row {
			set[key] = struct{}{}
		}
	}
	headers := make([]string, 0, len(set))
	for key := range set {
		headers = append(headers, key)
	}
	sort.Strings(headers)
	return headers
}

func stringify(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case []string:
		return strings.Join(x, ",")
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprint(x)
		}
		s := string(b)
		if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
			var plain string
			if err := json.Unmarshal(b, &plain); err == nil {
				return plain
			}
		}
		return s
	}
}
