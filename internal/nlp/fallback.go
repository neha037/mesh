package nlp

import (
	"log/slog"
	"strings"

	"github.com/jdkato/prose/v2"
	"github.com/neha037/mesh/internal/ollama"
)

type FallbackExtractor struct{}

func NewFallbackExtractor() *FallbackExtractor {
	return &FallbackExtractor{}
}

func (f *FallbackExtractor) ExtractTags(content string) (ollama.TagResult, error) {
	if content == "" {
		return ollama.TagResult{}, nil
	}

	doc, err := prose.NewDocument(content)
	if err != nil {
		return ollama.TagResult{}, err
	}

	tagMap := make(map[string]struct{})

	// 1. Collect named entities
	for _, entity := range doc.Entities() {
		tag := strings.ToLower(entity.Text)
		if f.isValidTag(tag) {
			tagMap[tag] = struct{}{}
		}
	}

	// 2. Collect nouns from POS tags
	for _, token := range doc.Tokens() {
		if strings.HasPrefix(token.Tag, "NN") {
			tag := strings.ToLower(token.Text)
			if f.isValidTag(tag) {
				tagMap[tag] = struct{}{}
			}
		}
	}

	// 3. Filter and limit
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
		if len(tags) >= 8 {
			break
		}
	}

	slog.Info("Fallback extraction complete", "tag_count", len(tags))

	return ollama.TagResult{
		Tags:       tags,
		Confidence: 0.6,
	}, nil
}

func (f *FallbackExtractor) isValidTag(tag string) bool {
	return len(tag) >= 2 && len(tag) <= 50
}
