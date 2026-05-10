package webui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"runtime"
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
