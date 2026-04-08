package storage

import (
	"context"
	"fmt"

	"github.com/neha037/mesh/internal/domain"
	pgvector "github.com/pgvector/pgvector-go"
)

type EdgeRepo struct {
	q *Queries
}

func NewEdgeRepo(q *Queries) *EdgeRepo {
	return &EdgeRepo{q: q}
}

func (r *EdgeRepo) BuildTagSharedEdges(ctx context.Context, nodeID string) error {
	uid, err := parseUUID(nodeID)
	if err != nil {
		return fmt.Errorf("parsing node id: %w", err)
	}
	if err := r.q.BuildTagSharedEdges(ctx, uid); err != nil {
		return fmt.Errorf("building tag-shared edges: %w", err)
	}
	return nil
}

func (r *EdgeRepo) UpsertSemanticEdge(ctx context.Context, sourceID, targetID string, weight float32) error {
	sid, err := parseUUID(sourceID)
	if err != nil {
		return fmt.Errorf("parsing source id: %w", err)
	}
	tid, err := parseUUID(targetID)
	if err != nil {
		return fmt.Errorf("parsing target id: %w", err)
	}
	if err := r.q.UpsertEdge(ctx, UpsertEdgeParams{
		SourceID: sid,
		TargetID: tid,
		RelType:  "semantic",
		Weight:   weight,
	}); err != nil {
		return fmt.Errorf("upserting semantic edge: %w", err)
	}
	return nil
}

func (r *EdgeRepo) FindSimilarNodes(ctx context.Context, embedding []float32, excludeID string, limit int32) ([]domain.SimilarNode, error) {
	uid, err := parseUUID(excludeID)
	if err != nil {
		return nil, fmt.Errorf("parsing exclude id: %w", err)
	}
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	rows, err := r.q.FindSimilarNodes(ctx, FindSimilarNodesParams{
		Embedding: pgvector.NewVector(embedding),
		ExcludeID: uid,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("finding similar nodes: %w", err)
	}
	nodes := make([]domain.SimilarNode, len(rows))
	for i, row := range rows {
		nodes[i] = domain.SimilarNode{
			ID:         uuidToString(row.ID),
			Title:      row.Title,
			Similarity: float32(row.Similarity),
		}
	}
	return nodes, nil
}
