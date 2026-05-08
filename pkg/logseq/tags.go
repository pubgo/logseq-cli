package logseq

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// GetAllTags collects all tags in current graph.
// Strategy:
// 1) Prefer DatascriptQuery for direct tag extraction.
// 2) Fallback to page properties (tag/tags) to improve compatibility.
func (c *Client) GetAllTags(ctx context.Context) ([]string, error) {
	tags := make(map[string]struct{})

	raw, err := c.DatascriptQuery(ctx, "[:find ?name :where [?b :block/tags ?t] [?t :block/name ?name]]")
	if err == nil {
		for _, tag := range parseSingleColumnStrings(raw) {
			if tag == "" {
				continue
			}
			tags[normalizeTag(tag)] = struct{}{}
		}
	}

	pages, pagesErr := c.GetAllPages(ctx)
	if pagesErr != nil && len(tags) == 0 {
		return nil, pagesErr
	}

	for _, p := range pages {
		for _, tag := range extractTagsFromProperties(p.Properties) {
			if tag == "" {
				continue
			}
			tags[normalizeTag(tag)] = struct{}{}
		}
	}

	list := make([]string, 0, len(tags))
	for tag := range tags {
		if tag != "" {
			list = append(list, tag)
		}
	}
	sort.Strings(list)
	return list, nil
}

func parseSingleColumnStrings(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	var rows [][]any
	if err := json.Unmarshal(raw, &rows); err == nil {
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			if len(row) == 0 {
				continue
			}
			if v := strings.TrimSpace(fmt.Sprintf("%v", row[0])); v != "" {
				out = append(out, v)
			}
		}
		return out
	}

	var flat []any
	if err := json.Unmarshal(raw, &flat); err == nil {
		out := make([]string, 0, len(flat))
		for _, item := range flat {
			if v := strings.TrimSpace(fmt.Sprintf("%v", item)); v != "" {
				out = append(out, v)
			}
		}
		return out
	}

	return nil
}

func extractTagsFromProperties(props map[string]any) []string {
	if len(props) == 0 {
		return nil
	}

	candidates := make([]string, 0)
	for k, v := range props {
		key := strings.ToLower(strings.TrimSpace(k))
		if key != "tag" && key != "tags" {
			continue
		}
		candidates = append(candidates, anyToStrings(v)...)
	}

	return dedupeTagStrings(candidates)
}

func anyToStrings(v any) []string {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		parts := strings.Split(x, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) > 0 {
			return out
		}
		if strings.TrimSpace(x) == "" {
			return nil
		}
		return []string{strings.TrimSpace(x)}
	case []string:
		out := make([]string, 0, len(x))
		for _, item := range x {
			if s := strings.TrimSpace(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	case []any:
		out := make([]string, 0, len(x))
		for _, item := range x {
			out = append(out, anyToStrings(item)...)
		}
		return out
	default:
		return []string{fmt.Sprintf("%v", x)}
	}
}

func normalizeTag(s string) string {
	v := strings.TrimSpace(strings.ToLower(s))
	v = strings.TrimPrefix(v, "#")
	return strings.TrimSpace(v)
}

func dedupeTagStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		tag := normalizeTag(s)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}
