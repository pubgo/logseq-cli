package webui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

type Server struct {
	client         *logseq.Client
	connectionInfo any
}

type apiResponse struct {
	OK    bool `json:"ok"`
	Data  any  `json:"data,omitempty"`
	Error any  `json:"error,omitempty"`
}

func New(client *logseq.Client, connectionInfo any) *Server {
	return &Server{client: client, connectionInfo: connectionInfo}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/connection", s.handleConnection)
	mux.HandleFunc("/api/graph", s.handleGraph)
	mux.HandleFunc("/api/graph/config", s.handleGraphConfig)
	mux.HandleFunc("/api/graph/app-info", s.handleGraphAppInfo)
	mux.HandleFunc("/api/graph/user-info", s.handleGraphUserInfo)
	mux.HandleFunc("/api/graph/current-config", s.handleGraphCurrentConfig)
	mux.HandleFunc("/api/graph/favorites", s.handleGraphFavorites)
	mux.HandleFunc("/api/graph/recent", s.handleGraphRecent)
	mux.HandleFunc("/api/graph/templates", s.handleGraphTemplates)
	mux.HandleFunc("/api/graph/state", s.handleGraphState)
	mux.HandleFunc("/api/pages", s.handlePages)
	mux.HandleFunc("/api/pages/filter", s.handlePagesFilter)
	mux.HandleFunc("/api/tags", s.handleTags)
	mux.HandleFunc("/api/tag", s.handleTagGet)
	mux.HandleFunc("/api/tag/search", s.handleTagSearch)
	mux.HandleFunc("/api/tag/create", s.handleTagCreate)
	mux.HandleFunc("/api/tag/objects", s.handleTagObjects)
	mux.HandleFunc("/api/tag/property", s.handleTagPropertyRelation)
	mux.HandleFunc("/api/tag/extends", s.handleTagExtendsRelation)
	mux.HandleFunc("/api/tag/block", s.handleTagBlockRelation)
	mux.HandleFunc("/api/property/list", s.handlePropertyList)
	mux.HandleFunc("/api/property", s.handlePropertyGet)
	mux.HandleFunc("/api/property/upsert", s.handlePropertyUpsert)
	mux.HandleFunc("/api/property/remove", s.handlePropertyRemove)
	mux.HandleFunc("/api/page", s.handlePage)
	mux.HandleFunc("/api/page/journal", s.handlePageJournal)
	mux.HandleFunc("/api/page/rename", s.handlePageRename)
	mux.HandleFunc("/api/page/refs", s.handlePageRefs)
	mux.HandleFunc("/api/page/namespace", s.handlePageNamespace)
	mux.HandleFunc("/api/page/properties", s.handlePageProperties)
	mux.HandleFunc("/api/block/current", s.handleBlockCurrent)
	mux.HandleFunc("/api/block/selected", s.handleBlockSelected)
	mux.HandleFunc("/api/block/selected/clear", s.handleBlockClearSelected)
	mux.HandleFunc("/api/block/new-uuid", s.handleBlockNewUUID)
	mux.HandleFunc("/api/block/prev-sibling", s.handleBlockPrevSibling)
	mux.HandleFunc("/api/block/next-sibling", s.handleBlockNextSibling)
	mux.HandleFunc("/api/block", s.handleBlockGet)
	mux.HandleFunc("/api/block/insert", s.handleBlockInsert)
	mux.HandleFunc("/api/block/append", s.handleBlockAppend)
	mux.HandleFunc("/api/block/prepend", s.handleBlockPrepend)
	mux.HandleFunc("/api/block/update", s.handleBlockUpdate)
	mux.HandleFunc("/api/block/remove", s.handleBlockRemove)
	mux.HandleFunc("/api/block/move", s.handleBlockMove)
	mux.HandleFunc("/api/block/property", s.handleBlockProperty)
	mux.HandleFunc("/api/block/collapse", s.handleBlockCollapse)
	mux.HandleFunc("/api/query/datalog", s.handleQueryDatalog)
	mux.HandleFunc("/api/query/dsl", s.handleQueryDSL)
	mux.HandleFunc("/api/search", s.handleSearch)
	return mux
}

func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeOK(w, map[string]any{"status": "ok"})
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeOK(w, s.connectionInfo)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := s.client.GetCurrentGraph(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, info)
}

func (s *Server) handleGraphConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	config, err := s.client.GetUserConfigs(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, config)
}

func (s *Server) handleGraphAppInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := s.client.GetInfo(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, info)
}

func (s *Server) handleGraphUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	info, err := s.client.GetUserInfo(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, info)
}

