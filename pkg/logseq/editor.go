package logseq

import (
	"context"
	"encoding/json"
)

// === Page Operations ===

// GetAllPages returns all pages in the graph.
func (c *Client) GetAllPages(ctx context.Context) ([]Page, error) {
	return decode[[]Page](c.CallAPI(ctx, "logseq.Editor.getAllPages"))
}

// GetPage returns a page by name or UUID.
func (c *Client) GetPage(ctx context.Context, nameOrUUID string) (*Page, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.getPage", nameOrUUID)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var page Page
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// CreatePage creates a new page.
func (c *Client) CreatePage(ctx context.Context, name string, properties map[string]any, opts *CreatePageOptions) (*Page, error) {
	args := []any{name}
	if properties != nil {
		args = append(args, properties)
	} else {
		args = append(args, map[string]any{})
	}
	if opts != nil {
		args = append(args, opts)
	}
	raw, err := c.CallAPI(ctx, "logseq.Editor.createPage", args...)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var page Page
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// CreateJournalPage creates a journal page for a specific date string.
func (c *Client) CreateJournalPage(ctx context.Context, date string) (*Page, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.createJournalPage", date)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var page Page
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// DeletePage deletes a page by name.
func (c *Client) DeletePage(ctx context.Context, name string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.deletePage", name)
	return err
}

// RenamePage renames a page.
func (c *Client) RenamePage(ctx context.Context, oldName, newName string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.renamePage", oldName, newName)
	return err
}

// GetPageBlocksTree returns the full block tree for a page.
func (c *Client) GetPageBlocksTree(ctx context.Context, nameOrUUID string) ([]Block, error) {
	return decode[[]Block](c.CallAPI(ctx, "logseq.Editor.getPageBlocksTree", nameOrUUID))
}

// GetPageLinkedReferences returns backlinks for a page.
func (c *Client) GetPageLinkedReferences(ctx context.Context, name string) (json.RawMessage, error) {
	return c.CallAPI(ctx, "logseq.Editor.getPageLinkedReferences", name)
}

// GetPagesFromNamespace returns all pages in a namespace (flat).
func (c *Client) GetPagesFromNamespace(ctx context.Context, namespace string) ([]Page, error) {
	return decode[[]Page](c.CallAPI(ctx, "logseq.Editor.getPagesFromNamespace", namespace))
}

// GetPagesTreeFromNamespace returns pages in a namespace as a tree.
func (c *Client) GetPagesTreeFromNamespace(ctx context.Context, namespace string) (json.RawMessage, error) {
	return c.CallAPI(ctx, "logseq.Editor.getPagesTreeFromNamespace", namespace)
}

// SetPageProperties sets page-level properties.
func (c *Client) SetPageProperties(ctx context.Context, name string, properties map[string]any) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.setPageProperties", name, properties)
	return err
}

// === Block Operations ===

// GetBlock returns a block by UUID with optional children.
func (c *Client) GetBlock(ctx context.Context, uuid string, includeChildren bool) (*Block, error) {
	opts := map[string]any{"includeChildren": includeChildren}
	raw, err := c.CallAPI(ctx, "logseq.Editor.getBlock", uuid, opts)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// InsertBlock inserts a block relative to a target block.
func (c *Client) InsertBlock(ctx context.Context, targetUUID, content string, opts *InsertBlockOptions) (*Block, error) {
	args := []any{targetUUID, content}
	if opts != nil {
		args = append(args, opts)
	}
	raw, err := c.CallAPI(ctx, "logseq.Editor.insertBlock", args...)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// InsertBatchBlock inserts multiple blocks at once.
func (c *Client) InsertBatchBlock(ctx context.Context, srcUUID string, blocks []BatchBlock, sibling bool) error {
	opts := map[string]any{"sibling": sibling}
	_, err := c.CallAPI(ctx, "logseq.Editor.insertBatchBlock", srcUUID, blocks, opts)
	return err
}

// UpdateBlock updates a block's content.
func (c *Client) UpdateBlock(ctx context.Context, uuid, content string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.updateBlock", uuid, content)
	return err
}

// RemoveBlock deletes a block by UUID.
func (c *Client) RemoveBlock(ctx context.Context, uuid string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.removeBlock", uuid)
	return err
}

// MoveBlock moves a block to a new position.
func (c *Client) MoveBlock(ctx context.Context, srcUUID, targetUUID string, opts map[string]any) error {
	args := []any{srcUUID, targetUUID}
	if opts != nil {
		args = append(args, opts)
	}
	_, err := c.CallAPI(ctx, "logseq.Editor.moveBlock", args...)
	return err
}

// GetPreviousSiblingBlock returns previous sibling block of given block.
func (c *Client) GetPreviousSiblingBlock(ctx context.Context, uuid string) (*Block, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.getPreviousSiblingBlock", uuid)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// GetNextSiblingBlock returns next sibling block of given block.
func (c *Client) GetNextSiblingBlock(ctx context.Context, uuid string) (*Block, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.getNextSiblingBlock", uuid)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// PrependBlockInPage inserts a block at the beginning of a page.
func (c *Client) PrependBlockInPage(ctx context.Context, page, content string) (*Block, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.prependBlockInPage", page, content)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// AppendBlockInPage appends a block to the end of a page.
// Note: logseq.Editor.appendBlock does not exist. This uses appendBlockInPage.
func (c *Client) AppendBlockInPage(ctx context.Context, page, content string) (*Block, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.appendBlockInPage", page, content)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// UpsertBlockProperty sets a block property.
func (c *Client) UpsertBlockProperty(ctx context.Context, uuid, key string, value any) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.upsertBlockProperty", uuid, key, value)
	return err
}

// RemoveBlockProperty removes a block property.
func (c *Client) RemoveBlockProperty(ctx context.Context, uuid, key string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.removeBlockProperty", uuid, key)
	return err
}

// GetBlockProperties returns all properties of a block.
func (c *Client) GetBlockProperties(ctx context.Context, uuid string) (map[string]any, error) {
	return decode[map[string]any](c.CallAPI(ctx, "logseq.Editor.getBlockProperties", uuid))
}

// SetBlockCollapsed sets whether a block is collapsed.
func (c *Client) SetBlockCollapsed(ctx context.Context, uuid string, collapsed bool) error {
	opts := map[string]any{"flag": collapsed}
	_, err := c.CallAPI(ctx, "logseq.Editor.setBlockCollapsed", uuid, opts)
	return err
}

// GetCurrentPage returns the currently focused page.
func (c *Client) GetCurrentPage(ctx context.Context) (*Page, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.getCurrentPage")
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var page Page
	if err := json.Unmarshal(raw, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// GetSelectedBlocks returns currently selected blocks.
func (c *Client) GetSelectedBlocks(ctx context.Context) ([]Block, error) {
	return decode[[]Block](c.CallAPI(ctx, "logseq.Editor.getSelectedBlocks"))
}

// ClearSelectedBlocks clears current selected blocks.
func (c *Client) ClearSelectedBlocks(ctx context.Context) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.clearSelectedBlocks")
	return err
}

// NewBlockUUID creates a unique UUID string.
func (c *Client) NewBlockUUID(ctx context.Context) (string, error) {
	return decode[string](c.CallAPI(ctx, "logseq.Editor.newBlockUUID"))
}

// GetCurrentPageBlocksTree returns block tree of currently focused page.
func (c *Client) GetCurrentPageBlocksTree(ctx context.Context) ([]Block, error) {
	return decode[[]Block](c.CallAPI(ctx, "logseq.Editor.getCurrentPageBlocksTree"))
}

// GetPageProperties returns properties of a page directly.
func (c *Client) GetPageProperties(ctx context.Context, page string) (map[string]any, error) {
	return decode[map[string]any](c.CallAPI(ctx, "logseq.Editor.getPageProperties", page))
}

// GetCurrentBlock returns the currently focused block.
func (c *Client) GetCurrentBlock(ctx context.Context) (*Block, error) {
	raw, err := c.CallAPI(ctx, "logseq.Editor.getCurrentBlock")
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var block Block
	if err := json.Unmarshal(raw, &block); err != nil {
		return nil, err
	}
	return &block, nil
}
