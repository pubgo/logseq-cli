package logseq

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CreateTag creates a tag page.
func (c *Client) CreateTag(ctx context.Context, tagName string, customUUID ...string) (*Page, error) {
	args := []any{tagName}
	if len(customUUID) > 0 && strings.TrimSpace(customUUID[0]) != "" {
		args = append(args, map[string]any{"uuid": strings.TrimSpace(customUUID[0])})
	}

	raw, err := c.CallAPI(ctx, "logseq.Editor.createTag", args...)
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

// GetTag gets a tag by name or numeric entity id (as string).
func (c *Client) GetTag(ctx context.Context, nameOrIdent string) (*Page, error) {
	ident := strings.TrimSpace(nameOrIdent)
	arg := any(ident)
	if id, err := strconv.ParseInt(ident, 10, 64); err == nil {
		arg = id
	}

	raw, err := c.CallAPI(ctx, "logseq.Editor.getTag", arg)
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

// GetTagsByName gets tags by fuzzy/equal name in Logseq.
// Falls back to GetAllTags + local filtering if the native API is unavailable.
func (c *Client) GetTagsByName(ctx context.Context, tagName string) ([]Page, error) {
	result, err := decode[[]Page](c.CallAPI(ctx, "logseq.Editor.getTagsByName", tagName))
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "methodnotexist") {
		return c.getTagsByNameFallback(ctx, tagName)
	}
	return result, err
}

func (c *Client) getTagsByNameFallback(ctx context.Context, tagName string) ([]Page, error) {
	allTags, err := c.GetAllTags(ctx)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(tagName)
	var pages []Page
	for _, t := range allTags {
		if strings.Contains(strings.ToLower(t), needle) {
			pages = append(pages, Page{Name: t, OriginalName: t})
		}
	}
	return pages, nil
}

// GetTagObjects gets tag object blocks for the given tag name.
func (c *Client) GetTagObjects(ctx context.Context, nameOrIdent string) ([]Block, error) {
	result, err := decode[[]Block](c.CallAPI(ctx, "logseq.Editor.getTagObjects", nameOrIdent))
	if err != nil {
		return nil, wrapCapabilityUnavailable("logseq.Editor.getTagObjects", err)
	}
	return result, nil
}

// AddTagProperty adds property relation to a tag.
func (c *Client) AddTagProperty(ctx context.Context, tagID, propertyIDOrName string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.addTagProperty", tagID, propertyIDOrName)
	return wrapCapabilityUnavailable("logseq.Editor.addTagProperty", err)
}

// RemoveTagProperty removes property relation from a tag.
func (c *Client) RemoveTagProperty(ctx context.Context, tagID, propertyIDOrName string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.removeTagProperty", tagID, propertyIDOrName)
	return wrapCapabilityUnavailable("logseq.Editor.removeTagProperty", err)
}

// AddTagExtends adds parent tag relation for a tag.
func (c *Client) AddTagExtends(ctx context.Context, tagID, parentTagIDOrName string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.addTagExtends", tagID, parentTagIDOrName)
	return wrapCapabilityUnavailable("logseq.Editor.addTagExtends", err)
}

// RemoveTagExtends removes parent tag relation for a tag.
func (c *Client) RemoveTagExtends(ctx context.Context, tagID, parentTagIDOrName string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.removeTagExtends", tagID, parentTagIDOrName)
	return wrapCapabilityUnavailable("logseq.Editor.removeTagExtends", err)
}

// AddBlockTag adds a tag to a block.
func (c *Client) AddBlockTag(ctx context.Context, blockID, tagID string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.addBlockTag", blockID, tagID)
	return wrapCapabilityUnavailable("logseq.Editor.addBlockTag", err)
}

// RemoveBlockTag removes a tag from a block.
func (c *Client) RemoveBlockTag(ctx context.Context, blockID, tagID string) error {
	_, err := c.CallAPI(ctx, "logseq.Editor.removeBlockTag", blockID, tagID)
	return wrapCapabilityUnavailable("logseq.Editor.removeBlockTag", err)
}

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

	// Fallback for Logseq variants where :block/tags is empty but refs are populated.
	// This may include both hashtag-style tags and page-link style tags.
	rawRefs, refsErr := c.DatascriptQuery(ctx, "[:find ?name :where [?b :block/refs ?r] [?r :block/name ?name]]")
	if refsErr == nil {
		for _, name := range parseSingleColumnStrings(rawRefs) {
			if !isLikelyTagCandidate(name) {
				continue
			}
			tags[normalizeTag(name)] = struct{}{}
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

func isLikelyTagCandidate(s string) bool {
	v := normalizeTag(s)
	if v == "" {
		return false
	}

	if isJournalName(v) {
		return false
	}

	r := []rune(v)
	if len(r) == 1 && strings.ContainsRune(".,;:!?，。；：！？()[]{}<>/\\'\"`~|+-=", r[0]) {
		return false
	}

	return true
}

func isJournalName(v string) bool {
	if len(v) == 10 {
		if (v[4] == '-' && v[7] == '-') || (v[4] == '_' && v[7] == '_') {
			for i := 0; i < len(v); i++ {
				if i == 4 || i == 7 {
					continue
				}
				if v[i] < '0' || v[i] > '9' {
					return false
				}
			}
			return true
		}
	}

	if len(v) == 8 {
		for i := 0; i < len(v); i++ {
			if v[i] < '0' || v[i] > '9' {
				return false
			}
		}
		return true
	}

	return false
}
