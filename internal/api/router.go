package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/neha037/mesh/internal/api/handler"
)

// NewRouter creates an HTTP router with all API routes and middleware.
func NewRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/ingest/raw", h.HandleIngestRaw)
		r.Get("/nodes/recent", h.HandleListRecent)
		r.Get("/nodes", h.HandleListNodes)
		r.Delete("/nodes/{id}", h.HandleDeleteNode)
	})

	return r
}
