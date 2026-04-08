package handler_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/neha037/mesh/internal/api/handler"
	"github.com/neha037/mesh/internal/domain"
)

// --- mocks ---

type mockNodeRepo struct {
	upsertFn     func(ctx context.Context, nodeType, title, content, sourceURL string) (domain.UpsertResult, error)
	getFn        func(ctx context.Context, id string) (domain.Node, error)
	listRecentFn func(ctx context.Context, limit int32) ([]domain.Node, error)
	listFn       func(ctx context.Context, params domain.ListNodesParams) (domain.ListNodesResult, error)
	deleteFn     func(ctx context.Context, id string) error
}

func (m *mockNodeRepo) UpsertRawNode(ctx context.Context, nodeType, title, content, sourceURL string) (domain.UpsertResult, error) {
	return m.upsertFn(ctx, nodeType, title, content, sourceURL)
}

func (m *mockNodeRepo) GetNode(ctx context.Context, id string) (domain.Node, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return domain.Node{}, errors.New("not implemented")
}

func (m *mockNodeRepo) ListRecentNodes(ctx context.Context, limit int32) ([]domain.Node, error) {
	return m.listRecentFn(ctx, limit)
}

func (m *mockNodeRepo) ListNodes(ctx context.Context, params domain.ListNodesParams) (domain.ListNodesResult, error) {
	return m.listFn(ctx, params)
}

func (m *mockNodeRepo) DeleteNode(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func (m *mockNodeRepo) UpdateNodeContent(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockNodeRepo) UpdateNodeStatus(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockNodeRepo) UpdateNodeEmbedding(_ context.Context, _ string, _ []float32, _ int32) (bool, error) {
	return true, nil
}

func (m *mockNodeRepo) GetNodeContent(_ context.Context, _ string) (domain.Node, error) {
	return domain.Node{}, nil
}

func (m *mockNodeRepo) GetNodeEmbedding(_ context.Context, _ string) ([]float32, error) {
	return nil, nil
}

func (m *mockNodeRepo) ResetStaleProcessingNodes(_ context.Context, _ time.Time) (int64, error) {
	return 0, nil
}

func (m *mockNodeRepo) ListNodesWithoutEmbedding(_ context.Context, _ int32) ([]string, error) {
	return nil, nil
}

type mockIngestService struct {
	ingestURLFn  func(ctx context.Context, url, nodeType string) (domain.IngestURLResult, error)
	ingestTextFn func(ctx context.Context, title, content, nodeType string) (domain.IngestTextResult, error)
}

func (m *mockIngestService) IngestURL(ctx context.Context, url, nodeType string) (domain.IngestURLResult, error) {
	if m.ingestURLFn != nil {
		return m.ingestURLFn(ctx, url, nodeType)
	}
	return domain.IngestURLResult{}, errors.New("not implemented")
}

func (m *mockIngestService) IngestText(ctx context.Context, title, content, nodeType string) (domain.IngestTextResult, error) {
	if m.ingestTextFn != nil {
		return m.ingestTextFn(ctx, title, content, nodeType)
	}
	return domain.IngestTextResult{}, errors.New("not implemented")
}

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

// --- helpers ---

var fixedTime = time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)

func newTestHandler(repo *mockNodeRepo, pinger *mockPinger) *handler.Handler {
	if pinger == nil {
		pinger = &mockPinger{}
	}
	return handler.New(repo, &mockIngestService{}, pinger)
}

func newTestHandlerWithIngest(repo *mockNodeRepo, ingest *mockIngestService, pinger *mockPinger) *handler.Handler {
	if pinger == nil {
		pinger = &mockPinger{}
	}
	if ingest == nil {
		ingest = &mockIngestService{}
	}
	return handler.New(repo, ingest, pinger)
}

func decodeBody(t *testing.T, resp *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
}

// --- HandleHealth tests ---

