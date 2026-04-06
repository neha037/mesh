package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/neha037/mesh/internal/domain"
)

func TestHandleIngestText(t *testing.T) {
	successIngest := &mockIngestService{
		ingestTextFn: func(_ context.Context, title, content, nodeType string) (domain.IngestTextResult, error) {
			return domain.IngestTextResult{
				NodeID: "node-789",
				JobID:  "job-012",
			}, nil
		},
	}

	tests := []struct {
		name       string
		body       string
		ingest     *mockIngestService
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid text creates node and job",
			body:       `{"title":"My Thought","content":"Some deep insight","type":"thought"}`,
			ingest:     successIngest,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "defaults type to thought",
			body:       `{"title":"My Thought","content":"Some content"}`,
			ingest:     successIngest,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid JSON body",
		},
		{
			name:       "missing title",
			body:       `{"content":"Some content"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "title is required",
		},
		{
			name:       "missing content",
			body:       `{"title":"My Thought"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "content is required",
		},
		{
			name:       "invalid type",
			body:       `{"title":"Test","content":"Content","type":"invalid"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "type must be one of: article, book, hobby, thought, journal, image, wildcard",
		},
		{
			name: "service error returns 500",
			body: `{"title":"Test","content":"Content","type":"thought"}`,
			ingest: &mockIngestService{
				ingestTextFn: func(_ context.Context, _, _, _ string) (domain.IngestTextResult, error) {
					return domain.IngestTextResult{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandlerWithIngest(&mockNodeRepo{}, tt.ingest, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/text", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.HandleIngestText(rec, req)

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

			if tt.wantStatus == http.StatusCreated {
				var body map[string]string
				decodeBody(t, rec, &body)
				if body["status"] != "pending" {
					t.Errorf("status = %q, want %q", body["status"], "pending")
				}
				if body["node_id"] == "" {
					t.Error("expected non-empty node_id")
				}
				if body["job_id"] == "" {
					t.Error("expected non-empty job_id")
				}
			}
		})
	}
}
