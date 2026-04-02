package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/neha037/mesh/internal/storage"
)

type ingestRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ingestResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	Updated   bool   `json:"updated,omitempty"`
}

type recentNodeResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	SourceURL string `json:"source_url,omitempty"`
	CreatedAt string `json:"created_at"`
}

// HandleIngestRaw accepts a URL, title, and page content, and stores it as a raw node.
func (h *Handler) HandleIngestRaw(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var req ingestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title is required"})
		return
	}

	node, err := h.queries.UpsertRawNode(r.Context(), storage.UpsertRawNodeParams{
		Type:      "article",
		Title:     req.Title,
		Content:   pgtype.Text{String: req.Content, Valid: req.Content != ""},
		SourceUrl: pgtype.Text{String: req.URL, Valid: true},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("storing node: %v", err)})
		return
	}

	status := http.StatusCreated
	if !node.Created {
		status = http.StatusOK
	}

	writeJSON(w, status, ingestResponse{
		ID:        uuidToString(node.ID),
		Title:     node.Title,
		CreatedAt: node.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		Updated:   !node.Created,
	})
}

// HandleListRecent returns the most recently saved nodes.
func (h *Handler) HandleListRecent(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.queries.ListRecentNodes(r.Context(), 20)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("listing nodes: %v", err)})
		return
	}

	resp := make([]recentNodeResponse, len(nodes))
	for i, n := range nodes {
		resp[i] = recentNodeResponse{
			ID:        uuidToString(n.ID),
			Title:     n.Title,
			CreatedAt: n.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if n.SourceUrl.Valid {
			resp[i].SourceURL = n.SourceUrl.String
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

type listNodesResponse struct {
	Nodes      []recentNodeResponse `json:"nodes"`
	NextCursor string               `json:"next_cursor,omitempty"`
	HasMore    bool                 `json:"has_more"`
}

// HandleListNodes returns a cursor-paginated list of all nodes.
func (h *Handler) HandleListNodes(w http.ResponseWriter, r *http.Request) {
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var cursor pgtype.Timestamptz
	if c := r.URL.Query().Get("cursor"); c != "" {
		t, err := time.Parse(time.RFC3339, c)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid cursor format, use RFC3339"})
			return
		}
		cursor = pgtype.Timestamptz{Time: t, Valid: true}
	}

	// Fetch one extra to determine if there are more results.
	nodes, err := h.queries.ListNodes(r.Context(), storage.ListNodesParams{
		Limit:  int32(perPage + 1),
		Cursor: cursor,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("listing nodes: %v", err)})
		return
	}

	hasMore := len(nodes) > perPage
	if hasMore {
		nodes = nodes[:perPage]
	}

	items := make([]recentNodeResponse, len(nodes))
	for i, n := range nodes {
		items[i] = recentNodeResponse{
			ID:        uuidToString(n.ID),
			Title:     n.Title,
			CreatedAt: n.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		if n.SourceUrl.Valid {
			items[i].SourceURL = n.SourceUrl.String
		}
	}

	resp := listNodesResponse{
		Nodes:   items,
		HasMore: hasMore,
	}
	if hasMore && len(items) > 0 {
		resp.NextCursor = items[len(items)-1].CreatedAt
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleDeleteNode deletes a node by ID.
func (h *Handler) HandleDeleteNode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	var uuid pgtype.UUID
	if err := uuid.Scan(idStr); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid node ID"})
		return
	}

	if err := h.queries.DeleteNode(r.Context(), uuid); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("deleting node: %v", err)})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

func uuidToString(u pgtype.UUID) string {
	b := u.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