func (s *Server) handleGraphCurrentConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	config, err := s.client.GetCurrentGraphConfigs(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, config)
}

func (s *Server) handleGraphFavorites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := s.client.GetCurrentGraphFavorites(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, items)
}

func (s *Server) handleGraphRecent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := s.client.GetCurrentGraphRecent(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, items)
}

func (s *Server) handleGraphTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := s.client.GetCurrentGraphTemplates(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, items)
}

type graphStateSetRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func (s *Server) handleGraphState(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		key := strings.TrimSpace(r.URL.Query().Get("key"))
		if key == "" {
			writeError(w, http.StatusBadRequest, "missing query: key")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		value, err := s.client.GetStateFromStore(ctx, key)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, value)
	case http.MethodPost:
		var req graphStateSetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		req.Key = strings.TrimSpace(req.Key)
		if req.Key == "" {
			writeError(w, http.StatusBadRequest, "key is required")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := s.client.SetStateFromStore(ctx, req.Key, req.Value); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, map[string]any{"message": "state updated", "key": req.Key})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

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

func (s *Server) handleTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tags, err := s.getAllTags(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeOK(w, tags)
}

func (s *Server) handleTagGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	nameOrID := strings.TrimSpace(r.URL.Query().Get("nameOrID"))
	if nameOrID == "" {
		writeError(w, http.StatusBadRequest, "missing query: nameOrID")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	tag, err := s.client.GetTag(ctx, nameOrID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if tag == nil {
		writeError(w, http.StatusNotFound, "tag not found")
		return
	}
	writeOK(w, tag)
}

func (s *Server) handleTagSearch(w http.ResponseWriter, r *http.Request) {
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

	tags, err := s.client.GetTagsByName(ctx, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, tags)
}

type createTagRequest struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

func (s *Server) handleTagCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.UUID = strings.TrimSpace(req.UUID)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var (
		tag *logseq.Page
		err error
	)
	if req.UUID != "" {
		tag, err = s.client.CreateTag(ctx, req.Name, req.UUID)
	} else {
		tag, err = s.client.CreateTag(ctx, req.Name)
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, tag)
}

func (s *Server) handleTagObjects(w http.ResponseWriter, r *http.Request) {
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

	items, err := s.client.GetTagObjects(ctx, name)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, items)
}

type tagRelationRequest struct {
	Action string `json:"action"`
	TagID  string `json:"tagID"`
	Target string `json:"target"`
}

func (s *Server) handleTagPropertyRelation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req tagRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.TagID = strings.TrimSpace(req.TagID)
	req.Target = strings.TrimSpace(req.Target)
	if req.TagID == "" || req.Target == "" {
		writeError(w, http.StatusBadRequest, "tagID and target are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var err error
	if req.Action == "remove" {
		err = s.client.RemoveTagProperty(ctx, req.TagID, req.Target)
	} else {
		err = s.client.AddTagProperty(ctx, req.TagID, req.Target)
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "tag property relation updated"})
}

func (s *Server) handleTagExtendsRelation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req tagRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.TagID = strings.TrimSpace(req.TagID)
	req.Target = strings.TrimSpace(req.Target)
	if req.TagID == "" || req.Target == "" {
		writeError(w, http.StatusBadRequest, "tagID and target are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var err error
	if req.Action == "remove" {
		err = s.client.RemoveTagExtends(ctx, req.TagID, req.Target)
	} else {
		err = s.client.AddTagExtends(ctx, req.TagID, req.Target)
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "tag extends relation updated"})
}

type tagBlockRelationRequest struct {
	Action  string `json:"action"`
	BlockID string `json:"blockID"`
	TagID   string `json:"tagID"`
}

func (s *Server) handleTagBlockRelation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req tagBlockRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Action = strings.ToLower(strings.TrimSpace(req.Action))
	req.BlockID = strings.TrimSpace(req.BlockID)
	req.TagID = strings.TrimSpace(req.TagID)
	if req.BlockID == "" || req.TagID == "" {
		writeError(w, http.StatusBadRequest, "blockID and tagID are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	var err error
	if req.Action == "remove" {
		err = s.client.RemoveBlockTag(ctx, req.BlockID, req.TagID)
	} else {
		err = s.client.AddBlockTag(ctx, req.BlockID, req.TagID)
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "block tag relation updated"})
}

func (s *Server) handlePropertyList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	items, err := s.client.GetAllProperties(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, items)
}

func (s *Server) handlePropertyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		writeError(w, http.StatusBadRequest, "missing query: key")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	item, err := s.client.GetProperty(ctx, key)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, item)
}

