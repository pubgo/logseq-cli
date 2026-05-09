package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pubgo/logseq-cli/pkg/logseq"
)

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

func (s *Server) getAllTags(ctx context.Context) ([]string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	return s.client.GetAllTags(queryCtx)
}
