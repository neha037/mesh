package handler

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/neha037/mesh/internal/domain"
)

type nodeResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	SourceURL string `json:"source_url,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type listNodesResponse struct {
	Nodes      []nodeResponse `json:"nodes"`
	NextCursor string         `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// HandleListRecent returns the most recently saved nodes.
func (h *Handler) HandleListRecent(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.nodes.ListRecentNodes(r.Context(), 20)
	if err != nil {
		logError(r, "listing recent nodes", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	resp := make([]nodeResponse, len(nodes))
	for i := range nodes {
		resp[i] = toNodeResponse(&nodes[i])
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleListNodes returns a cursor-paginated list of all nodes.
func (h *Handler) HandleListNodes(w http.ResponseWriter, r *http.Request) {
	perPage64, _ := strconv.ParseInt(r.URL.Query().Get("per_page"), 10, 32)
	perPage := int32(perPage64)
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	params := domain.ListNodesParams{
		Limit: perPage,
	}

	if b64c := r.URL.Query().Get("cursor"); b64c != "" {
		cBytes, err := base64.URLEncoding.DecodeString(b64c)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody("invalid cursor format"))
			return
		}
		parts := strings.SplitN(string(cBytes), "|", 2)
		if len(parts) > 0 && parts[0] != "" {
			t, err := time.Parse(time.RFC3339Nano, parts[0])
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorBody("invalid cursor time format"))
				return
			}
			params.CursorAt = &t
		}
		if len(parts) == 2 && parts[1] != "" {
			id := parts[1]
			params.CursorID = &id
		}
	}

	result, err := h.nodes.ListNodes(r.Context(), params)
	if err != nil {
		logError(r, "listing nodes", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	items := make([]nodeResponse, len(result.Nodes))
	for i := range result.Nodes {
		items[i] = toNodeResponse(&result.Nodes[i])
	}

	resp := listNodesResponse{
		Nodes:   items,
		HasMore: result.HasMore,
	}
	if result.HasMore && len(result.Nodes) > 0 {
		last := result.Nodes[len(result.Nodes)-1]
		cStr := last.CreatedAt.Format(time.RFC3339Nano) + "|" + last.ID
		resp.NextCursor = base64.URLEncoding.EncodeToString([]byte(cStr))
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleGetNode retrieves a single node by its ID.
func (h *Handler) HandleGetNode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	node, err := h.nodes.GetNode(r.Context(), idStr)
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorBody("node not found"))
		return
	}
	if err != nil {
		logError(r, "getting node", "error", err, "id", idStr)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, toNodeResponse(&node))
}

// HandleDeleteNode deletes a node by ID.
func (h *Handler) HandleDeleteNode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	err := h.nodes.DeleteNode(r.Context(), idStr)
	if errors.Is(err, domain.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorBody("node not found"))
		return
	}
	if err != nil {
		logError(r, "deleting node", "error", err, "id", idStr)
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toNodeResponse(n *domain.Node) nodeResponse {
	resp := nodeResponse{
		ID:        n.ID,
		Type:      n.Type,
		Title:     n.Title,
		Status:    n.Status,
		CreatedAt: n.CreatedAt.Format(time.RFC3339Nano),
	}
	if n.SourceURL != "" {
		resp.SourceURL = n.SourceURL
	}
	return resp
}
