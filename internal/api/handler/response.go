package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// HandleHealth returns a health check that verifies database connectivity.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if err := h.db.Ping(r.Context()); err != nil {
		slog.Error("health check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unhealthy", "error": "database unreachable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writeJSON encode error", "error", err)
	}
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}

// logError logs an error with the request ID for correlation.
func logError(r *http.Request, msg string, args ...any) {
	args = append([]any{"request_id", middleware.GetReqID(r.Context())}, args...)
	slog.Error(msg, args...)
}
