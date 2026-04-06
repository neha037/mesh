package handler

import (
	"context"

	"github.com/neha037/mesh/internal/domain"
)

// Pinger checks database connectivity.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	nodes domain.NodeRepository
	db    Pinger
}

// New creates a Handler with the given node repository and database pinger.
func New(nodes domain.NodeRepository, db Pinger) *Handler {
	return &Handler{nodes: nodes, db: db}
}
