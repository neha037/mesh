package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/neha037/mesh/internal/domain"
)

// JobRepo adapts sqlc-generated Queries to the domain.JobRepository interface.
type JobRepo struct {
	q *Queries
}

// NewJobRepo returns a JobRepo backed by the given Queries.
func NewJobRepo(q *Queries) *JobRepo {
	return &JobRepo{q: q}
}

// ClaimJob atomically claims the next pending job using FOR UPDATE SKIP LOCKED.
// Returns nil, nil when no jobs are available.
func (r *JobRepo) ClaimJob(ctx context.Context) (*domain.Job, error) {
	row, err := r.q.ClaimJob(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claiming job: %w", err)
	}
	return claimRowToJob(&row), nil
}

// CompleteJob marks a job as done.
func (r *JobRepo) CompleteJob(ctx context.Context, id string) error {
	uuid, err := parseUUID(id)
	if err != nil {
		return err
	}
	if err := r.q.CompleteJob(ctx, uuid); err != nil {
		return fmt.Errorf("completing job %s: %w", id, err)
	}
	return nil
}

// FailJob marks a job as failed (or dead if max attempts reached).
func (r *JobRepo) FailJob(ctx context.Context, id, errMsg string) error {
	uuid, err := parseUUID(id)
	if err != nil {
		return err
	}
	if err := r.q.FailJob(ctx, FailJobParams{
		ID:    uuid,
		Error: pgtype.Text{String: errMsg, Valid: errMsg != ""},
	}); err != nil {
		return fmt.Errorf("failing job %s: %w", id, err)
	}
	return nil
}

// RetryJob resets a failed job to pending with a backoff delay.
func (r *JobRepo) RetryJob(ctx context.Context, id string, backoffSeconds int) error {
	uuid, err := parseUUID(id)
	if err != nil {
		return err
	}
	if err := r.q.RetryJob(ctx, RetryJobParams{
		ID:             uuid,
		BackoffSeconds: float64(backoffSeconds),
	}); err != nil {
		return fmt.Errorf("retrying job %s: %w", id, err)
	}
	return nil
}

func claimRowToJob(row *ClaimJobRow) *domain.Job {
	j := &domain.Job{
		ID:           uuidToString(row.ID),
		Type:         row.Type,
		Payload:      row.Payload,
		Status:       row.Status,
		Attempts:     row.Attempts,
		MaxAttempts:  row.MaxAttempts,
		Error:        row.Error.String,
		ScheduledFor: row.ScheduledFor.Time,
		CreatedAt:    row.CreatedAt.Time,
	}
	if row.ClaimedAt.Valid {
		t := row.ClaimedAt.Time
		j.ClaimedAt = &t
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		j.CompletedAt = &t
	}
	return j
}

func parseUUID(id string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid ID %q: %w", id, err)
	}
	return uuid, nil
}
