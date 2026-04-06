package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

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
	for i, row := range rows {
		nodes[i] = domain.Node{
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
	if params.CursorAt != nil {
		dbParams.CursorTime = pgtype.Timestamptz{Time: *params.CursorAt, Valid: true}
	}
	if params.CursorID != nil {
		var uuid pgtype.UUID
		if err := uuid.Scan(*params.CursorID); err != nil {
			return domain.ListNodesResult{}, fmt.Errorf("invalid cursor ID: %w", err)
		}
		dbParams.CursorID = uuid
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
	for i, row := range rows {
		nodes[i] = domain.Node{
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
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return fmt.Errorf("invalid node ID %q: %w", id, err)
	}

	tag, err := r.q.DeleteNodeReturningTag(ctx, uuid)
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
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return domain.Node{}, fmt.Errorf("invalid node ID %q: %w", id, err)
	}

	row, err := r.q.GetNode(ctx, uuid)
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

// uuidToString formats a pgtype.UUID as a standard string.
func uuidToString(u pgtype.UUID) string {
	return uuid.UUID(u.Bytes).String()
}
