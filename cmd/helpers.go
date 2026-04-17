package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ferdikt/sensortower-cli/internal/export"
	"github.com/ferdikt/sensortower-cli/internal/output"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
)

func writeOutput(v any) error {
	return writeOutputWithMeta(v, nil)
}

func writeOutputWithMeta(v any, meta *sensortower.ResponseMeta) error {
	w, err := export.Writer(opts.OutputFile)
	if err != nil {
		return err
	}
	defer w.Close()

	switch opts.Output {
	case "json":
		if meta != nil {
			return output.RenderJSON(w, map[string]any{"data": v, "meta": meta})
		}
		return output.RenderJSON(w, v)
	case "jsonl":
		return export.WriteJSONL(w, export.Flatten(structToAny(v)))
	case "csv":
		return export.WriteCSV(w, export.Flatten(structToAny(v)))
	case "table":
		return output.RenderJSON(w, v)
	default:
		return fmt.Errorf("unsupported output format: %s", opts.Output)
	}
}

func outputWriter() (io.WriteCloser, error) {
	return export.Writer(opts.OutputFile)
}

func structToMap(v any) map[string]any {
	var out map[string]any
	b, _ := json.Marshal(v)
	_ = json.Unmarshal(b, &out)
	return out
}

func structToAny(v any) any {
	var out any
	b, _ := json.Marshal(v)
	_ = json.Unmarshal(b, &out)
	return out
}

func emitMeta(meta *sensortower.ResponseMeta) {
	if meta == nil {
		return
	}
	if meta.Cached || meta.Retried > 0 || meta.RetryAfterSeconds > 0 || len(meta.RateLimitHeaders) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "retry_after_seconds=%d rate_limit_headers=%v request_url=%s cached=%v retried=%d\n",
			meta.RetryAfterSeconds, meta.RateLimitHeaders, meta.RequestURL, meta.Cached, meta.Retried)
	}
}

func readInt64Lines(path string) ([]int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id %q: %w", line, err)
		}
		out = append(out, n)
	}
	return out, scanner.Err()
}

func parseHeadersJSON(s string) (map[string]string, error) {
	if strings.TrimSpace(s) == "" {
		return map[string]string{}, nil
	}
	var out map[string]string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseInts(s string) ([]int, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, part := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func appendUnique(in []string, value string) []string {
	for _, existing := range in {
		if existing == value {
			return in
		}
	}
	return append(in, value)
}

func appendUniqueInt(in []int, value int) []int {
	for _, existing := range in {
		if existing == value {
			return in
		}
	}
	return append(in, value)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
