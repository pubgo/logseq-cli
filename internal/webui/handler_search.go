package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	if q == "" {
		writeError(w, http.StatusBadRequest, "missing query: q")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	keywords := splitSearchKeywords(q)
	result, err := s.searchByKeywords(ctx, keywords)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	if strings.TrimSpace(tag) != "" {
		pageSet, err := s.getSearchPageSetByTag(ctx, tag)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		result = filterSearchResultByPageSet(result, pageSet)
	}

	// Filter out empty block objects from search results
	if result != nil {
		result.Blocks = filterEmptyBlocks(result.Blocks)
		s.resolveSearchBlockPageIDs(ctx, result)
	}

	writeOK(w, result)
}

func (s *Server) searchByKeywords(ctx context.Context, keywords []string) (*logseq.SearchResult, error) {
	if len(keywords) == 0 {
		return &logseq.SearchResult{}, nil
	}

	first, err := s.client.Search(ctx, keywords[0])
	if err != nil {
		return nil, err
	}
	if first == nil {
		return &logseq.SearchResult{}, nil
	}

	result := cloneSearchResult(first)
	for _, kw := range keywords[1:] {
		next, err := s.client.Search(ctx, kw)
		if err != nil {
			return nil, err
		}
		result = intersectSearchResults(result, next)
		if len(result.Blocks) == 0 && len(result.Pages) == 0 && len(result.Files) == 0 {
			break
		}
	}

	return result, nil
}

func splitSearchKeywords(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	out := make([]string, 0)
	var buf strings.Builder
	inQuote := false
	var quote rune

	flush := func() {
		v := strings.TrimSpace(buf.String())
		buf.Reset()
		if v != "" {
			out = append(out, v)
		}
	}

	for _, r := range raw {
		switch {
		case (r == '"' || r == '\'') && !inQuote:
			inQuote = true
			quote = r
		case inQuote && r == quote:
			inQuote = false
			quote = 0
		case !inQuote && (r == ' ' || r == '\t' || r == '\n'):
			flush()
		default:
			buf.WriteRune(r)
		}
	}
	flush()

	if len(out) == 0 {
		return []string{raw}
	}

	return dedupeStrings(out)
}

func (s *Server) getSearchPageSetByTag(ctx context.Context, tag string) (map[string]struct{}, error) {
	tag = normalizeTag(tag)
	if tag == "" {
		return nil, nil
	}

	if set, ok := s.getPageNameSetByTag(ctx, tag); ok {
		return set, nil
	}

	pages, err := s.getAllPages(ctx)
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{})
	for _, p := range pages {
		if !matchesTagFilter(p, tag) {
			continue
		}
		for _, name := range []string{p.Name, p.OriginalName, p.DisplayName()} {
			k := strings.ToLower(strings.TrimSpace(name))
			if k == "" {
				continue
			}
			set[k] = struct{}{}
		}
	}

	return set, nil
}

func filterSearchResultByPageSet(in *logseq.SearchResult, pageSet map[string]struct{}) *logseq.SearchResult {
	if in == nil {
		return nil
	}
	if len(pageSet) == 0 {
		return &logseq.SearchResult{}
	}

	out := &logseq.SearchResult{
		Blocks: make([]logseq.SearchBlock, 0, len(in.Blocks)),
		Pages:  make([]string, 0, len(in.Pages)),
		Files:  nil,
	}

	for _, b := range in.Blocks {
		if pageNameInSet(b.Page, pageSet) {
			out.Blocks = append(out.Blocks, b)
		}
	}

	for _, p := range in.Pages {
		if pageNameInSet(p, pageSet) {
			out.Pages = append(out.Pages, p)
		}
	}

	return out
}

func cloneSearchResult(in *logseq.SearchResult) *logseq.SearchResult {
	if in == nil {
		return &logseq.SearchResult{}
	}
	out := &logseq.SearchResult{}
	out.Pages = append(out.Pages, in.Pages...)
	out.Files = append(out.Files, in.Files...)
	out.Blocks = append(out.Blocks, in.Blocks...)
	return out
}

