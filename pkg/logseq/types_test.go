package logseq

import (
	"encoding/json"
	"testing"
)

func TestSearchBlock_UnmarshalJSON_StandardKeys(t *testing.T) {
	input := `{"uuid":"abc-123","content":"hello world","page":"my-page"}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.UUID != "abc-123" {
		t.Errorf("UUID = %q, want %q", b.UUID, "abc-123")
	}
	if b.Content != "hello world" {
		t.Errorf("Content = %q, want %q", b.Content, "hello world")
	}
	if b.Page != "my-page" {
		t.Errorf("Page = %q, want %q", b.Page, "my-page")
	}
}

func TestSearchBlock_UnmarshalJSON_NamespacedKeys(t *testing.T) {
	input := `{"block/uuid":"def-456","block/content":"namespaced content","block/page":"ns-page"}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.UUID != "def-456" {
		t.Errorf("UUID = %q, want %q", b.UUID, "def-456")
	}
	if b.Content != "namespaced content" {
		t.Errorf("Content = %q, want %q", b.Content, "namespaced content")
	}
	if b.Page != "ns-page" {
		t.Errorf("Page = %q, want %q", b.Page, "ns-page")
	}
}

func TestSearchBlock_UnmarshalJSON_PageAsInt(t *testing.T) {
	input := `{"uuid":"aaa","content":"test","page":56312}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.Page != "#56312" {
		t.Errorf("Page = %q, want %q", b.Page, "#56312")
	}
}

func TestSearchBlock_UnmarshalJSON_PageAsObject(t *testing.T) {
	input := `{"uuid":"bbb","content":"test","page":{"id":100,"name":"my-page","original-name":"My Page"}}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.Page != "my-page" {
		t.Errorf("Page = %q, want %q", b.Page, "my-page")
	}
}

func TestSearchBlock_UnmarshalJSON_PageAsObjectOriginalName(t *testing.T) {
	input := `{"uuid":"ccc","content":"test","page":{"id":200,"originalName":"Original Page"}}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.Page != "Original Page" {
		t.Errorf("Page = %q, want %q", b.Page, "Original Page")
	}
}

func TestSearchBlock_UnmarshalJSON_EmptyObject(t *testing.T) {
	input := `{}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.UUID != "" || b.Content != "" || b.Page != "" {
		t.Errorf("expected empty SearchBlock, got %+v", b)
	}
}

func TestSearchBlock_UnmarshalJSON_MixedKeys(t *testing.T) {
	// Standard UUID but namespaced content
	input := `{"uuid":"mix-789","block/content":"mixed content","page":42}`
	var b SearchBlock
	if err := json.Unmarshal([]byte(input), &b); err != nil {
		t.Fatal(err)
	}
	if b.UUID != "mix-789" {
		t.Errorf("UUID = %q, want %q", b.UUID, "mix-789")
	}
	if b.Content != "mixed content" {
		t.Errorf("Content = %q, want %q", b.Content, "mixed content")
	}
	if b.Page != "#42" {
		t.Errorf("Page = %q, want %q", b.Page, "#42")
	}
}

func TestSearchResult_UnmarshalJSON(t *testing.T) {
	input := `{
		"blocks": [
			{"uuid":"a1","content":"block one","page":"p1"},
			{"block/uuid":"a2","block/content":"block two","block/page":123}
		],
		"pages": ["page-one","page-two"],
		"files": ["file.md"]
	}`
	var sr SearchResult
	if err := json.Unmarshal([]byte(input), &sr); err != nil {
		t.Fatal(err)
	}
	if len(sr.Blocks) != 2 {
		t.Fatalf("Blocks len = %d, want 2", len(sr.Blocks))
	}
	if sr.Blocks[0].UUID != "a1" {
		t.Errorf("Blocks[0].UUID = %q, want %q", sr.Blocks[0].UUID, "a1")
	}
	if sr.Blocks[1].UUID != "a2" {
		t.Errorf("Blocks[1].UUID = %q, want %q", sr.Blocks[1].UUID, "a2")
	}
	if sr.Blocks[1].Page != "#123" {
		t.Errorf("Blocks[1].Page = %q, want %q", sr.Blocks[1].Page, "#123")
	}
	if len(sr.Pages) != 2 {
		t.Errorf("Pages len = %d, want 2", len(sr.Pages))
	}
}

func TestPageRef_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{"integer", "42", 42},
		{"object", `{"id":99}`, 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r PageRef
			if err := json.Unmarshal([]byte(tt.input), &r); err != nil {
				t.Fatal(err)
			}
			if r.ID != tt.want {
				t.Errorf("ID = %d, want %d", r.ID, tt.want)
			}
		})
	}
}
