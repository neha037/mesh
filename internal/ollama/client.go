package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker/v2"
)

type HTTPClient struct {
	baseURL    string
	tagModel   string
	embedModel string
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker[any]
}

func NewClient(baseURL, tagModel, embedModel string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		tagModel:   tagModel,
		embedModel: embedModel,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		breaker: gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
			Name:        "ollama",
			MaxRequests: 1,
			Timeout:     60 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 3
			},
		}),
	}
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format"`
}

type generateResponse struct {
	Response string `json:"response"`
}

type tagResponse struct {
	Tags       []string `json:"tags"`
	Confidence float32  `json:"confidence"`
}

func (c *HTTPClient) ExtractTags(ctx context.Context, content string) (TagResult, error) {
	if content == "" {
		return TagResult{}, fmt.Errorf("ollama: extracting tags: content is empty")
	}

	result, err := c.breaker.Execute(func() (any, error) {
		return c.extractTags(ctx, content)
	})
	if err != nil {
		return TagResult{}, fmt.Errorf("ollama: extracting tags: %w", err)
	}
	return result.(TagResult), nil
}

func (c *HTTPClient) extractTags(ctx context.Context, content string) (TagResult, error) {
	prompt := "Extract 3-8 key domain-specific concept tags from the following content. Return JSON: {\"tags\": [\"tag1\", \"tag2\"], \"confidence\": 0.0-1.0}. Tags must be lowercase, 1-3 words each. Avoid generic words like \"article\" or \"content\".\n\nContent:\n" + truncate(content, 4000)

	reqBody := generateRequest{
		Model:  c.tagModel,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return TagResult{}, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return TagResult{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TagResult{}, fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TagResult{}, fmt.Errorf("API error status %d", resp.StatusCode)
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return TagResult{}, fmt.Errorf("decoding response: %w", err)
	}

	var tResp tagResponse
	if err := json.Unmarshal([]byte(genResp.Response), &tResp); err != nil {
		return TagResult{}, fmt.Errorf("parsing nested json: %w", err)
	}

	if len(tResp.Tags) == 0 {
		return TagResult{}, fmt.Errorf("no tags identified")
	}

	return TagResult{
		Tags:       tResp.Tags,
		Confidence: tResp.Confidence,
	}, nil
}

type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (c *HTTPClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("ollama: generating embedding: input is empty")
	}

	result, err := c.breaker.Execute(func() (any, error) {
		return c.generateEmbedding(ctx, text)
	})
	if err != nil {
		return nil, fmt.Errorf("ollama: generating embedding: %w", err)
	}
	return result.([]float32), nil
}

func (c *HTTPClient) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	reqBody := embedRequest{
		Model: c.embedModel,
		Input: truncate(text, 8000),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/embed", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error status %d", resp.StatusCode)
	}

	var eResp embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&eResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(eResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	embedding := eResp.Embeddings[0]
	if len(embedding) != 768 {
		return nil, fmt.Errorf("expected 768 dimensions, got %d", len(embedding))
	}

	return embedding, nil
}

func (c *HTTPClient) Healthy(ctx context.Context) bool {
	if c.breaker.State() == gobreaker.StateOpen {
		return false
	}

	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(tctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == http.StatusOK
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
