package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/neha037/mesh/internal/domain"
	"github.com/neha037/mesh/internal/nlp"
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
	tags    domain.TagRepository
	edges   domain.EdgeRepository
	jobs    domain.JobRepository
	nlp     *nlp.Service
}

// NewProcessor creates a DefaultProcessor with the given dependencies.
func NewProcessor(
	s *scraper.Service,
	nodes domain.NodeRepository,
	tags domain.TagRepository,
	edges domain.EdgeRepository,
	jobs domain.JobRepository,
	nlpService *nlp.Service,
) *DefaultProcessor {
	return &DefaultProcessor{
		scraper: s,
		nodes:   nodes,
		tags:    tags,
		edges:   edges,
		jobs:    jobs,
		nlp:     nlpService,
	}
}

// Process dispatches the job to the appropriate handler based on job type.
func (p *DefaultProcessor) Process(ctx context.Context, job *domain.Job) error {
	switch job.Type {
	case "process_url":
		return p.processURL(ctx, job)
	case "process_text":
		return p.processText(ctx, job)
	case "generate_embedding":
		return p.generateEmbedding(ctx, job)
	case "build_edges":
		return p.buildEdges(ctx, job)
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

	// Queue next step: process_text
	_, err = p.jobs.CreateJob(ctx, "process_text", map[string]string{"node_id": payload.NodeID}, 3)
	return err
}

type textPayload struct {
	NodeID string `json:"node_id"`
}

func (p *DefaultProcessor) processText(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshaling payload: %w", err)
	}

	node, err := p.nodes.GetNodeContent(ctx, payload.NodeID)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}

	res, err := p.nlp.ProcessContent(ctx, node.Content)
	if err != nil {
		return fmt.Errorf("nlp processing: %w", err)
	}

	for _, tagName := range res.Tags {
		tagID, err := p.tags.UpsertTag(ctx, tagName)
		if err != nil {
			slog.Error("failed to upsert tag", "tag", tagName, "error", err)
			continue
		}
		if err := p.tags.AssociateNodeTag(ctx, payload.NodeID, tagID, res.Confidence); err != nil {
			slog.Error("failed to associate tag", "node", payload.NodeID, "tag", tagName, "error", err)
		}
	}

	// Enqueue embedding generation
	_, err = p.jobs.CreateJob(ctx, "generate_embedding", map[string]string{"node_id": payload.NodeID}, 3)
	return err
}

func (p *DefaultProcessor) generateEmbedding(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshaling payload: %w", err)
	}

	node, err := p.nodes.GetNodeContent(ctx, payload.NodeID)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}

	res, err := p.nlp.ProcessContent(ctx, node.Content)
	if err != nil {
		return fmt.Errorf("generating embedding: %w", err)
	}

	if len(res.Embedding) > 0 {
		ok, err := p.nodes.UpdateNodeEmbedding(ctx, payload.NodeID, res.Embedding, node.Version)
		if err != nil {
			return fmt.Errorf("updating embedding: %w", err)
		}
		if !ok {
			return fmt.Errorf("embedding update failed (stale version)")
		}
	}

	// Enqueue edge building
	_, err = p.jobs.CreateJob(ctx, "build_edges", map[string]string{"node_id": payload.NodeID}, 2)
	return err
}

func (p *DefaultProcessor) buildEdges(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshaling payload: %w", err)
	}

	// 1. Semantic edges using pgvector
	node, err := p.nodes.GetNodeContent(ctx, payload.NodeID)
	if err != nil {
		return fmt.Errorf("getting node for edges: %w", err)
	}

	res, err := p.nlp.ProcessContent(ctx, node.Content)
	if err == nil && len(res.Embedding) > 0 {
		similar, err := p.edges.FindSimilarNodes(ctx, res.Embedding, payload.NodeID, 5)
		if err != nil {
			slog.Error("failed to find similar nodes", "error", err)
		} else {
			for _, sim := range similar {
				if err := p.edges.UpsertSemanticEdge(ctx, payload.NodeID, sim.ID, sim.Similarity); err != nil {
					slog.Error("failed to create semantic edge", "error", err)
				}
			}
		}
	}

	// 2. Tag-shared edges
	if err := p.edges.BuildTagSharedEdges(ctx, payload.NodeID); err != nil {
		return fmt.Errorf("building tag edges: %w", err)
	}

	// Final status update - node is now fully processed
	return p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "processed")
}
