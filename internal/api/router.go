package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/neha037/mesh/internal/api/handler"
)

func NewRouter(h *handler.Handler, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Throttle(100))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         300,
	}))

	r.Get("/healthz", h.HandleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(maxRequestBody(1 << 20)) // 1 MB default for all API routes
		r.Post("/ingest/raw", h.HandleIngestRaw)
		r.Post("/ingest/url", h.HandleIngestURL)
		r.Post("/ingest/text", h.HandleIngestText)
		r.Get("/nodes/recent", h.HandleListRecent)
		r.Get("/nodes", h.HandleListNodes)
		r.Get("/nodes/{id}", h.HandleGetNode)
		r.Delete("/nodes/{id}", h.HandleDeleteNode)
	})

	return r
}

// maxRequestBody limits request body size for all routes in the group.
func maxRequestBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
