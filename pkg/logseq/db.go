package logseq

import (
	"context"
	"encoding/json"
)

// DatascriptQuery executes a Datalog query against the Logseq database.
func (c *Client) DatascriptQuery(ctx context.Context, query string, inputs ...any) (json.RawMessage, error) {
	args := []any{query}
	args = append(args, inputs...)
	return c.CallAPI(ctx, "logseq.DB.datascriptQuery", args...)
}

// DSLQuery executes a Logseq DSL query.
func (c *Client) DSLQuery(ctx context.Context, query string) (json.RawMessage, error) {
	return c.CallAPI(ctx, "logseq.DB.q", query)
}
