package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/neha037/mesh/internal/domain"
	"github.com/neha037/mesh/internal/nlp"
	"github.com/neha037/mesh/internal/scraper"

	"github.com/microcosm-cc/bluemonday"
)

var htmlPolicy = bluemonday.UGCPolicy()

// Processor handles the execution of a claimed job.
type Processor interface {
	Process(ctx context.Context, job *domain.Job) error
	OnDeadLetter(ctx context.Context, job *domain.Job)
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
	case "reembed_batch":
		return p.reembedBatch(ctx, job)
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
		return fmt.Errorf("%w: unmarshaling payload: %v", domain.ErrFatal, err)
	}

	if err := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "processing"); err != nil {
		return fmt.Errorf("setting node processing: %w", err)
	}

	result, err := p.scraper.Scrape(ctx, payload.URL)
	if err != nil {
		// Return error to trigger job retry via pool.loop()
		// Node stays in "processing" — will be marked "failed" only when dead-lettered
		return fmt.Errorf("scraping URL: %w", err)
	}

	result.Title = htmlPolicy.Sanitize(result.Title)
	result.Content = htmlPolicy.Sanitize(result.Content)

	if err := p.nodes.UpdateNodeContent(ctx, payload.NodeID, result.Content); err != nil {
		return fmt.Errorf("updating node content: %w", err)
	}

	// Queue next step: process_text with 5 max attempts
	_, err = p.jobs.CreateJob(ctx, "process_text", map[string]string{"node_id": payload.NodeID}, 5)
	return err
}

type textPayload struct {
	NodeID string `json:"node_id"`
}

func (p *DefaultProcessor) processText(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("%w: unmarshaling payload: %v", domain.ErrFatal, err)
	}

	node, err := p.nodes.GetNodeContent(ctx, payload.NodeID)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}

	res, err := p.nlp.ProcessContent(ctx, node.Content)
	if err != nil {
		return fmt.Errorf("nlp processing: %w", err)
	}

	var tagIDs []string
	for _, tagName := range res.Tags {
		tagID, err := p.tags.UpsertTag(ctx, tagName)
		if err != nil {
			slog.Error("failed to upsert tag", "tag", tagName, "error", err)
			continue
		}
		tagIDs = append(tagIDs, tagID)
	}
	if len(tagIDs) > 0 {
		if err := p.tags.BulkAssociateNodeTags(ctx, payload.NodeID, tagIDs, res.Confidence); err != nil {
			slog.Error("failed to bulk associate tags", "node", payload.NodeID, "error", err)
		}
	}

	// Enqueue embedding generation with 5 max attempts
	_, err = p.jobs.CreateJob(ctx, "generate_embedding", map[string]string{"node_id": payload.NodeID}, 5)
	return err
}

func (p *DefaultProcessor) generateEmbedding(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("%w: unmarshaling payload: %v", domain.ErrFatal, err)
	}

	node, err := p.nodes.GetNodeContent(ctx, payload.NodeID)
	if err != nil {
		return fmt.Errorf("getting node: %w", err)
	}

	embedding, err := p.nlp.GenerateEmbedding(ctx, node.Content)
	if err != nil {
		return fmt.Errorf("generating embedding: %w", err)
	}

	if len(embedding) > 0 {
		ok, err := p.nodes.UpdateNodeEmbedding(ctx, payload.NodeID, embedding, node.Version)
		if err != nil {
			return fmt.Errorf("updating embedding: %w", err)
		}
		if !ok {
			return fmt.Errorf("embedding update failed (stale version)")
		}
	}

	// Enqueue edge building with 3 max attempts
	_, err = p.jobs.CreateJob(ctx, "build_edges", map[string]string{"node_id": payload.NodeID}, 3)
	return err
}

func (p *DefaultProcessor) buildEdges(ctx context.Context, job *domain.Job) error {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("%w: unmarshaling payload: %v", domain.ErrFatal, err)
	}

	// 1. Semantic edges - read stored embedding from DB
	embedding, err := p.nodes.GetNodeEmbedding(ctx, payload.NodeID)
	if err != nil {
		slog.Warn("could not retrieve embedding for semantic edges", "node_id", payload.NodeID, "error", err)
	}
	if len(embedding) > 0 {
		similar, err := p.edges.FindSimilarNodes(ctx, embedding, payload.NodeID, 5)
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

func (p *DefaultProcessor) reembedBatch(ctx context.Context, _ *domain.Job) error {
	nodeIDs, err := p.nodes.ListNodesWithoutEmbedding(ctx, 100)
	if err != nil {
		return fmt.Errorf("listing nodes without embedding: %w", err)
	}

	for _, nodeID := range nodeIDs {
		if _, err := p.jobs.CreateJob(ctx, "generate_embedding", map[string]string{"node_id": nodeID}, 5); err != nil {
			slog.Error("failed to create reembed job", "node_id", nodeID, "error", err)
		}
	}

	slog.Info("reembed_batch complete", "enqueued", len(nodeIDs))
	return nil
}

// markNodeFailedIfApplicable sets node status to "failed" for node-related jobs.
func (p *DefaultProcessor) markNodeFailedIfApplicable(ctx context.Context, job *domain.Job) {
	var payload textPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return
	}
	if payload.NodeID == "" {
		var up urlPayload
		if err := json.Unmarshal(job.Payload, &up); err != nil {
			return
		}
		payload.NodeID = up.NodeID
	}
	if payload.NodeID != "" {
		if err := p.nodes.UpdateNodeStatus(ctx, payload.NodeID, "failed"); err != nil {
			slog.Error("failed to mark node as failed after dead-letter", "node_id", payload.NodeID, "error", err)
		}
	}
}

// OnDeadLetter is called when a job exhausts all retry attempts.
func (p *DefaultProcessor) OnDeadLetter(ctx context.Context, job *domain.Job) {
	p.markNodeFailedIfApplicable(ctx, job)
}
