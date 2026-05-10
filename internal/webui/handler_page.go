package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

func (s *Server) handlePages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pages, err := s.getAllPages(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, pages)
}

func (s *Server) handlePagesFilter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	propertyKey := strings.TrimSpace(q.Get("property"))
	propertyValue := strings.TrimSpace(q.Get("value"))
	mode := strings.ToLower(strings.TrimSpace(q.Get("mode")))
	if mode != "equals" {
		mode = "contains"
	}
	tag := strings.TrimSpace(q.Get("tag"))
	name := strings.TrimSpace(q.Get("name"))
	includeJournal := parseBoolOrDefault(strings.TrimSpace(q.Get("includeJournal")), true)

	taggedPages, hasTagIndex := s.getPageNameSetByTag(r.Context(), tag)

	pages, err := s.getAllPages(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	filtered := make([]logseq.Page, 0, len(pages))
	for _, p := range pages {
		if !includeJournal && p.IsJournal {
			continue
		}
		if !matchesNameFilter(p, name) {
			continue
		}
		if !matchesTagFilterWithIndex(p, tag, taggedPages, hasTagIndex) {
			continue
		}
		if !matchesMetadataFilter(p, propertyKey, propertyValue, mode) {
			continue
		}
		filtered = append(filtered, p)
	}

	writeOK(w, map[string]any{
		"items":         filtered,
		"total":         len(pages),
		"filteredTotal": len(filtered),
	})
}

func (s *Server) getAllPages(ctx context.Context) ([]logseq.Page, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	pages, err := s.client.GetAllPages(ctx)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

func parseDatalogSingleColumnStrings(raw json.RawMessage) []string {
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

func extractTagsFromPage(p logseq.Page) []string {
	if p.Properties == nil {
		return nil
	}

	candidates := make([]string, 0)
	for k, v := range p.Properties {
		key := strings.ToLower(strings.TrimSpace(k))
		if key != "tags" && key != "tag" {
			continue
		}
		candidates = append(candidates, anyToStringSlice(v)...)
	}

	tags := make([]string, 0, len(candidates))
	for _, c := range candidates {
		t := normalizeTag(c)
		if t != "" {
			tags = append(tags, t)
		}
	}

	return dedupeStrings(tags)
}

func normalizeTag(s string) string {
	v := strings.TrimSpace(strings.ToLower(s))
	v = strings.TrimPrefix(v, "#")
	v = strings.TrimSpace(v)
	return v
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func matchesNameFilter(p logseq.Page, name string) bool {
	if strings.TrimSpace(name) == "" {
		return true
	}
	needle := strings.ToLower(strings.TrimSpace(name))
	display := strings.ToLower(strings.TrimSpace(p.DisplayName()))
	return strings.Contains(display, needle)
}

func matchesTagFilter(p logseq.Page, tag string) bool {
	tag = normalizeTag(tag)
	if tag == "" {
		return true
	}
	for _, t := range extractTagsFromPage(p) {
		if normalizeTag(t) == tag {
			return true
		}
	}
	return false
}

func matchesTagFilterWithIndex(p logseq.Page, tag string, taggedPages map[string]struct{}, hasIndex bool) bool {
	tag = normalizeTag(tag)
	if tag == "" {
		return true
	}

	if hasIndex {
		if pageInNameSet(p, taggedPages) {
			return true
		}
	}

	return matchesTagFilter(p, tag)
}

func pageInNameSet(p logseq.Page, set map[string]struct{}) bool {
	if len(set) == 0 {
		return false
	}

	candidates := []string{p.Name, p.OriginalName, p.DisplayName()}
	for _, c := range candidates {
		k := strings.ToLower(strings.TrimSpace(c))
		if k == "" {
			continue
		}
		if _, ok := set[k]; ok {
			return true
		}
	}

	return false
}

func (s *Server) getPageNameSetByTag(ctx context.Context, tag string) (map[string]struct{}, bool) {
	tag = normalizeTag(tag)
	if tag == "" {
		return nil, false
	}

	queryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		"[:find ?pageName :where [?page :block/name ?pageName] [?b :block/page ?page] [?b :block/refs ?r] [?r :block/name \"%s\"]]",
		escapeDatalogString(tag),
	)
	raw, err := s.client.DatascriptQuery(queryCtx, query)
	if err != nil {
		return nil, false
	}

	set := make(map[string]struct{})
	for _, name := range parseDatalogSingleColumnStrings(raw) {
		k := strings.ToLower(strings.TrimSpace(name))
		if k == "" {
			continue
		}
		set[k] = struct{}{}
	}

	return set, true
}

func escapeDatalogString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

func matchesMetadataFilter(p logseq.Page, propertyKey, propertyValue, mode string) bool {
	propertyKey = strings.TrimSpace(propertyKey)
	propertyValue = strings.TrimSpace(propertyValue)
	if propertyKey == "" {
		return true
	}

	v, ok := getPropertyValueCaseInsensitive(p.Properties, propertyKey)
	if !ok {
		return false
	}

	if propertyValue == "" {
		return true
	}

	candidates := anyToStringSlice(v)
	if len(candidates) == 0 {
		return false
	}

	needle := strings.ToLower(propertyValue)
	for _, c := range candidates {
		cv := strings.ToLower(strings.TrimSpace(c))
		if mode == "equals" {
			if cv == needle {
				return true
			}
			continue
		}
		if strings.Contains(cv, needle) {
			return true
		}
	}

	return false
}

func getPropertyValueCaseInsensitive(m map[string]any, key string) (any, bool) {
	if len(m) == 0 {
		return nil, false
	}
	if v, ok := m[key]; ok {
		return v, true
	}
	target := strings.ToLower(strings.TrimSpace(key))
	for k, v := range m {
		if strings.ToLower(strings.TrimSpace(k)) == target {
			return v, true
		}
	}
	return nil, false
}

func anyToStringSlice(v any) []string {
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
			out = append(out, anyToStringSlice(item)...)
		}
		return out
	case map[string]any:
		b, err := json.Marshal(x)
		if err != nil {
			return []string{fmt.Sprintf("%v", x)}
		}
		return []string{string(b)}
	default:
		return []string{fmt.Sprintf("%v", x)}
	}
}

