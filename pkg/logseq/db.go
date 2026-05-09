package logseq

import (
	"context"
	"encoding/json"
	"strings"
)

// DatascriptQuery executes a Datalog query against the Logseq database.
func (c *Client) DatascriptQuery(ctx context.Context, query string, inputs ...any) (json.RawMessage, error) {
	args := []any{query}
	args = append(args, inputs...)
	return c.CallAPI(ctx, "logseq.DB.datascriptQuery", args...)
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
