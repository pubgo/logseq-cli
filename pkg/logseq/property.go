package logseq

import (
	"context"
	"sort"
	"strings"
)

// GetAllProperties returns all property entities.
// Falls back to extracting property keys from all pages if the native API is unavailable.
func (c *Client) GetAllProperties(ctx context.Context) ([]Page, error) {
	result, err := decode[[]Page](c.CallAPI(ctx, "logseq.Editor.getAllProperties"))
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "methodnotexist") {
		return c.getAllPropertiesFallback(ctx)
	}
	return result, err
}

func (c *Client) getAllPropertiesFallback(ctx context.Context) ([]Page, error) {
	allPages, err := c.GetAllPages(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, p := range allPages {
		for k := range p.Properties {
			seen[k] = struct{}{}
		}
	}
	pages := make([]Page, 0, len(seen))
	for k := range seen {
		pages = append(pages, Page{Name: k, OriginalName: k})
	}
	sort.Slice(pages, func(i, j int) bool { return pages[i].Name < pages[j].Name })
	return pages, nil
}

// GetProperty returns a property entity by key.
func (c *Client) GetProperty(ctx context.Context, key string) (any, error) {
	return decode[any](c.CallAPI(ctx, "logseq.Editor.getProperty", key))
}

// UpsertProperty creates or updates property schema.
// schema and opts are raw JSON-like objects accepted by Logseq.
func (c *Client) UpsertProperty(ctx context.Context, key string, schema map[string]any, opts map[string]any) (any, error) {
	args := []any{key}
	if schema != nil {
		args = append(args, schema)
	}
	if opts != nil {
		args = append(args, opts)
	}
	return decode[any](c.CallAPI(ctx, "logseq.Editor.upsertProperty", args...))
}

// RemoveProperty removes a property schema by key.
func (c *Client) RemoveProperty(ctx context.Context, key string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.removeProperty", key)
	return err
}
