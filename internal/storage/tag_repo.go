package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/neha037/mesh/internal/domain"
)

type TagRepo struct {
	q *Queries
}

func NewTagRepo(q *Queries) *TagRepo {
	return &TagRepo{q: q}
}

func (r *TagRepo) UpsertTag(ctx context.Context, name string) (string, error) {
	row, err := r.q.UpsertTag(ctx, name)
	if err != nil {
		return "", fmt.Errorf("upserting tag: %w", err)
	}
	return uuidToString(row.ID), nil
}

func (r *TagRepo) AssociateNodeTag(ctx context.Context, nodeID, tagID string, confidence float32) error {
	nid, err := parseUUID(nodeID)
	if err != nil {
		return fmt.Errorf("parsing node id: %w", err)
	}
	tid, err := parseUUID(tagID)
	if err != nil {
		return fmt.Errorf("parsing tag id: %w", err)
	}
	if err := r.q.AssociateNodeTag(ctx, AssociateNodeTagParams{
		NodeID:     nid,
		TagID:      tid,
		Confidence: pgtype.Float4{Float32: confidence, Valid: true},
	}); err != nil {
		return fmt.Errorf("associating tag: %w", err)
	}
	return nil
}

func (r *TagRepo) GetNodeTags(ctx context.Context, nodeID string) ([]domain.Tag, error) {
	nid, err := parseUUID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("parsing node id: %w", err)
	}
	rows, err := r.q.GetNodeTags(ctx, nid)
	if err != nil {
		return nil, fmt.Errorf("getting node tags: %w", err)
	}
	tags := make([]domain.Tag, len(rows))
	for i, row := range rows {
		tags[i] = domain.Tag{
			ID:         uuidToString(row.ID),
			Name:       row.Name,
			Confidence: row.Confidence.Float32,
		}
	}
	return tags, nil
}
