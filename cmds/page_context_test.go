package cmds

import (
	"testing"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

func TestPageContextBuilderRespectsMaxDepth(t *testing.T) {
	blocks := []logseq.Block{
		{
			UUID: "A",
			Children: []logseq.Block{
				{
					UUID: "B",
					Children: []logseq.Block{
						{UUID: "C"},
					},
				},
			},
		},
		{UUID: "D"},
	}

	builder := &pageContextBuilder{maxBlocks: 20, maxDepth: 1}
	out := builder.buildBlocks(blocks, 1)

	if got, want := len(out), 2; got != want {
		t.Fatalf("len(out)=%d, want %d", got, want)
	}
	if got, want := builder.seen, 2; got != want {
		t.Fatalf("seen=%d, want %d", got, want)
	}
	if got, want := builder.clippedByDepth, 2; got != want {
		t.Fatalf("clippedByDepth=%d, want %d", got, want)
	}
	if got := builder.clippedByLimit; got != 0 {
		t.Fatalf("clippedByLimit=%d, want 0", got)
	}

	if got := len(out[0].Children); got != 0 {
		t.Fatalf("root children should be clipped by depth, got %d", got)
	}
}

func TestPageContextBuilderRespectsMaxBlocks(t *testing.T) {
	blocks := []logseq.Block{
		{
			UUID: "A",
			Children: []logseq.Block{
				{UUID: "B"},
			},
		},
		{UUID: "C"},
	}

	builder := &pageContextBuilder{maxBlocks: 2, maxDepth: 10}
	out := builder.buildBlocks(blocks, 1)

	if got, want := builder.seen, 2; got != want {
		t.Fatalf("seen=%d, want %d", got, want)
	}
	if got, want := builder.clippedByLimit, 1; got != want {
		t.Fatalf("clippedByLimit=%d, want %d", got, want)
	}
	if got := builder.clippedByDepth; got != 0 {
		t.Fatalf("clippedByDepth=%d, want 0", got)
	}

	if got, want := len(out), 1; got != want {
		t.Fatalf("len(out)=%d, want %d", got, want)
	}
	if out[0].UUID != "A" {
		t.Fatalf("first uuid=%q, want A", out[0].UUID)
	}
	if got, want := len(out[0].Children), 1; got != want {
		t.Fatalf("len(out[0].Children)=%d, want %d", got, want)
	}
	if out[0].Children[0].UUID != "B" {
		t.Fatalf("child uuid=%q, want B", out[0].Children[0].UUID)
	}
}

func TestCountBlockSubtreeValue(t *testing.T) {
	root := logseq.Block{
		UUID: "A",
		Children: []logseq.Block{
			{UUID: "B"},
			{
				UUID:     "C",
				Children: []logseq.Block{{UUID: "D"}},
			},
		},
	}

	if got, want := countBlockSubtreeValue(root), 4; got != want {
		t.Fatalf("countBlockSubtreeValue=%d, want %d", got, want)
	}
}
