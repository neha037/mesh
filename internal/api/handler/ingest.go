package handler

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const maxContentLen = 500_000 // 500 KB

var validNodeTypes = map[string]bool{
	"article": true, "book": true, "hobby": true,
	"thought": true, "journal": true, "image": true,
	"wildcard": true,
}

type ingestRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type ingestResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	Updated   bool   `json:"updated,omitempty"`
}

var htmlPolicy = bluemonday.UGCPolicy()

// HandleIngestRaw accepts a URL, title, and page content, and stores it as a raw node.
func (h *Handler) HandleIngestRaw(w http.ResponseWriter, r *http.Request) {
	var req ingestRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Type = strings.TrimSpace(req.Type)

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("url is required"))
		return
	}
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid url format"))
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("title is required"))
		return
	}
	if req.Type == "" {
		req.Type = "article"
	}
	if !validNodeTypes[req.Type] {
		writeJSON(w, http.StatusBadRequest, errorBody("type must be one of: article, book, hobby, thought, journal, image, wildcard"))
		return
	}
	if len(req.Content) > maxContentLen {
		writeJSON(w, http.StatusBadRequest, errorBody("content exceeds 500KB limit"))
		return
	}

	req.Title = htmlPolicy.Sanitize(req.Title)
	req.Content = htmlPolicy.Sanitize(req.Content)

	result, err := h.nodes.UpsertRawNode(r.Context(), req.Type, req.Title, req.Content, req.URL)
	if err != nil {
		logError(r, "storing node", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}

	writeJSON(w, status, ingestResponse{
		ID:        result.Node.ID,
		Title:     result.Node.Title,
		CreatedAt: result.Node.CreatedAt.Format(time.RFC3339Nano),
		Updated:   !result.Created,
	})
}