func TestHandleHealth(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "healthy",
			pingErr:    nil,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "unhealthy when db unreachable",
			pingErr:    errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(&mockNodeRepo{}, &mockPinger{err: tt.pingErr})

			req := httptest.NewRequest(http.MethodGet, "/healthz", http.NoBody)
			rec := httptest.NewRecorder()

			h.HandleHealth(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body map[string]string
			decodeBody(t, rec, &body)
			if body["status"] != tt.wantBody {
				t.Errorf("status body = %q, want %q", body["status"], tt.wantBody)
			}
		})
	}
}

// --- HandleIngestRaw tests ---

func TestHandleIngestRaw(t *testing.T) {
	successRepo := &mockNodeRepo{
		upsertFn: func(_ context.Context, nodeType, title, content, sourceURL string) (domain.UpsertResult, error) {
			return domain.UpsertResult{
				Node: domain.Node{
					ID:        "test-id",
					Title:     title,
					CreatedAt: fixedTime,
				},
				Created: true,
			}, nil
		},
	}

	tests := []struct {
		name       string
		body       string
		repo       *mockNodeRepo
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid request creates node",
			body:       `{"url":"https://example.com","title":"Test","content":"Hello"}`,
			repo:       successRepo,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid`,
			repo:       successRepo,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid JSON body",
		},
		{
			name:       "missing url",
			body:       `{"title":"Test"}`,
			repo:       successRepo,
			wantStatus: http.StatusBadRequest,
			wantError:  "url is required",
		},
		{
			name:       "invalid url format",
			body:       `{"url":"not-a-url","title":"Test"}`,
			repo:       successRepo,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid url format",
		},
		{
			name:       "rejects ftp url scheme",
			body:       `{"url":"ftp://example.com","title":"Test"}`,
			repo:       successRepo,
			wantStatus: http.StatusBadRequest,
			wantError:  "url must be a valid HTTP/HTTPS URL",
		},
		{
			name:       "rejects javascript url scheme",
			body:       `{"url":"javascript:alert(1)","title":"Test"}`,
			repo:       successRepo,
			wantStatus: http.StatusBadRequest,
			wantError:  "url must be a valid HTTP/HTTPS URL",
		},
		{
			name:       "missing title",
			body:       `{"url":"https://example.com"}`,
			repo:       successRepo,
			wantStatus: http.StatusBadRequest,
			wantError:  "title is required",
		},
		{
			name: "update returns 200",
			body: `{"url":"https://example.com","title":"Test","content":"Updated"}`,
			repo: &mockNodeRepo{
				upsertFn: func(_ context.Context, _, _, _, _ string) (domain.UpsertResult, error) {
					return domain.UpsertResult{
						Node:    domain.Node{ID: "test-id", Title: "Test", CreatedAt: fixedTime},
						Created: false,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "repo error returns 500",
			body: `{"url":"https://example.com","title":"Test"}`,
			repo: &mockNodeRepo{
				upsertFn: func(_ context.Context, _, _, _, _ string) (domain.UpsertResult, error) {
					return domain.UpsertResult{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
	}

	// Test content exceeding 500KB limit (not table-driven due to large body).
	t.Run("content exceeds 500KB limit", func(t *testing.T) {
		largeContent := strings.Repeat("a", 500_001)
		body := `{"url":"https://example.com","title":"Test","content":"` + largeContent + `"}`
		h := newTestHandler(successRepo, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/raw", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.HandleIngestRaw(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		var errBody map[string]string
		decodeBody(t, rec, &errBody)
		if errBody["error"] != "content exceeds 500KB limit" {
			t.Errorf("error = %q, want %q", errBody["error"], "content exceeds 500KB limit")
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(tt.repo, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/raw", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.HandleIngestRaw(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantError != "" {
				var body map[string]string
				decodeBody(t, rec, &body)
				if body["error"] != tt.wantError {
					t.Errorf("error = %q, want %q", body["error"], tt.wantError)
				}
			}
		})
	}
}

// --- HandleListRecent tests ---

func TestHandleListRecent(t *testing.T) {
	tests := []struct {
		name       string
		repo       *mockNodeRepo
		wantStatus int
		wantLen    int
	}{
		{
			name: "returns nodes",
			repo: &mockNodeRepo{
				listRecentFn: func(_ context.Context, limit int32) ([]domain.Node, error) {
					if limit != 20 {
						t.Errorf("limit = %d, want 20", limit)
					}
					return []domain.Node{
						{ID: "1", Title: "First", CreatedAt: fixedTime},
						{ID: "2", Title: "Second", CreatedAt: fixedTime},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "empty list",
			repo: &mockNodeRepo{
				listRecentFn: func(_ context.Context, _ int32) ([]domain.Node, error) {
					return nil, nil
				},
			},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name: "repo error",
			repo: &mockNodeRepo{
				listRecentFn: func(_ context.Context, _ int32) ([]domain.Node, error) {
					return nil, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(tt.repo, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/recent", http.NoBody)
			rec := httptest.NewRecorder()

			h.HandleListRecent(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var body []map[string]any
				decodeBody(t, rec, &body)
				if len(body) != tt.wantLen {
					t.Errorf("len = %d, want %d", len(body), tt.wantLen)
				}
			}
		})
	}
}

// --- HandleListNodes tests ---

func TestHandleListNodes(t *testing.T) {
	twoNodes := []domain.Node{
		{ID: "1", Title: "First", CreatedAt: fixedTime},
		{ID: "2", Title: "Second", CreatedAt: fixedTime.Add(-time.Hour)},
	}

	tests := []struct {
		name       string
		query      string
		repo       *mockNodeRepo
		wantStatus int
		wantMore   bool
	}{
		{
			name:  "default pagination",
			query: "",
			repo: &mockNodeRepo{
				listFn: func(_ context.Context, p domain.ListNodesParams) (domain.ListNodesResult, error) {
					if p.Limit != 20 {
						t.Errorf("limit = %d, want 20", p.Limit)
					}
					return domain.ListNodesResult{Nodes: twoNodes, HasMore: false}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantMore:   false,
		},
		{
			name:  "custom per_page",
			query: "?per_page=5",
			repo: &mockNodeRepo{
				listFn: func(_ context.Context, p domain.ListNodesParams) (domain.ListNodesResult, error) {
					if p.Limit != 5 {
						t.Errorf("limit = %d, want 5", p.Limit)
					}
					return domain.ListNodesResult{Nodes: twoNodes, HasMore: true}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantMore:   true,
		},
		{
			name:  "with cursor",
			query: "?cursor=" + base64.URLEncoding.EncodeToString([]byte(fixedTime.Format(time.RFC3339Nano)+"|some-id")),
			repo: &mockNodeRepo{
				listFn: func(_ context.Context, p domain.ListNodesParams) (domain.ListNodesResult, error) {
					if p.CursorAt == nil {
						t.Error("expected cursor time to be set")
					}
					if p.CursorID == nil || *p.CursorID != "some-id" {
						t.Error("expected cursor ID to be 'some-id'")
					}
					return domain.ListNodesResult{Nodes: twoNodes, HasMore: false}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "invalid cursor",
			query: "?cursor=not-base64!!!",
			repo: &mockNodeRepo{
				listFn: func(_ context.Context, _ domain.ListNodesParams) (domain.ListNodesResult, error) {
					t.Error("should not reach repo with invalid cursor")
					return domain.ListNodesResult{}, nil
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "repo error",
			query: "",
			repo: &mockNodeRepo{
				listFn: func(_ context.Context, _ domain.ListNodesParams) (domain.ListNodesResult, error) {
					return domain.ListNodesResult{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(tt.repo, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes"+tt.query, http.NoBody)
			rec := httptest.NewRecorder()

			h.HandleListNodes(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var body struct {
					Nodes      []map[string]any `json:"nodes"`
					NextCursor string           `json:"next_cursor"`
					HasMore    bool             `json:"has_more"`
				}
				decodeBody(t, rec, &body)
				if body.HasMore != tt.wantMore {
					t.Errorf("has_more = %v, want %v", body.HasMore, tt.wantMore)
				}
				if tt.wantMore && body.NextCursor == "" {
					t.Error("expected next_cursor when has_more=true")
				}
			}
		})
	}
}

// --- HandleGetNode tests ---

func TestHandleGetNode(t *testing.T) {
	tests := []struct {
		name       string
		nodeID     string
		repo       *mockNodeRepo
		wantStatus int
		wantTitle  string
	}{
		{
			name:   "found",
			nodeID: "abc-123",
			repo: &mockNodeRepo{
				getFn: func(_ context.Context, id string) (domain.Node, error) {
					if id != "abc-123" {
						t.Errorf("id = %q, want %q", id, "abc-123")
					}
					return domain.Node{
						ID:        "abc-123",
						Type:      "article",
						Title:     "Found Node",
						Status:    "pending",
						CreatedAt: fixedTime,
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantTitle:  "Found Node",
		},
		{
			name:   "not found",
			nodeID: "missing",
			repo: &mockNodeRepo{
				getFn: func(_ context.Context, _ string) (domain.Node, error) {
					return domain.Node{}, domain.ErrNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "repo error",
			nodeID: "abc-123",
			repo: &mockNodeRepo{
				getFn: func(_ context.Context, _ string) (domain.Node, error) {
					return domain.Node{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(tt.repo, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/"+tt.nodeID, http.NoBody)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.nodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			h.HandleGetNode(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantTitle != "" {
				var body map[string]any
				decodeBody(t, rec, &body)
				if body["title"] != tt.wantTitle {
					t.Errorf("title = %q, want %q", body["title"], tt.wantTitle)
				}
			}
		})
	}
}

// --- HandleIngestRaw type validation tests ---

func TestHandleIngestRaw_TypeValidation(t *testing.T) {
	successRepo := &mockNodeRepo{
		upsertFn: func(_ context.Context, nodeType, title, _, _ string) (domain.UpsertResult, error) {
			return domain.UpsertResult{
				Node: domain.Node{
					ID:        "test-id",
					Type:      nodeType,
					Title:     title,
					Status:    "pending",
					CreatedAt: fixedTime,
				},
				Created: true,
			}, nil
		},
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "defaults to article when type omitted",
			body:       `{"url":"https://example.com/1","title":"Test"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "accepts book type",
			body:       `{"url":"https://example.com/2","title":"Test","type":"book"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "accepts thought type",
			body:       `{"url":"https://example.com/3","title":"Test","type":"thought"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "accepts image type",
			body:       `{"url":"https://example.com/4","title":"Test","type":"image"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "accepts wildcard type",
			body:       `{"url":"https://example.com/5","title":"Test","type":"wildcard"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "rejects invalid type",
			body:       `{"url":"https://example.com/6","title":"Test","type":"invalid"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "type must be one of: article, book, hobby, thought, journal, image, wildcard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(successRepo, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/raw", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.HandleIngestRaw(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantError != "" {
				var body map[string]string
				decodeBody(t, rec, &body)
				if body["error"] != tt.wantError {
					t.Errorf("error = %q, want %q", body["error"], tt.wantError)
				}
			}
		})
	}
}

// --- HandleDeleteNode tests ---

func TestHandleDeleteNode(t *testing.T) {
	tests := []struct {
		name       string
		nodeID     string
		repo       *mockNodeRepo
		wantStatus int
	}{
		{
			name:   "successful delete",
			nodeID: "abc-123",
			repo: &mockNodeRepo{
				deleteFn: func(_ context.Context, id string) error {
					if id != "abc-123" {
						t.Errorf("id = %q, want %q", id, "abc-123")
					}
					return nil
				},
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "not found",
			nodeID: "missing",
			repo: &mockNodeRepo{
				deleteFn: func(_ context.Context, _ string) error {
					return domain.ErrNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "repo error",
			nodeID: "abc-123",
			repo: &mockNodeRepo{
				deleteFn: func(_ context.Context, _ string) error {
					return errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandler(tt.repo, nil)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/nodes/"+tt.nodeID, http.NoBody)
			// Set chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.nodeID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()

			h.HandleDeleteNode(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
