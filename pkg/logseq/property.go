package logseq

import "context"

// GetAllProperties returns all property entities.
func (c *Client) GetAllProperties(ctx context.Context) ([]Page, error) {
	return decode[[]Page](c.CallAPI(ctx, "logseq.Editor.getAllProperties"))
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
