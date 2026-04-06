package ollama

import "context"

// TagResult holds extracted tags with a confidence score.
type TagResult struct {
	Tags       []string
	Confidence float32
}

// Client defines the interface for Ollama API operations.
// This is the system boundary interface — mock this in tests.
type Client interface {
	ExtractTags(ctx context.Context, content string) (TagResult, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	Healthy(ctx context.Context) bool
}
