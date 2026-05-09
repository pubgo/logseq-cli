package cmds

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
	"github.com/pubgo/redant"
)

type searchNotesItem struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Snippet string `json:"snippet,omitempty"`
	Page    string `json:"page,omitempty"`
	UUID    string `json:"uuid,omitempty"`
}

type searchNotesResult struct {
	Query      string            `json:"query"`
	Tag        string            `json:"tag,omitempty"`
	Limit      int               `json:"limit"`
	Cursor     string            `json:"cursor,omitempty"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
	Total      int               `json:"total"`
	Items      []searchNotesItem `json:"items"`
}

func SearchNotesCmd() *redant.Command {
	var (
		tag      string
		limitRaw string
		cursor   string
		include  string
	)

	return &redant.Command{
		Use:   "search-notes <query>",
		Short: "LLM-friendly search with pagination/cursor",
		Args: redant.ArgSet{
			{Name: "query", Required: true, Value: redant.StringOf(new(string)), Description: "Search query"},
		},
		Options: redant.OptionSet{
			{
				Flag:        "tag",
				Description: "Optional tag filter",
				Value:       redant.StringOf(&tag),
			},
			{
				Flag:        "limit",
				Description: "Page size (1-200, default 20)",
				Default:     "20",
				Value:       redant.StringOf(&limitRaw),
			},
			{
				Flag:        "cursor",
				Description: "Opaque cursor (current implementation uses offset integer)",
				Value:       redant.StringOf(&cursor),
			},
			{
				Flag:        "include",
				Description: "Comma-separated types: pages,blocks,files (default pages,blocks)",
				Default:     "pages,blocks",
				Value:       redant.StringOf(&include),
			},
		},
		ResponseHandler: redant.Unary(func(ctx context.Context, inv *redant.Invocation) (*llmEnvelope, error) {
			start := time.Now()
			q := strings.TrimSpace(inv.Args[0])
			if q == "" {
				return envelopeFailure(start, fmt.Errorf("query is required"), "BAD_REQUEST", "请提供非空 query"), nil
			}

			limit := 20
			if strings.TrimSpace(limitRaw) != "" {
				v, convErr := strconv.Atoi(strings.TrimSpace(limitRaw))
				if convErr != nil {
					return envelopeFailure(start, fmt.Errorf("invalid limit: %w", convErr), "BAD_REQUEST", "limit 需要是数字"), nil
				}
				limit = v
			}

			if limit <= 0 {
				limit = 20
			}
			if limit > 200 {
				limit = 200
			}

			offset, err := parseCursorOffset(cursor)
			if err != nil {
				return envelopeFailure(start, fmt.Errorf("invalid cursor: %w", err), "BAD_REQUEST", "cursor 需要是非负整数"), nil
			}

			includeSet := parseIncludeSet(include)
			if len(includeSet) == 0 {
				includeSet = map[string]bool{"pages": true, "blocks": true}
			}

			client := NewClient()
			res, err := client.Search(ctx, q)
			if err != nil {
				return envelopeFailure(start, err, "UPSTREAM_ERROR", "可先执行 capabilities get 确认 search 可用性", withCapabilityUsed("logseq.App.search")), nil
			}
			if res == nil {
				res = &logseq.SearchResult{}
			}

			items := normalizeSearchItems(res, includeSet)
			if strings.TrimSpace(tag) != "" {
				items = filterSearchItemsByTag(items, strings.TrimSpace(tag))
			}

			total := len(items)
			paged, nextCursor, hasMore := paginateSearchItems(items, offset, limit)

			out := &searchNotesResult{
				Query:      q,
				Tag:        strings.TrimSpace(tag),
				Limit:      limit,
				Cursor:     strings.TrimSpace(cursor),
				NextCursor: nextCursor,
				HasMore:    hasMore,
				Total:      total,
				Items:      paged,
			}

			hints := make([]string, 0, 2)

			if len(res.Files) > 0 && !includeSet["files"] {
				hints = append(hints, "files 结果已被 include 过滤")
			}
			if strings.TrimSpace(tag) != "" {
				hints = append(hints, "tag 过滤基于结果文本匹配（page/title/snippet）")
			}

			return envelopeSuccess(
				start,
				out,
				withCapabilityUsed("logseq.App.search"),
				withPageMeta(nextCursor, hasMore),
				withHints(hints...),
			), nil
		}),
	}
}

func parseCursorOffset(cursor string) (int, error) {
	c := strings.TrimSpace(cursor)
	if c == "" {
		return 0, nil
	}
	o, err := strconv.Atoi(c)
	if err != nil {
		return 0, err
	}
	if o < 0 {
		return 0, fmt.Errorf("offset must be >= 0")
	}
	return o, nil
}

func parseIncludeSet(include string) map[string]bool {
	set := map[string]bool{}
	for _, p := range strings.Split(strings.ToLower(include), ",") {
		k := strings.TrimSpace(p)
		if k == "pages" || k == "blocks" || k == "files" {
			set[k] = true
		}
	}
	return set
}

func normalizeSearchItems(res *logseq.SearchResult, include map[string]bool) []searchNotesItem {
	items := make([]searchNotesItem, 0, len(res.Pages)+len(res.Blocks)+len(res.Files))

	if include["pages"] {
		pages := append([]string(nil), res.Pages...)
		sort.Strings(pages)
		for _, p := range pages {
			t := strings.TrimSpace(p)
			if t == "" {
				continue
			}
			items = append(items, searchNotesItem{Type: "page", Title: t})
		}
	}

	if include["blocks"] {
		for _, b := range res.Blocks {
			content := strings.TrimSpace(b.Content)
			page := strings.TrimSpace(b.Page)
			title := content
			if title == "" {
				title = "(empty block)"
			}
			items = append(items, searchNotesItem{
				Type:    "block",
				Title:   truncate(title, 120),
				Snippet: truncate(content, 500),
				Page:    page,
				UUID:    strings.TrimSpace(b.UUID),
			})
		}
	}

	if include["files"] {
		files := append([]string(nil), res.Files...)
		sort.Strings(files)
		for _, f := range files {
			t := strings.TrimSpace(f)
			if t == "" {
				continue
			}
			items = append(items, searchNotesItem{Type: "file", Title: t})
		}
	}

	return items
}

func filterSearchItemsByTag(items []searchNotesItem, tag string) []searchNotesItem {
	needle := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(tag, "#")))
	if needle == "" {
		return items
	}

	out := make([]searchNotesItem, 0, len(items))
	for _, it := range items {
		hay := strings.ToLower(it.Title + "\n" + it.Snippet + "\n" + it.Page)
		if strings.Contains(hay, "#"+needle) || strings.Contains(hay, "[["+needle+"]]") || strings.Contains(hay, needle) {
			out = append(out, it)
		}
	}
	return out
}

func paginateSearchItems(items []searchNotesItem, offset, limit int) ([]searchNotesItem, string, bool) {
	if offset >= len(items) {
		return []searchNotesItem{}, "", false
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	paged := items[offset:end]
	hasMore := end < len(items)
	if !hasMore {
		return paged, "", false
	}
	return paged, strconv.Itoa(end), true
}

func truncate(s string, n int) string {
	v := strings.TrimSpace(s)
	if n <= 0 || len(v) <= n {
		return v
	}
	return v[:n] + "..."
}
