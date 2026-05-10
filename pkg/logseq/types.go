package logseq

import "encoding/json"

// Page represents a Logseq page entity.
type Page struct {
	Name         string         `json:"name,omitempty"`
	OriginalName string         `json:"originalName,omitempty"`
	UUID         string         `json:"uuid,omitempty"`
	Properties   map[string]any `json:"properties,omitempty"`
	IsJournal    bool           `json:"journal?,omitempty"`
	JournalDay   int            `json:"journalDay,omitempty"`
	CreatedAt    int64          `json:"createdAt,omitempty"`
	UpdatedAt    int64          `json:"updatedAt,omitempty"`
}

// DisplayName returns the best available name for the page.
func (p *Page) DisplayName() string {
	if p.OriginalName != "" {
		return p.OriginalName
	}
	return p.Name
}

// Block represents a Logseq block entity.
type Block struct {
	UUID       string         `json:"uuid,omitempty"`
	Content    string         `json:"content,omitempty"`
	Page       *PageRef       `json:"page,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
	Children   []Block        `json:"children,omitempty"`
	Level      int            `json:"level,omitempty"`
	Format     string         `json:"format,omitempty"`
	Marker     string         `json:"marker,omitempty"`
	Priority   string         `json:"priority,omitempty"`
}

// PageRef is a reference to a page, which can be either an integer ID or {id: int}.
type PageRef struct {
	ID int64 `json:"id"`
}

// UnmarshalJSON handles both integer and object forms of page reference.
func (r *PageRef) UnmarshalJSON(data []byte) error {
	// Try integer first
	var id int64
	if err := json.Unmarshal(data, &id); err == nil {
		r.ID = id
		return nil
	}
	// Try object form {id: int}
	var obj struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	r.ID = obj.ID
	return nil
}

// GraphInfo represents current graph metadata.
type GraphInfo struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`
}

// BatchBlock is the structure for insertBatchBlock API.
type BatchBlock struct {
	Content    string         `json:"content"`
	Children   []BatchBlock   `json:"children,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// InsertBlockOptions configures insertBlock behavior.
type InsertBlockOptions struct {
	Sibling    bool           `json:"sibling,omitempty"`
	Before     bool           `json:"before,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// CreatePageOptions configures createPage behavior.
type CreatePageOptions struct {
	CreateFirstBlock bool `json:"createFirstBlock,omitempty"`
	Journal          bool `json:"journal,omitempty"`
}

// SearchResult represents a search result from logseq.App.search.
type SearchResult struct {
	Blocks []SearchBlock `json:"blocks,omitempty"`
	Pages  []string      `json:"pages,omitempty"`
	Files  []string      `json:"files,omitempty"`
}

// SearchBlock is a block entry from search results.
// Uses custom UnmarshalJSON to handle both standard and namespaced (block/xxx) field names.
type SearchBlock struct {
	UUID    string `json:"uuid,omitempty"`
	Content string `json:"content,omitempty"`
	Page    string `json:"page,omitempty"`
}

// UnmarshalJSON handles Logseq search results which may use namespaced keys
// like "block/uuid", "block/content", and have page as int, string, or object.
func (b *SearchBlock) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	tryString := func(keys ...string) string {
		for _, k := range keys {
			v, ok := raw[k]
			if !ok {
				continue
			}
			var s string
			if json.Unmarshal(v, &s) == nil {
				return s
			}
		}
		return ""
	}

	b.UUID = tryString("uuid", "block/uuid")
	b.Content = tryString("content", "block/content")
	b.Page = tryString("page", "block/page")

	// Page may be an integer ID — convert to string for display
	if b.Page == "" {
		for _, k := range []string{"page", "block/page"} {
			v, ok := raw[k]
			if !ok {
				continue
			}
			var id int64
			if json.Unmarshal(v, &id) == nil {
				b.Page = "#" + json.Number(v).String()
				break
			}
			// Try object form {id: N, name: "...", ...}
			var obj map[string]json.RawMessage
			if json.Unmarshal(v, &obj) == nil {
				for _, nk := range []string{"name", "original-name", "originalName", "block/name"} {
					if nv, ok := obj[nk]; ok {
						var s string
						if json.Unmarshal(nv, &s) == nil && s != "" {
							b.Page = s
							break
						}
					}
				}
			}
		}
	}

	return nil
}
