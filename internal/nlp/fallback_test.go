package nlp_test

import (
	"testing"

	"github.com/neha037/mesh/internal/nlp"
)

func TestExtractTags_ExtractsNouns(t *testing.T) {
	extractor := nlp.NewFallbackExtractor()
	content := "Machine learning and neural networks are transforming artificial intelligence research at Google and OpenAI."
	result, err := extractor.ExtractTags(content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) < 2 {
		t.Errorf("expected at least 2 tags, got %d", len(result.Tags))
	}
	if result.Confidence < 0.5 || result.Confidence > 0.7 {
		t.Errorf("expected confidence between 0.5 and 0.7, got %f", result.Confidence)
	}
}

func TestExtractTags_EmptyContent(t *testing.T) {
	extractor := nlp.NewFallbackExtractor()
	result, err := extractor.ExtractTags("")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) != 0 {
		t.Errorf("expected 0 tags for empty content, got %d", len(result.Tags))
	}
}

func TestExtractTags_ShortContent(t *testing.T) {
	extractor := nlp.NewFallbackExtractor()
	_, err := extractor.ExtractTags("Hello")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// At most 0 or some tags, but no error
}

func TestExtractTags_DeduplicatesTags(t *testing.T) {
	extractor := nlp.NewFallbackExtractor()
	content := "Golang is great. Golang is powerful. Golang."
	result, err := extractor.ExtractTags(content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	counts := make(map[string]int)
	for _, tag := range result.Tags {
		counts[tag]++
		if counts[tag] > 1 {
			t.Errorf("duplicate tag found: %s", tag)
		}
	}
}

func TestExtractTags_MaxEightTags(t *testing.T) {
	extractor := nlp.NewFallbackExtractor()
	content := "The quick brown fox jumps over the lazy dog. Programming is fun. Go is a language. Software engineering, artificial intelligence, machine learning, data science, deep learning, computer vision, natural language processing, robotics."
	result, err := extractor.ExtractTags(content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tags) > 8 {
		t.Errorf("expected at most 8 tags, got %d", len(result.Tags))
	}
}
