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
	mux.HandleFunc("/api/pages", s.handlePages)
	mux.HandleFunc("/api/page", s.handlePage)
	mux.HandleFunc("/api/block/append", s.handleBlockAppend)
	mux.HandleFunc("/api/block/update", s.handleBlockUpdate)
	mux.HandleFunc("/api/block/remove", s.handleBlockRemove)
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

func (s *Server) handlePages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	pages, err := s.client.GetAllPages(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeOK(w, pages)
}

type createPageRequest struct {
	Name string `json:"name"`
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

	writeOK(w, result)
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
