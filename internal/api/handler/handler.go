package handler

import (
	"github.com/neha037/mesh/internal/storage"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	queries *storage.Queries
}

// New creates a Handler with the given storage queries.
func New(queries *storage.Queries) *Handler {
	return &Handler{queries: queries}
}