type propertyUpsertRequest struct {
	Key    string         `json:"key"`
	Schema map[string]any `json:"schema"`
}

func (s *Server) handlePropertyUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req propertyUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Key = strings.TrimSpace(req.Key)
	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result, err := s.client.UpsertProperty(ctx, req.Key, req.Schema, nil)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, result)
}

type propertyRemoveRequest struct {
	Key string `json:"key"`
}

func (s *Server) handlePropertyRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req propertyRemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Key = strings.TrimSpace(req.Key)
	if req.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.client.RemoveProperty(ctx, req.Key); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "property removed", "key": req.Key})
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

func (s *Server) getAllTags(ctx context.Context) ([]string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return s.client.GetAllTags(queryCtx)
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

type appendBlockRequest struct {
	Page    string `json:"page"`
	Content string `json:"content"`
}

func (s *Server) handleBlockAppend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req appendBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Page = strings.TrimSpace(req.Page)
	req.Content = strings.TrimSpace(req.Content)
	if req.Page == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "page and content are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	block, err := s.client.AppendBlockInPage(ctx, req.Page, req.Content)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, block)
}

type updateBlockRequest struct {
	UUID    string `json:"uuid"`
	Content string `json:"content"`
}

func (s *Server) handleBlockUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req updateBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.UUID = strings.TrimSpace(req.UUID)
	req.Content = strings.TrimSpace(req.Content)
	if req.UUID == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "uuid and content are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.client.UpdateBlock(ctx, req.UUID, req.Content); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeOK(w, map[string]any{
		"message": "updated block: " + req.UUID,
	})
}

type removeBlockRequest struct {
	UUID string `json:"uuid"`
}

func (s *Server) handleBlockRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req removeBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.UUID = strings.TrimSpace(req.UUID)
	if req.UUID == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.client.RemoveBlock(ctx, req.UUID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeOK(w, map[string]any{
		"message": "removed block: " + req.UUID,
	})
}

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

	result, err := s.client.Search(ctx, q)
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

	writeOK(w, result)
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
		Files:  nil, // files usually lack page metadata; hide them under tag filtering to avoid noise.
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

func pageNameInSet(name string, set map[string]struct{}) bool {
	k := strings.ToLower(strings.TrimSpace(name))
	if k == "" {
		return false
	}
	_, ok := set[k]
	return ok
}

type queryRequest struct {
	Query string `json:"query"`
}

func (s *Server) handleQueryDatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query, ok := parseQueryRequest(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	result, err := s.client.DatascriptQuery(ctx, query)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeRawJSONOrFallback(w, result)
}

func (s *Server) handleQueryDSL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query, ok := parseQueryRequest(w, r)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	result, err := s.client.DSLQuery(ctx, query)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeRawJSONOrFallback(w, result)
}

func parseQueryRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return "", false
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return "", false
	}
	return query, true
}

// === Page: rename ===

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

// === Page: refs (backlinks) ===

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

// === Page: namespace ===

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

// === Page: properties ===

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

// === Block: get ===

func (s *Server) handleBlockCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	block, err := s.client.GetCurrentBlock(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if block == nil {
		writeError(w, http.StatusNotFound, "no current block")
		return
	}
	writeOK(w, block)
}

func (s *Server) handleBlockSelected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	blocks, err := s.client.GetSelectedBlocks(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, blocks)
}

func (s *Server) handleBlockClearSelected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := s.client.ClearSelectedBlocks(ctx); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "selection cleared"})
}

func (s *Server) handleBlockNewUUID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	uuid, err := s.client.NewBlockUUID(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"uuid": uuid})
}

func (s *Server) handleBlockPrevSibling(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "missing query: uuid")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	block, err := s.client.GetPreviousSiblingBlock(ctx, uuid)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if block == nil {
		writeError(w, http.StatusNotFound, "previous sibling not found")
		return
	}
	writeOK(w, block)
}

func (s *Server) handleBlockNextSibling(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "missing query: uuid")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	block, err := s.client.GetNextSiblingBlock(ctx, uuid)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if block == nil {
		writeError(w, http.StatusNotFound, "next sibling not found")
		return
	}
	writeOK(w, block)
}

func (s *Server) handleBlockGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
	if uuid == "" {
		writeError(w, http.StatusBadRequest, "missing query: uuid")
		return
	}
	children := strings.TrimSpace(r.URL.Query().Get("children")) == "true"
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	block, err := s.client.GetBlock(ctx, uuid, children)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if block == nil {
		writeError(w, http.StatusNotFound, "block not found: "+uuid)
		return
	}
	writeOK(w, block)
}

