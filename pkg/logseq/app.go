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

// GetInfo returns application information.
func (c *Client) GetInfo(ctx context.Context) (map[string]any, error) {
	return decode[map[string]any](c.CallAPI(ctx, "logseq.App.getInfo"))
}

// GetUserInfo returns current user information.
func (c *Client) GetUserInfo(ctx context.Context) (map[string]any, error) {
	return decode[map[string]any](c.CallAPI(ctx, "logseq.App.getUserInfo"))
}

// CheckCurrentIsDBGraph reports whether current graph is DB graph mode.
func (c *Client) CheckCurrentIsDBGraph(ctx context.Context) (bool, error) {
	return decode[bool](c.CallAPI(ctx, "logseq.App.checkCurrentIsDbGraph"))
}

// GetCurrentGraphConfigs returns graph configs by keys (or all if keys empty, depending on Logseq version behavior).
func (c *Client) GetCurrentGraphConfigs(ctx context.Context, keys ...string) (any, error) {
	args := make([]any, 0, len(keys))
	for _, k := range keys {
		args = append(args, k)
	}
	return decode[any](c.CallAPI(ctx, "logseq.App.getCurrentGraphConfigs", args...))
}

// GetCurrentGraphFavorites returns favorites of current graph.
func (c *Client) GetCurrentGraphFavorites(ctx context.Context) ([]any, error) {
	return decode[[]any](c.CallAPI(ctx, "logseq.App.getCurrentGraphFavorites"))
}

// GetCurrentGraphRecent returns recently visited pages in current graph.
func (c *Client) GetCurrentGraphRecent(ctx context.Context) ([]any, error) {
	return decode[[]any](c.CallAPI(ctx, "logseq.App.getCurrentGraphRecent"))
}

// GetCurrentGraphTemplates returns template map in current graph.
func (c *Client) GetCurrentGraphTemplates(ctx context.Context) (map[string]any, error) {
	return decode[map[string]any](c.CallAPI(ctx, "logseq.App.getCurrentGraphTemplates"))
}

// GetStateFromStore returns value from Logseq app state store by key.
func (c *Client) GetStateFromStore(ctx context.Context, key string) (any, error) {
	return decode[any](c.CallAPI(ctx, "logseq.App.getStateFromStore", key))
}

// SetStateFromStore sets value in Logseq app state store by key.
func (c *Client) SetStateFromStore(ctx context.Context, key string, value any) error {
	_, err := c.CallAPI(ctx, "logseq.App.setStateFromStore", key, value)
	return err
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