func parseBoolOrDefault(raw string, def bool) bool {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return def
	}
	if raw == "1" || raw == "true" || raw == "yes" || raw == "on" {
		return true
	}
	if raw == "0" || raw == "false" || raw == "no" || raw == "off" {
		return false
	}
	return def
}

type createPageRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type createJournalRequest struct {
	Date string `json:"date"`
}

func (s *Server) handlePage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handlePageGet(w, r)
	case http.MethodPost:
		s.handlePageCreate(w, r)
	case http.MethodDelete:
		s.handlePageDelete(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handlePageGet(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing query: name")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	page, err := s.client.GetPage(ctx, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if page == nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("page '%s' not found", name))
		return
	}

	blocks, err := s.client.GetPageBlocksTree(ctx, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeOK(w, map[string]any{
		"page":   page,
		"blocks": blocks,
	})
}

func (s *Server) handlePageCreate(w http.ResponseWriter, r *http.Request) {
	var req createPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	page, err := s.client.CreatePage(ctx, name, nil, &logseq.CreatePageOptions{CreateFirstBlock: true})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	content := strings.TrimSpace(req.Content)
	if content != "" {
		if _, err := s.client.AppendBlockInPage(ctx, name, content); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}
	writeOK(w, page)
}

func (s *Server) handlePageJournal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createJournalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	date := strings.TrimSpace(req.Date)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	page, err := s.client.CreateJournalPage(ctx, date)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, page)
}

func (s *Server) handlePageDelete(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing query: name")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.client.DeletePage(ctx, name); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeOK(w, map[string]any{
		"message": "deleted page: " + name,
	})
}

type renamePageRequest struct {
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

func (s *Server) handlePageRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req renamePageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.OldName = strings.TrimSpace(req.OldName)
	req.NewName = strings.TrimSpace(req.NewName)
	if req.OldName == "" || req.NewName == "" {
		writeError(w, http.StatusBadRequest, "oldName and newName are required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := s.client.RenamePage(ctx, req.OldName, req.NewName); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "renamed: " + req.OldName + " -> " + req.NewName})
}

func (s *Server) handlePageRefs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "missing query: name")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	result, err := s.client.GetPageLinkedReferences(ctx, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeRawJSONOrFallback(w, result)
}

func (s *Server) handlePageNamespace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ns := strings.TrimSpace(r.URL.Query().Get("name"))
	if ns == "" {
		writeError(w, http.StatusBadRequest, "missing query: name")
		return
	}
	tree := strings.TrimSpace(r.URL.Query().Get("tree")) == "true"
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if tree {
		result, err := s.client.GetPagesTreeFromNamespace(ctx, ns)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeRawJSONOrFallback(w, result)
		return
	}

	pages, err := s.client.GetPagesFromNamespace(ctx, ns)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, pages)
}

type setPagePropertiesRequest struct {
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties"`
}

func (s *Server) handlePageProperties(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			writeError(w, http.StatusBadRequest, "missing query: name")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		page, err := s.client.GetPage(ctx, name)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if page == nil {
			writeError(w, http.StatusNotFound, "page not found: "+name)
			return
		}
		props := page.Properties
		if props == nil {
			props = map[string]any{}
		}
		writeOK(w, props)
	case http.MethodPost:
		var req setPagePropertiesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || len(req.Properties) == 0 {
			writeError(w, http.StatusBadRequest, "name and properties are required")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		if err := s.client.SetPageProperties(ctx, req.Name, req.Properties); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, map[string]any{"message": "properties updated"})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
