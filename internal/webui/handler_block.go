package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

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

	// Fetch updated block to return consistent response
	block, err := s.client.GetBlock(ctx, req.UUID, false)
	if err == nil && block != nil {
		writeOK(w, block)
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

type blockPropertyRequest struct {
	UUID   string `json:"uuid"`
	Key    string `json:"key"`
	Value  any    `json:"value"`
	Action string `json:"action"`
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
	default:
		if err := s.client.UpsertBlockProperty(ctx, req.UUID, req.Key, req.Value); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeOK(w, map[string]any{"message": "set property: " + req.Key})
	}
}

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