func intersectSearchResults(left, right *logseq.SearchResult) *logseq.SearchResult {
	if left == nil || right == nil {
		return &logseq.SearchResult{}
	}

	pageSet := make(map[string]struct{}, len(right.Pages))
	for _, p := range right.Pages {
		k := strings.ToLower(strings.TrimSpace(p))
		if k != "" {
			pageSet[k] = struct{}{}
		}
	}

	fileSet := make(map[string]struct{}, len(right.Files))
	for _, f := range right.Files {
		k := strings.ToLower(strings.TrimSpace(f))
		if k != "" {
			fileSet[k] = struct{}{}
		}
	}

	blockSet := make(map[string]struct{}, len(right.Blocks))
	for _, b := range right.Blocks {
		k := searchBlockKey(b)
		if k != "" {
			blockSet[k] = struct{}{}
		}
	}

	out := &logseq.SearchResult{
		Blocks: make([]logseq.SearchBlock, 0, len(left.Blocks)),
		Pages:  make([]string, 0, len(left.Pages)),
		Files:  make([]string, 0, len(left.Files)),
	}

	for _, p := range left.Pages {
		k := strings.ToLower(strings.TrimSpace(p))
		if _, ok := pageSet[k]; ok {
			out.Pages = append(out.Pages, p)
		}
	}

	for _, f := range left.Files {
		k := strings.ToLower(strings.TrimSpace(f))
		if _, ok := fileSet[k]; ok {
			out.Files = append(out.Files, f)
		}
	}

	for _, b := range left.Blocks {
		if _, ok := blockSet[searchBlockKey(b)]; ok {
			out.Blocks = append(out.Blocks, b)
		}
	}

	return out
}

func searchBlockKey(b logseq.SearchBlock) string {
	if uuid := strings.ToLower(strings.TrimSpace(b.UUID)); uuid != "" {
		return "uuid:" + uuid
	}
	content := strings.ToLower(strings.TrimSpace(b.Content))
	page := strings.ToLower(strings.TrimSpace(b.Page))
	if content == "" && page == "" {
		return ""
	}
	return "cp:" + page + "|" + content
}

func filterEmptyBlocks(blocks []logseq.SearchBlock) []logseq.SearchBlock {
	result := make([]logseq.SearchBlock, 0, len(blocks))
	for _, b := range blocks {
		if b.UUID != "" || b.Content != "" {
			result = append(result, b)
		}
	}
	return result
}

func pageNameInSet(name string, set map[string]struct{}) bool {
	k := strings.ToLower(strings.TrimSpace(name))
	if k == "" {
		return false
	}
	_, ok := set[k]
	return ok
}

// resolveSearchBlockPageIDs replaces integer page IDs (like "#12345") with
// actual page names by batch-querying Logseq's Datalog store.
func (s *Server) resolveSearchBlockPageIDs(ctx context.Context, result *logseq.SearchResult) {
	if result == nil || len(result.Blocks) == 0 {
		return
	}

	// Collect unique DB IDs that need resolution
	ids := make(map[int64]struct{})
	for _, b := range result.Blocks {
		if id, ok := parsePageDBID(b.Page); ok {
			ids[id] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return
	}

	// Build a single Datalog query to resolve all IDs at once:
	// [:find ?id ?name :where (or [?id ...]) [?id :block/name ?name]]
	var orClauses strings.Builder
	for id := range ids {
		if orClauses.Len() > 0 {
			orClauses.WriteString(" ")
		}
		fmt.Fprintf(&orClauses, "[(= ?id %d)]", id)
	}
	query := fmt.Sprintf(
		"[:find ?id ?name :where (or %s) [?id :block/name ?name]]",
		orClauses.String(),
	)

	raw, err := s.client.DatascriptQuery(ctx, query)
	if err != nil {
		return // best-effort: keep the #ID if resolution fails
	}

	nameMap := parseIDNamePairs(raw)
	if len(nameMap) == 0 {
		return
	}

	// Replace page IDs with resolved names
	for i := range result.Blocks {
		if id, ok := parsePageDBID(result.Blocks[i].Page); ok {
			if name, found := nameMap[id]; found {
				result.Blocks[i].Page = name
			}
		}
	}
}

// parsePageDBID checks if a page string is a DB integer ID like "#12345" and
// returns the numeric value.
func parsePageDBID(page string) (int64, bool) {
	if !strings.HasPrefix(page, "#") {
		return 0, false
	}
	id, err := strconv.ParseInt(page[1:], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// parseIDNamePairs parses Datalog results of [[id, "name"], ...] into a map.
func parseIDNamePairs(raw []byte) map[int64]string {
	if len(raw) == 0 {
		return nil
	}
	var rows [][]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil
	}
	m := make(map[int64]string, len(rows))
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		var id int64
		switch v := row[0].(type) {
		case float64:
			id = int64(v)
		case json.Number:
			n, err := v.Int64()
			if err != nil {
				continue
			}
			id = n
		default:
			continue
		}
		if name, ok := row[1].(string); ok && name != "" {
			m[id] = name
		}
	}
	return m
}
