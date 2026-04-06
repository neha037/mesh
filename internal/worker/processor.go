package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/neha037/mesh/internal/domain"
	"github.com/neha037/mesh/internal/scraper"
)

// Processor handles the execution of a claimed job.
type Processor interface {
	Process(ctx context.Context, job *domain.Job) error
}

// DefaultProcessor routes jobs by type and executes them.
type DefaultProcessor struct {
	scraper *scraper.Service
	nodes   domain.NodeRepository
}

// NewProcessor creates a DefaultProcessor with the given dependencies.
func NewProcessor(s *scraper.Service, nodes domain.NodeRepository) *DefaultProcessor {
	return &DefaultProcessor{scraper: s, nodes: nodes}
}

// Process dispatches the job to the appropriate handler based on job type.
func (p *DefaultProcessor) Process(ctx context.Context, job *domain.Job) error {
	switch job.Type {
	case "process_url":
		return p.processURL(ctx, job)
	case "process_text":
		return p.processText(ctx, job)
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

type urlPayload struct {
	URL    string `json:"url"`
	NodeID string `json:"node_id"`
}

func (p *DefaultProcessor) processURL(ctx context.Context, job *domain.Job) error {
	var payload urlPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshaling payload: %w", err)
	}

	if err := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "processing"); err != nil {
		return fmt.Errorf("setting node processing: %w", err)
	}

	result, err := p.scraper.Scrape(ctx, payload.URL)
	if err != nil {
		if statusErr := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "failed"); statusErr != nil {
			slog.Error("failed to mark node as failed", "node_id", payload.NodeID, "error", statusErr)
		}
		return fmt.Errorf("scraping URL: %w", err)
	}

	if err := p.nodes.UpdateNodeContent(ctx, payload.NodeID, result.Content); err != nil {
		return fmt.Errorf("updating node content: %w", err)
	}

	return nil
}

type textPayload struct {
	NodeID string `json:"node_id"`
}

// TODO(phase2): add NLP tagging and embedding generation before marking as processed.
func (p *DefaultProcessor) processText(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshaling payload: %w", err)
	}

	if err := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "processed"); err != nil {
		return fmt.Errorf("updating node status: %w", err)
	}

	return nil
}
