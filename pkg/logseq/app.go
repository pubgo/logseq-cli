package logseq

import (
	"context"
	"encoding/json"
)

// GetCurrentGraph returns metadata about the current graph.
func (c *Client) GetCurrentGraph(ctx context.Context) (*GraphInfo, error) {
	raw, err := c.CallAPI(ctx, "logseq.App.getCurrentGraph")
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var info GraphInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// GetUserConfigs returns the user's Logseq configuration.
func (c *Client) GetUserConfigs(ctx context.Context) (map[string]any, error) {
	return decode[map[string]any](c.CallAPI(ctx, "logseq.App.getUserConfigs"))
}

// Search performs a full-text search across the graph.
// Note: This method may not be available in all Logseq versions.
func (c *Client) Search(ctx context.Context, query string) (*SearchResult, error) {
	raw, err := c.CallAPI(ctx, "logseq.App.search", query)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var result SearchResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
