package handler

import (
	"net/http"
	"strings"
)

type ingestURLRequest struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type ingestURLResponse struct {
	JobID  string `json:"job_id"`
	NodeID string `json:"node_id"`
	Status string `json:"status"`
}

// HandleIngestURL accepts a URL for asynchronous scraping and processing.
func (h *Handler) HandleIngestURL(w http.ResponseWriter, r *http.Request) {
	var req ingestURLRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid JSON body"))
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	req.Type = strings.TrimSpace(req.Type)

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("url is required"))
		return
	}

	if err := validateHTTPURL(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody(err.Error()))
		return
	}

	if req.Type == "" {
		req.Type = "article"
	}
	if !validNodeTypes[req.Type] {
		writeJSON(w, http.StatusBadRequest, errorBody(invalidTypeMsg))
		return
	}

	result, err := h.ingest.IngestURL(r.Context(), req.URL, req.Type)
	if err != nil {
		logError(r, "ingesting URL", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	writeJSON(w, http.StatusAccepted, ingestURLResponse{
		JobID:  result.JobID,
		NodeID: result.NodeID,
		Status: "pending",
	})
}
