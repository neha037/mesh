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

func TestHandleIngestURL(t *testing.T) {
	successIngest := &mockIngestService{
		ingestURLFn: func(_ context.Context, url, nodeType string) (domain.IngestURLResult, error) {
			return domain.IngestURLResult{
				NodeID: "node-123",
				JobID:  "job-456",
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
			name:       "valid URL enqueues job",
			body:       `{"url":"https://example.com/article","type":"article"}`,
			ingest:     successIngest,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "defaults type to article",
			body:       `{"url":"https://example.com/page"}`,
			ingest:     successIngest,
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid JSON body",
		},
		{
			name:       "missing url",
			body:       `{"type":"article"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "url is required",
		},
		{
			name:       "invalid url scheme",
			body:       `{"url":"ftp://example.com/file"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "url must be a valid HTTP/HTTPS URL",
		},
		{
			name:       "url without host",
			body:       `{"url":"https://"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "url must be a valid HTTP/HTTPS URL",
		},
		{
			name:       "invalid type",
			body:       `{"url":"https://example.com","type":"invalid"}`,
			ingest:     successIngest,
			wantStatus: http.StatusBadRequest,
			wantError:  "type must be one of: article, book, hobby, thought, journal, image, wildcard",
		},
		{
			name: "service error returns 500",
			body: `{"url":"https://example.com","type":"article"}`,
			ingest: &mockIngestService{
				ingestURLFn: func(_ context.Context, _, _ string) (domain.IngestURLResult, error) {
					return domain.IngestURLResult{}, errors.New("db error")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestHandlerWithIngest(&mockNodeRepo{}, tt.ingest, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/ingest/url", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.HandleIngestURL(rec, req)

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

			if tt.wantStatus == http.StatusAccepted {
				var body map[string]string
				decodeBody(t, rec, &body)
				if body["status"] != "pending" {
					t.Errorf("status = %q, want %q", body["status"], "pending")
				}
				if body["job_id"] == "" {
					t.Error("expected non-empty job_id")
				}
				if body["node_id"] == "" {
					t.Error("expected non-empty node_id")
				}
			}
		})
	}
}
