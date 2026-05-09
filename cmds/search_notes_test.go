package cmds

import (
	"testing"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

func TestParseCursorOffset(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"  ", 0, false},
		{"0", 0, false},
		{"10", 10, false},
		{"200", 200, false},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run("cursor="+tt.input, func(t *testing.T) {
			got, err := parseCursorOffset(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseCursorOffset(%q)=%d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseIncludeSet(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]bool
	}{
		{"", map[string]bool{}},
		{"pages", map[string]bool{"pages": true}},
		{"pages,blocks", map[string]bool{"pages": true, "blocks": true}},
		{"PAGES,BLOCKS,FILES", map[string]bool{"pages": true, "blocks": true, "files": true}},
		{" pages , blocks ", map[string]bool{"pages": true, "blocks": true}},
		{"pages,invalid,blocks", map[string]bool{"pages": true, "blocks": true}},
	}

	for _, tt := range tests {
		t.Run("include="+tt.input, func(t *testing.T) {
			got := parseIncludeSet(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseIncludeSet(%q)=%v, want %v", tt.input, got, tt.want)
				return
			}
			for k := range tt.want {
				if !got[k] {
					t.Errorf("missing key %q", k)
				}
			}
		})
	}
}

func TestPaginateSearchItems(t *testing.T) {
	items := make([]searchNotesItem, 10)
	for i := range items {
		items[i] = searchNotesItem{Type: "page", Title: "p" + string(rune('0'+i))}
	}

	// First page
	paged, cursor, hasMore := paginateSearchItems(items, 0, 3)
	if len(paged) != 3 {
		t.Errorf("len=%d, want 3", len(paged))
	}
	if cursor != "3" {
		t.Errorf("cursor=%q, want %q", cursor, "3")
	}
	if !hasMore {
		t.Error("expected hasMore=true")
	}

	// Middle page
	paged, cursor, hasMore = paginateSearchItems(items, 3, 3)
	if len(paged) != 3 {
		t.Errorf("len=%d, want 3", len(paged))
	}
	if cursor != "6" {
		t.Errorf("cursor=%q, want %q", cursor, "6")
	}
	if !hasMore {
		t.Error("expected hasMore=true")
	}

	// Last page (partial)
	paged, cursor, hasMore = paginateSearchItems(items, 8, 3)
	if len(paged) != 2 {
		t.Errorf("len=%d, want 2", len(paged))
	}
	if cursor != "" {
		t.Errorf("cursor=%q, want empty", cursor)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}

	// Offset beyond items
	paged, _, hasMore = paginateSearchItems(items, 100, 3)
	if len(paged) != 0 {
		t.Errorf("len=%d, want 0", len(paged))
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}

	// Exact fit
	paged, cursor, hasMore = paginateSearchItems(items, 0, 10)
	if len(paged) != 10 {
		t.Errorf("len=%d, want 10", len(paged))
	}
	if cursor != "" {
		t.Errorf("cursor=%q, want empty", cursor)
	}
	if hasMore {
		t.Error("expected hasMore=false")
	}
}

func TestFilterSearchItemsByTag(t *testing.T) {
	items := []searchNotesItem{
		{Type: "page", Title: "My #golang notes"},
		{Type: "block", Title: "Some block", Snippet: "about [[python]]"},
		{Type: "page", Title: "Unrelated page"},
		{Type: "block", Title: "golang tips", Page: "golang"},
	}

	// Filter by "golang"
	filtered := filterSearchItemsByTag(items, "golang")
	if len(filtered) != 2 {
		t.Errorf("len=%d, want 2", len(filtered))
	}

	// Filter with # prefix
	filtered = filterSearchItemsByTag(items, "#python")
	if len(filtered) != 1 {
		t.Errorf("len=%d, want 1", len(filtered))
	}

	// Empty tag returns all
	filtered = filterSearchItemsByTag(items, "")
	if len(filtered) != len(items) {
		t.Errorf("len=%d, want %d", len(filtered), len(items))
	}
}

func TestNormalizeSearchItems(t *testing.T) {
	res := &logseq.SearchResult{
		Pages:  []string{"Page A", "Page B"},
		Blocks: []logseq.SearchBlock{{UUID: "u1", Content: "block content", Page: "pg"}},
		Files:  []string{"file1.md"},
	}

	// pages+blocks only
	items := normalizeSearchItems(res, map[string]bool{"pages": true, "blocks": true})
	pageCount, blockCount, fileCount := 0, 0, 0
	for _, it := range items {
		switch it.Type {
		case "page":
			pageCount++
		case "block":
			blockCount++
		case "file":
			fileCount++
		}
	}
	if pageCount != 2 {
		t.Errorf("pageCount=%d, want 2", pageCount)
	}
	if blockCount != 1 {
		t.Errorf("blockCount=%d, want 1", blockCount)
	}
	if fileCount != 0 {
		t.Error("files should be excluded")
	}

	// include files
	items = normalizeSearchItems(res, map[string]bool{"files": true})
	for _, it := range items {
		if it.Type != "file" {
			t.Errorf("unexpected type %q", it.Type)
		}
	}
	if len(items) != 1 {
		t.Errorf("len=%d, want 1", len(items))
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 0, "abc"},
		{"  spaces  ", 3, "spa..."},
	}

	for _, tt := range tests {
		got := truncate(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("truncate(%q, %d)=%q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}
