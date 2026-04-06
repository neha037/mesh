package handler

import (
	"net/http"
	"strings"
)

type ingestTextRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type ingestTextResponse struct {
	NodeID string `json:"node_id"`
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// HandleIngestText accepts text content for asynchronous processing.
func (h *Handler) HandleIngestText(w http.ResponseWriter, r *http.Request) {
	var req ingestTextRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Type = strings.TrimSpace(req.Type)

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("title is required"))
		return
	}
	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("content is required"))
		return
	}

	if req.Type == "" {
		req.Type = "thought"
	}
	if !validNodeTypes[req.Type] {
		writeJSON(w, http.StatusBadRequest, errorBody(invalidTypeMsg))
		return
	}

	req.Title = htmlPolicy.Sanitize(req.Title)
	req.Content = htmlPolicy.Sanitize(req.Content)

	result, err := h.ingest.IngestText(r.Context(), req.Title, req.Content, req.Type)
	if err != nil {
		logError(r, "ingesting text", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	writeJSON(w, http.StatusCreated, ingestTextResponse{
		NodeID: result.NodeID,
		JobID:  result.JobID,
		Status: "pending",
	})
}
