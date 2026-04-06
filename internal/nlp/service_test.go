package nlp_test

import (
	"context"
	"testing"

	"github.com/neha037/mesh/internal/nlp"
	"github.com/neha037/mesh/internal/ollama"
)

type mockOllamaClient struct {
	extractFn func(ctx context.Context, content string) (ollama.TagResult, error)
	embedFn   func(ctx context.Context, text string) ([]float32, error)
	healthyFn func(ctx context.Context) bool
}

func (m *mockOllamaClient) ExtractTags(ctx context.Context, content string) (ollama.TagResult, error) {
	return m.extractFn(ctx, content)
}
func (m *mockOllamaClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return m.embedFn(ctx, text)
}
func (m *mockOllamaClient) Healthy(ctx context.Context) bool {
	return m.healthyFn(ctx)
}

func TestProcessContent_OllamaAvailable(t *testing.T) {
	mockOllama := &mockOllamaClient{
		healthyFn: func(_ context.Context) bool { return true },
		extractFn: func(_ context.Context, _ string) (ollama.TagResult, error) {
			return ollama.TagResult{Tags: []string{"tag1", "tag2", "tag3"}, Confidence: 0.92}, nil
		},
		embedFn: func(_ context.Context, _ string) ([]float32, error) {
			return make([]float32, 768), nil
		},
	}

	svc := nlp.NewService(mockOllama)
	result, err := svc.ProcessContent(context.Background(), "some content")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(result.Tags))
	}
	if result.Confidence != 0.92 {
		t.Errorf("expected confidence 0.92, got %f", result.Confidence)
	}
	if len(result.Embedding) != 768 {
		t.Errorf("expected 768 dimensions, got %d", len(result.Embedding))
	}
}

func TestProcessContent_OllamaDown_Fallback(t *testing.T) {
	mockOllama := &mockOllamaClient{
		healthyFn: func(_ context.Context) bool { return false },
	}

	svc := nlp.NewService(mockOllama)
	result, err := svc.ProcessContent(context.Background(), "Machine learning is great.")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) == 0 {
		t.Error("expected tags from fallback, got none")
	}
	if result.Confidence < 0.5 || result.Confidence > 0.7 {
		t.Errorf("expected fallback confidence between 0.5 and 0.7, got %f", result.Confidence)
	}
	if result.Embedding != nil {
		t.Error("expected nil embedding when Ollama is down")
	}
}

func TestProcessContent_OllamaTagsFail_FallbackUsed(t *testing.T) {
	mockOllama := &mockOllamaClient{
		healthyFn: func(_ context.Context) bool { return true },
		extractFn: func(_ context.Context, _ string) (ollama.TagResult, error) {
			return ollama.TagResult{}, context.DeadlineExceeded
		},
		embedFn: func(_ context.Context, _ string) ([]float32, error) {
			return make([]float32, 768), nil
		},
	}

	svc := nlp.NewService(mockOllama)
	result, err := svc.ProcessContent(context.Background(), "Artificial intelligence.")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) == 0 {
		t.Error("expected tags from fallback after Ollama error")
	}
	if len(result.Embedding) != 768 {
		t.Errorf("expected embedding to be generated even if tags failed")
	}
}

func TestProcessContent_EmptyContent(t *testing.T) {
	svc := nlp.NewService(&mockOllamaClient{})
	result, err := svc.ProcessContent(context.Background(), "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) != 0 {
		t.Error("expected zero tags for empty content")
	}
}
