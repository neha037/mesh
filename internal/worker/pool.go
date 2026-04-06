package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/neha037/mesh/internal/domain"
)

const (
	minBackoff = 1 * time.Second
	maxBackoff = 30 * time.Second
)

// Pool manages a fixed number of goroutines that claim and process jobs.
type Pool struct {
	jobs      domain.JobRepository
	processor Processor
	count     int
}

// NewPool creates a worker pool with the given concurrency.
func NewPool(jobs domain.JobRepository, processor Processor, count int) *Pool {
	return &Pool{
		jobs:      jobs,
		processor: processor,
		count:     count,
	}
}

// Run starts the worker goroutines and blocks until ctx is cancelled.
func (p *Pool) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := range p.count {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.loop(ctx, id)
		}(i)
	}
	slog.Info("worker pool started", "workers", p.count)
	wg.Wait()
	slog.Info("worker pool stopped")
}

func (p *Pool) loop(ctx context.Context, id int) {
	backoff := minBackoff

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		job, err := p.jobs.ClaimJob(ctx)
		if err != nil {
			slog.Error("claim failed", "worker", id, "error", err)
			if !sleep(ctx, backoff) {
				return
			}
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if job == nil {
			// No jobs available, back off.
			if !sleep(ctx, backoff) {
				return
			}
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		// Reset backoff on successful claim.
		backoff = minBackoff

		slog.Info("processing job", "worker", id, "job_id", job.ID, "type", job.Type)

		if err := p.processor.Process(ctx, job); err != nil {
			slog.Error("job failed", "worker", id, "job_id", job.ID, "error", err)
			if job.Attempts < job.MaxAttempts {
				retryDelay := retryBackoff(int(job.Attempts))
				if retryErr := p.jobs.RetryJob(ctx, job.ID, retryDelay); retryErr != nil {
					slog.Error("retrying job", "worker", id, "job_id", job.ID, "error", retryErr)
				} else {
					slog.Info("job scheduled for retry", "worker", id, "job_id", job.ID, "attempt", job.Attempts, "delay_s", retryDelay)
				}
			} else {
				if failErr := p.jobs.FailJob(ctx, job.ID, err.Error()); failErr != nil {
					slog.Error("marking job failed", "worker", id, "job_id", job.ID, "error", failErr)
				} else {
					slog.Warn("job dead-lettered", "worker", id, "job_id", job.ID, "attempts", job.Attempts)
				}
			}
			continue
		}

		if err := p.jobs.CompleteJob(ctx, job.ID); err != nil {
			slog.Error("completing job", "worker", id, "job_id", job.ID, "error", err)
		} else {
			slog.Info("job completed", "worker", id, "job_id", job.ID)
		}
	}
}

// retryBackoff returns the delay in seconds for the given attempt number.
func retryBackoff(attempt int) int {
	switch attempt {
	case 1:
		return 0
	case 2:
		return 60
	default:
		return 300
	}
}

// sleep waits for the duration or until ctx is cancelled. Returns false if cancelled.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
