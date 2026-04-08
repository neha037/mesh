package nlp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neha037/mesh/internal/ollama"
)

type ProcessResult struct {
	Tags       []string
	Confidence float32
	Embedding  []float32
}

type Service struct {
	ollama   ollama.Client
	fallback *FallbackExtractor
}

func NewService(ollamaClient ollama.Client) *Service {
	return &Service{
		ollama:   ollamaClient,
		fallback: NewFallbackExtractor(),
	}
}

func (s *Service) ProcessContent(ctx context.Context, content string) (ProcessResult, error) {
	if content == "" {
		return ProcessResult{}, nil
	}

	var result ProcessResult

	if s.ollama.Healthy(ctx) {
		tagResult, err := s.ollama.ExtractTags(ctx, content)
		if err != nil {
			slog.Warn("ollama tag extraction failed, using fallback", "error", err)
			fallbackResult, _ := s.fallback.ExtractTags(content)
			result.Tags = fallbackResult.Tags
			result.Confidence = fallbackResult.Confidence
		} else {
			result.Tags = tagResult.Tags
			result.Confidence = tagResult.Confidence
		}

		embedding, err := s.ollama.GenerateEmbedding(ctx, content)
		if err != nil {
			slog.Warn("embedding generation failed", "error", err)
		} else {
			result.Embedding = embedding
		}
	} else {
		slog.Info("ollama unavailable, using fallback NLP")
		fallbackResult, _ := s.fallback.ExtractTags(content)
		result.Tags = fallbackResult.Tags
		result.Confidence = fallbackResult.Confidence
	}

	return result, nil
}

// GenerateEmbedding generates only an embedding vector for the given content.
func (s *Service) GenerateEmbedding(ctx context.Context, content string) ([]float32, error) {
	if content == "" {
		return nil, nil
	}
	if !s.ollama.Healthy(ctx) {
		return nil, fmt.Errorf("ollama unavailable for embedding generation")
	}
	return s.ollama.GenerateEmbedding(ctx, content)
}
