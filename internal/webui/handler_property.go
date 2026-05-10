package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

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
