package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	pgvector "github.com/pgvector/pgvector-go"

	"github.com/neha037/mesh/internal/domain"
)

// NodeRepo adapts sqlc-generated Queries to the domain.NodeRepository interface.
type NodeRepo struct {
	q *Queries
}

// NewNodeRepo returns a NodeRepo backed by the given Queries.
func NewNodeRepo(q *Queries) *NodeRepo {
	return &NodeRepo{q: q}
}

// UpsertRawNode inserts or updates a node keyed by source URL.
func (r *NodeRepo) UpsertRawNode(ctx context.Context, nodeType, title, content, sourceURL string) (domain.UpsertResult, error) {
	row, err := r.q.UpsertRawNode(ctx, UpsertRawNodeParams{
		Type:      nodeType,
		Title:     title,
		Content:   pgtype.Text{String: content, Valid: content != ""},
		SourceUrl: pgtype.Text{String: sourceURL, Valid: sourceURL != ""},
	})
	if err != nil {
		return domain.UpsertResult{}, fmt.Errorf("upserting node: %w", err)
	}

	return domain.UpsertResult{
		Node: domain.Node{
			ID:        uuidToString(row.ID),
			Type:      row.Type,
			Title:     row.Title,
			SourceURL: row.SourceUrl.String,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Created: row.Created,
	}, nil
}

// ListRecentNodes returns the most recently created nodes.
func (r *NodeRepo) ListRecentNodes(ctx context.Context, limit int32) ([]domain.Node, error) {
	rows, err := r.q.ListRecentNodes(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("listing recent nodes: %w", err)
	}

	nodes := make([]domain.Node, len(rows))
	for i := range rows {
		nodes[i] = domain.Node{
			ID:        uuidToString(rows[i].ID),
			Type:      rows[i].Type,
			Title:     rows[i].Title,
			Content:   rows[i].Content.String,
			Summary:   rows[i].Summary.String,
			SourceURL: rows[i].SourceUrl.String,
			ImageKey:  rows[i].ImageKey.String,
			Status:    rows[i].Status,
			Version:   rows[i].Version,
			CreatedAt: rows[i].CreatedAt.Time,
			UpdatedAt: rows[i].UpdatedAt.Time,
		}
	}
	return nodes, nil
}

// ListNodes returns a paginated list of nodes ordered by creation time.
func (r *NodeRepo) ListNodes(ctx context.Context, params domain.ListNodesParams) (domain.ListNodesResult, error) {
	// Fetch one extra to determine if there are more results.
	dbParams := ListNodesParams{
		Limit: params.Limit + 1,
	}
	if (params.CursorAt == nil) != (params.CursorID == nil) {
		return domain.ListNodesResult{}, fmt.Errorf("cursor_at and cursor_id must both be set or both be nil")
	}
	if params.CursorAt != nil {
		dbParams.CursorTime = pgtype.Timestamptz{Time: *params.CursorAt, Valid: true}
	}
	if params.CursorID != nil {
		cursorUUID, err := parseUUID(*params.CursorID)
		if err != nil {
			return domain.ListNodesResult{}, fmt.Errorf("invalid cursor ID: %w", err)
		}
		dbParams.CursorID = cursorUUID
	}

	rows, err := r.q.ListNodes(ctx, dbParams)
	if err != nil {
		return domain.ListNodesResult{}, fmt.Errorf("listing nodes: %w", err)
	}

	hasMore := len(rows) > int(params.Limit)
	if hasMore {
		rows = rows[:params.Limit]
	}

	nodes := make([]domain.Node, len(rows))
	for i := range rows {
		nodes[i] = domain.Node{
			ID:        uuidToString(rows[i].ID),
			Type:      rows[i].Type,
			Title:     rows[i].Title,
			Content:   rows[i].Content.String,
			Summary:   rows[i].Summary.String,
			SourceURL: rows[i].SourceUrl.String,
			ImageKey:  rows[i].ImageKey.String,
			Status:    rows[i].Status,
			Version:   rows[i].Version,
			CreatedAt: rows[i].CreatedAt.Time,
			UpdatedAt: rows[i].UpdatedAt.Time,
		}
	}

	return domain.ListNodesResult{
		Nodes:   nodes,
		HasMore: hasMore,
	}, nil
}

// DeleteNode removes a node by ID. Returns domain.ErrNotFound if the node
// does not exist.
func (r *NodeRepo) DeleteNode(ctx context.Context, id string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}

	tag, err := r.q.DeleteNodeReturningTag(ctx, uid)
	if err != nil {
		return fmt.Errorf("deleting node: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// GetNode retrieves a node by its ID.
func (r *NodeRepo) GetNode(ctx context.Context, id string) (domain.Node, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.Node{}, err
	}

	row, err := r.q.GetNode(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Node{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Node{}, fmt.Errorf("getting node: %w", err)
	}

	return domain.Node{
		ID:        uuidToString(row.ID),
		Type:      row.Type,
		Title:     row.Title,
		Content:   row.Content.String,
		Summary:   row.Summary.String,
		SourceURL: row.SourceUrl.String,
		ImageKey:  row.ImageKey.String,
		Status:    row.Status,
		Version:   row.Version,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// UpdateNodeContent updates a node's content and sets status to processed.
func (r *NodeRepo) UpdateNodeContent(ctx context.Context, id, content string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	if err := r.q.UpdateNodeContent(ctx, UpdateNodeContentParams{
		ID:      uid,
		Content: pgtype.Text{String: content, Valid: content != ""},
	}); err != nil {
		return fmt.Errorf("updating node content: %w", err)
	}
	return nil
}

// UpdateNodeStatus updates a node's status.
func (r *NodeRepo) UpdateNodeStatus(ctx context.Context, id, status string) error {
	uid, err := parseUUID(id)
	if err != nil {
		return err
	}
	if err := r.q.UpdateNodeStatus(ctx, UpdateNodeStatusParams{
		ID:     uid,
		Status: status,
	}); err != nil {
		return fmt.Errorf("updating node status: %w", err)
	}
	return nil
}

func (r *NodeRepo) UpdateNodeEmbedding(ctx context.Context, id string, embedding []float32, expectedVersion int32) (bool, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return false, fmt.Errorf("parsing node id: %w", err)
	}
	result, err := r.q.UpdateNodeEmbedding(ctx, UpdateNodeEmbeddingParams{
		ID:        uid,
		Embedding: pgvector.NewVector(embedding),
		Version:   expectedVersion,
	})
	if err != nil {
		return false, fmt.Errorf("updating embedding: %w", err)
	}
	return result.RowsAffected() > 0, nil
}

func (r *NodeRepo) GetNodeContent(ctx context.Context, id string) (domain.Node, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return domain.Node{}, fmt.Errorf("parsing node id: %w", err)
	}
	row, err := r.q.GetNodeContent(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Node{}, domain.ErrNotFound
		}
		return domain.Node{}, fmt.Errorf("getting node content: %w", err)
	}
	return domain.Node{
		ID:      uuidToString(row.ID),
		Type:    row.Type,
		Title:   row.Title,
		Content: row.Content.String,
		Status:  row.Status,
		Version: row.Version,
	}, nil
}

func (r *NodeRepo) GetNodeEmbedding(ctx context.Context, id string) ([]float32, error) {
	uid, err := parseUUID(id)
	if err != nil {
		return nil, fmt.Errorf("parsing node id: %w", err)
	}
	embedding, err := r.q.GetNodeEmbedding(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("getting node embedding: %w", err)
	}
	return embedding.Slice(), nil
}

func (r *NodeRepo) ResetStaleProcessingNodes(ctx context.Context, cutoff time.Time) (int64, error) {
	return r.q.ResetStaleProcessingNodes(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
}

func (r *NodeRepo) ListNodesWithoutEmbedding(ctx context.Context, limit int32) ([]string, error) {
	rows, err := r.q.ListNodesWithoutEmbedding(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("listing nodes without embedding: %w", err)
	}
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = uuidToString(row)
	}
	return ids, nil
}