// === Block: insert ===

type insertBlockRequest struct {
	TargetUUID string `json:"targetUUID"`
	Content    string `json:"content"`
	Sibling    bool   `json:"sibling"`
}

func (s *Server) handleBlockInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req insertBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.TargetUUID = strings.TrimSpace(req.TargetUUID)
	req.Content = strings.TrimSpace(req.Content)
	if req.TargetUUID == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "targetUUID and content are required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	block, err := s.client.InsertBlock(ctx, req.TargetUUID, req.Content, &logseq.InsertBlockOptions{Sibling: req.Sibling})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, block)
}

// === Block: prepend ===

type prependBlockRequest struct {
	Page    string `json:"page"`
	Content string `json:"content"`
}

func (s *Server) handleBlockPrepend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req prependBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Page = strings.TrimSpace(req.Page)
	req.Content = strings.TrimSpace(req.Content)
	if req.Page == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "page and content are required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	block, err := s.client.PrependBlockInPage(ctx, req.Page, req.Content)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, block)
}

// === Block: move ===

type moveBlockRequest struct {
	SrcUUID    string `json:"srcUUID"`
	TargetUUID string `json:"targetUUID"`
	Before     bool   `json:"before"`
}

func (s *Server) handleBlockMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req moveBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.SrcUUID = strings.TrimSpace(req.SrcUUID)
	req.TargetUUID = strings.TrimSpace(req.TargetUUID)
	if req.SrcUUID == "" || req.TargetUUID == "" {
		writeError(w, http.StatusBadRequest, "srcUUID and targetUUID are required")
		return
	}
	var opts map[string]any
	if req.Before {
		opts = map[string]any{"before": true}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := s.client.MoveBlock(ctx, req.SrcUUID, req.TargetUUID, opts); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, map[string]any{"message": "moved block: " + req.SrcUUID + " -> " + req.TargetUUID})
}

// === Block: property (get/set/remove) ===

type blockPropertyRequest struct {
	UUID   string `json:"uuid"`
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Action string `json:"action"` // "get", "set", "remove"
}

func (s *Server) handleBlockProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		uuid := strings.TrimSpace(r.URL.Query().Get("uuid"))
		if uuid == "" {
			writeError(w, http.StatusBadRequest, "missing query: uuid")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		props, err := s.client.GetBlockProperties(ctx, uuid)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, props)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req blockPropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.UUID = strings.TrimSpace(req.UUID)
	req.Key = strings.TrimSpace(req.Key)
	if req.UUID == "" || req.Key == "" {
		writeError(w, http.StatusBadRequest, "uuid and key are required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	switch req.Action {
	case "remove":
		if err := s.client.RemoveBlockProperty(ctx, req.UUID, req.Key); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, map[string]any{"message": "removed property: " + req.Key})
	default: // "set" or empty
		if err := s.client.UpsertBlockProperty(ctx, req.UUID, req.Key, req.Value); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, map[string]any{"message": "set property: " + req.Key})
	}
}

// === Block: collapse/expand ===

type blockCollapseRequest struct {
	UUID      string `json:"uuid"`
	Collapsed bool   `json:"collapsed"`
}

func (s *Server) handleBlockCollapse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req blockCollapseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.UUID = strings.TrimSpace(req.UUID)
	if req.UUID == "" {
		writeError(w, http.StatusBadRequest, "uuid is required")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	if err := s.client.SetBlockCollapsed(ctx, req.UUID, req.Collapsed); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	action := "collapsed"
	if !req.Collapsed {
		action = "expanded"
	}
	writeOK(w, map[string]any{"message": action + " block: " + req.UUID})
}

func writeRawJSONOrFallback(w http.ResponseWriter, raw json.RawMessage) {
	if len(raw) == 0 || string(raw) == "null" {
		writeOK(w, nil)
		return
	}

	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		writeOK(w, string(raw))
		return
	}
	writeOK(w, data)
}

func writeOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, err any) {
	writeJSON(w, status, apiResponse{OK: false, Error: err})
}

func writeJSON(w http.ResponseWriter, status int, resp apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func Run(ctx context.Context, addr string, open bool, client *logseq.Client, connectionInfo any) error {
	srv := New(client, connectionInfo)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if open {
		go func() {
			time.Sleep(250 * time.Millisecond)
			_ = OpenBrowser("http://" + addr)
		}()
	}

	stop := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = httpSrv.Shutdown(shutdownCtx)
		case <-stop:
		}
	}()

	err := httpSrv.ListenAndServe()
	close(stop)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
