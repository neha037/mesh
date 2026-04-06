package domain

import (
	"context"
	"encoding/json"
	"time"
)

// Job represents a background processing task.
type Job struct {
	ID           string
	Type         string
	Payload      json.RawMessage
	Status       string
	Attempts     int32
	MaxAttempts  int32
	Error        string
	CreatedAt    time.Time
	ClaimedAt    *time.Time
	CompletedAt  *time.Time
	ScheduledFor time.Time
}

// IngestURLResult holds the result of enqueueing a URL for processing.
type IngestURLResult struct {
	NodeID string
	JobID  string
}

// IngestTextResult holds the result of enqueueing text for processing.
type IngestTextResult struct {
	NodeID string
	JobID  string
}

// JobRepository defines the interface for job storage operations.
type JobRepository interface {
	CreateJob(ctx context.Context, jobType string, payload any, maxAttempts int32) (string, error)
	ClaimJob(ctx context.Context) (*Job, error)
	CompleteJob(ctx context.Context, id string) error
	FailJob(ctx context.Context, id string, errMsg string) error
	RetryJob(ctx context.Context, id string, backoffSeconds int) error
}

// IngestService handles the transactional creation of nodes and their
// associated processing jobs.
type IngestService interface {
	IngestURL(ctx context.Context, url, nodeType string) (IngestURLResult, error)
	IngestText(ctx context.Context, title, content, nodeType string) (IngestTextResult, error)
}
