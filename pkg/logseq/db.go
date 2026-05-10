package logseq

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var limitRegexp = regexp.MustCompile(`:limit\s+(\d+)`)
var orderByRegexp = regexp.MustCompile(`:order-by\s+\[.*?\]`)

// DatascriptQuery executes a Datalog query against the Logseq database.
// Note: Datascript does not support :limit or :order-by clauses natively.
// This method strips those clauses before sending to the API and applies
// :limit as client-side truncation on the results.
func (c *Client) DatascriptQuery(ctx context.Context, query string, inputs ...any) (json.RawMessage, error) {
	// Extract and strip :limit N (Datascript ignores it, but may error)
	clientLimit := 0
	if m := limitRegexp.FindStringSubmatch(query); len(m) == 2 {
		clientLimit, _ = strconv.Atoi(m[1])
	}
	cleanQuery := limitRegexp.ReplaceAllString(query, "")
	cleanQuery = orderByRegexp.ReplaceAllString(cleanQuery, "")

	args := []any{cleanQuery}
	args = append(args, inputs...)
	raw, err := c.CallAPI(ctx, "logseq.DB.datascriptQuery", args...)
	if err != nil {
		return nil, err
	}

	// Apply client-side limit
	if clientLimit > 0 && len(raw) > 0 {
		var arr []json.RawMessage
		if json.Unmarshal(raw, &arr) == nil && len(arr) > clientLimit {
			arr = arr[:clientLimit]
			if truncated, err := json.Marshal(arr); err == nil {
				return truncated, nil
			}
		}
	}

	return raw, nil
}

// DSLQuery executes a Logseq DSL query.
func (c *Client) DSLQuery(ctx context.Context, query string) (json.RawMessage, error) {
	raw, err := c.CallAPI(ctx, "logseq.DB.q", query)
	if err == nil {
		return raw, nil
	}

	if !strings.Contains(strings.ToLower(err.Error()), "methodnotexist") {
		return nil, err
	}

	// Compatibility fallback for older/newer Logseq variants.
	raw2, err2 := c.CallAPI(ctx, "logseq.DB.dslQuery", query)
	if err2 == nil {
		return raw2, nil
	}

	return nil, err
}
