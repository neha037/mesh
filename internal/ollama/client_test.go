package ollama_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/neha037/mesh/internal/ollama"
)

func TestExtractTags_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected path /api/generate, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		resp := map[string]string{
			"response": `{"tags":["machine-learning","neural-networks","deep-learning"],"confidence":0.92}`,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	result, err := client.ExtractTags(context.Background(), "some content")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(result.Tags))
	}
	if result.Confidence != 0.92 {
		t.Errorf("expected confidence 0.92, got %f", result.Confidence)
	}
}

func TestExtractTags_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"response": `{"tags": invalid}`,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	_, err := client.ExtractTags(context.Background(), "some content")

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestExtractTags_EmptyContent(t *testing.T) {
	client := ollama.NewClient("http://localhost", "test-model", "test-embed", 768)
	_, err := client.ExtractTags(context.Background(), "")

	if err == nil || err.Error() == "" {
		t.Error("expected error for empty content, got nil")
	}
}

func TestExtractTags_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	_, err := client.ExtractTags(context.Background(), "some content")

	if err == nil {
		t.Error("expected error for server error, got nil")
	}
}

func TestGenerateEmbedding_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embed" {
			t.Errorf("expected path /api/embed, got %s", r.URL.Path)
		}

		embedding := make([]float32, 768)
		for i := range embedding {
			embedding[i] = float32(i) / 768.0
		}
		resp := map[string][][]float32{
			"embeddings": {embedding},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	result, err := client.GenerateEmbedding(context.Background(), "some text")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 768 {
		t.Errorf("expected 768 dimensions, got %d", len(result))
	}
}

func TestGenerateEmbedding_WrongDimensions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		embedding := make([]float32, 384)
		resp := map[string][][]float32{
			"embeddings": {embedding},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	_, err := client.GenerateEmbedding(context.Background(), "some text")

	if err == nil {
		t.Error("expected error for wrong dimensions, got nil")
	}
}

func TestGenerateEmbedding_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	_, err := client.GenerateEmbedding(context.Background(), "some text")

	if err == nil {
		t.Error("expected error for server error, got nil")
	}
}

func TestHealthy_Up(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected path /api/tags, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	if !client.Healthy(context.Background()) {
		t.Error("expected Healthy to return true")
	}
}

func TestHealthy_Down(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	server.Close()

	// Use a short timeout for Healthy to avoid waiting 5s in test
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if client.Healthy(ctx) {
		t.Error("expected Healthy to return false for closed server")
	}
}

func TestCircuitBreaker_OpensAfterConsecutiveFailures(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	ctx := context.Background()

	// Trigger 3 consecutive failures to trip the breaker
	for range 3 {
		_, _ = client.ExtractTags(ctx, "content")
	}

	countAfterTripping := callCount

	// Next call should fail fast without hitting the server
	_, err := client.ExtractTags(ctx, "content")
	if err == nil {
		t.Fatal("expected error when circuit breaker is open")
	}

	if callCount != countAfterTripping {
		t.Errorf("expected no additional server calls after breaker opened, got %d extra", callCount-countAfterTripping)
	}
}

func TestCircuitBreaker_SharedBetweenMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	ctx := context.Background()

	// Trip breaker with ExtractTags failures
	for range 3 {
		_, _ = client.ExtractTags(ctx, "content")
	}

	// GenerateEmbedding should also fail fast
	_, err := client.GenerateEmbedding(ctx, "text")
	if err == nil {
		t.Fatal("expected GenerateEmbedding to fail when breaker is open from ExtractTags failures")
	}
}

func TestHealthy_FalseWhenBreakerOpen(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	ctx := context.Background()

	// Verify Healthy works before breaker trips
	if !client.Healthy(ctx) {
		t.Fatal("expected Healthy=true before breaker trips")
	}

	// Trip the breaker
	for range 3 {
		_, _ = client.ExtractTags(ctx, "content")
	}

	countBeforeHealthy := callCount

	// Healthy should return false without making an HTTP call
	if client.Healthy(ctx) {
		t.Error("expected Healthy=false when circuit breaker is open")
	}

	if callCount != countBeforeHealthy {
		t.Errorf("expected no HTTP call when breaker is open, got %d extra calls", callCount-countBeforeHealthy)
	}
}

func TestGenerateEmbedding_AllZerosRejected(t *testing.T) {
	embedding := make([]float32, 768)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string][][]float32{"embeddings": {embedding}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 768)
	_, err := client.GenerateEmbedding(context.Background(), "text")
	if err == nil {
		t.Fatal("expected error for all-zeros embedding")
	}
	wantSubstr := "embedding is all zeros"
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Errorf("error = %v, want containing %q", err, wantSubstr)
	}
}

func TestGenerateEmbedding_Normalization(t *testing.T) {
	embedding := []float32{3.0, 4.0}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string][][]float32{"embeddings": {embedding}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := ollama.NewClient(server.URL, "test-model", "test-embed", 2)
	result, err := client.GenerateEmbedding(context.Background(), "text")
	if err != nil {
		t.Fatal(err)
	}

	// Norm of [3, 4] is 5. Expected normalized: [0.6, 0.8]
	if result[0] != 0.6 || result[1] != 0.8 {
		t.Errorf("normalization failed, got %v", result)
	}
}
