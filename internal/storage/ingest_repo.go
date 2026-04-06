package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neha037/mesh/internal/domain"
)

// IngestRepo handles transactional creation of nodes and their processing jobs.
type IngestRepo struct {
	pool *pgxpool.Pool
}

// NewIngestRepo returns an IngestRepo backed by the given connection pool.
func NewIngestRepo(pool *pgxpool.Pool) *IngestRepo {
	return &IngestRepo{pool: pool}
}

// IngestURL creates a pending node and enqueues a process_url job atomically.
func (r *IngestRepo) IngestURL(ctx context.Context, url, nodeType string) (domain.IngestURLResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.IngestURLResult{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is best-effort after commit

	q := New(tx)

	node, err := q.InsertPendingNode(ctx, InsertPendingNodeParams{
		Type:      nodeType,
		Title:     url, // title defaults to URL; worker updates after scraping
		SourceUrl: pgtype.Text{String: url, Valid: true},
	})
	if err != nil {
		return domain.IngestURLResult{}, fmt.Errorf("inserting node: %w", err)
	}

	nodeID := uuidToString(node.ID)
	payload, err := json.Marshal(map[string]string{
		"url":     url,
		"node_id": nodeID,
	})
	if err != nil {
		return domain.IngestURLResult{}, fmt.Errorf("marshaling payload: %w", err)
	}

	job, err := q.CreateJob(ctx, CreateJobParams{
		Type:        "process_url",
		Payload:     payload,
		MaxAttempts: 3,
	})
	if err != nil {
		return domain.IngestURLResult{}, fmt.Errorf("creating job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.IngestURLResult{}, fmt.Errorf("committing transaction: %w", err)
	}

	return domain.IngestURLResult{
		NodeID: nodeID,
		JobID:  uuidToString(job.ID),
	}, nil
}

// IngestText creates a pending node and enqueues a process_text job atomically.
func (r *IngestRepo) IngestText(ctx context.Context, title, content, nodeType string) (domain.IngestTextResult, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.IngestTextResult{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback is best-effort after commit

	q := New(tx)

	node, err := q.InsertPendingNode(ctx, InsertPendingNodeParams{
		Type:    nodeType,
		Title:   title,
		Content: pgtype.Text{String: content, Valid: content != ""},
	})
	if err != nil {
		return domain.IngestTextResult{}, fmt.Errorf("inserting node: %w", err)
	}

	nodeID := uuidToString(node.ID)
	payload, err := json.Marshal(map[string]string{
		"node_id": nodeID,
	})
	if err != nil {
		return domain.IngestTextResult{}, fmt.Errorf("marshaling payload: %w", err)
	}

	job, err := q.CreateJob(ctx, CreateJobParams{
		Type:        "process_text",
		Payload:     payload,
		MaxAttempts: 3,
	})
	if err != nil {
		return domain.IngestTextResult{}, fmt.Errorf("creating job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.IngestTextResult{}, fmt.Errorf("committing transaction: %w", err)
	}

	return domain.IngestTextResult{
		NodeID: nodeID,
		JobID:  uuidToString(job.ID),
	}, nil
}
