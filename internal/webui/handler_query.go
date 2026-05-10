package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

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
