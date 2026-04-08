package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neha037/mesh/internal/domain"
)

type mockJobRepo struct {
	claimFn    func(ctx context.Context) (*domain.Job, error)
	completeFn func(ctx context.Context, id string) error
	failFn     func(ctx context.Context, id string, errMsg string) error
	retryFn    func(ctx context.Context, id string, backoffSeconds int) error
}

func (m *mockJobRepo) CreateJob(_ context.Context, _ string, _ any, _ int32) (string, error) {
	return "new-job", nil
}

func (m *mockJobRepo) ClaimJob(ctx context.Context) (*domain.Job, error) {
	return m.claimFn(ctx)
}

func (m *mockJobRepo) CompleteJob(ctx context.Context, id string) error {
	if m.completeFn != nil {
		return m.completeFn(ctx, id)
	}
	return nil
}

func (m *mockJobRepo) FailJob(ctx context.Context, id, errMsg string) error {
	if m.failFn != nil {
		return m.failFn(ctx, id, errMsg)
	}
	return nil
}

func (m *mockJobRepo) RetryJob(ctx context.Context, id string, backoffSeconds int) error {
	if m.retryFn != nil {
		return m.retryFn(ctx, id, backoffSeconds)
	}
	return nil
}

type mockProcessor struct {
	processFn      func(ctx context.Context, job *domain.Job) error
	onDeadLetterFn func(ctx context.Context, job *domain.Job)
}

func (m *mockProcessor) Process(ctx context.Context, job *domain.Job) error {
	return m.processFn(ctx, job)
}

func (m *mockProcessor) OnDeadLetter(ctx context.Context, job *domain.Job) {
	if m.onDeadLetterFn != nil {
		m.onDeadLetterFn(ctx, job)
	}
}

func TestPool_ProcessesJob(t *testing.T) {
	var processed atomic.Int32

	payload, _ := json.Marshal(map[string]string{"node_id": "n1"})
	calls := atomic.Int32{}

	repo := &mockJobRepo{
		claimFn: func(_ context.Context) (*domain.Job, error) {
			if calls.Add(1) <= 1 {
				return &domain.Job{
					ID:          "job-1",
					Type:        "process_text",
					Payload:     payload,
					Status:      "running",
					Attempts:    1,
					MaxAttempts: 3,
				}, nil
			}
			return nil, nil // no more jobs
		},
	}

	proc := &mockProcessor{
		processFn: func(_ context.Context, job *domain.Job) error {
			processed.Add(1)
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewPool(repo, proc, 1, 5*time.Minute)
	go pool.Run(ctx)

	// Wait for job to be processed
	deadline := time.After(2 * time.Second)
	for {
		if processed.Load() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for job to be processed")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()
}

func TestPool_RetriesJobUnderMaxAttempts(t *testing.T) {
	var failCalled atomic.Int32
	var retryCalled atomic.Int32

	payload, _ := json.Marshal(map[string]string{"node_id": "n1"})
	calls := atomic.Int32{}

	repo := &mockJobRepo{
		claimFn: func(_ context.Context) (*domain.Job, error) {
			if calls.Add(1) <= 1 {
				return &domain.Job{
					ID:          "job-fail",
					Type:        "process_url",
					Payload:     payload,
					Status:      "running",
					Attempts:    1,
					MaxAttempts: 3,
				}, nil
			}
			return nil, nil
		},
		failFn: func(_ context.Context, _ string, _ string) error {
			failCalled.Add(1)
			return nil
		},
		retryFn: func(_ context.Context, _ string, _ int) error {
			retryCalled.Add(1)
			return nil
		},
	}

	proc := &mockProcessor{
		processFn: func(_ context.Context, _ *domain.Job) error {
			return errors.New("scrape failed")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewPool(repo, proc, 1, 5*time.Minute)
	go pool.Run(ctx)

	deadline := time.After(2 * time.Second)
	for {
		if retryCalled.Load() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out: retry=%d", retryCalled.Load())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()

	// RetryJob should be called (attempts < maxAttempts), not FailJob.
	if failCalled.Load() != 0 {
		t.Errorf("expected FailJob not called, got %d calls", failCalled.Load())
	}
}

func TestPool_DeadLettersJobAtMaxAttempts(t *testing.T) {
	var failCalled atomic.Int32
	var retryCalled atomic.Int32

	payload, _ := json.Marshal(map[string]string{"node_id": "n1"})
	calls := atomic.Int32{}

	repo := &mockJobRepo{
		claimFn: func(_ context.Context) (*domain.Job, error) {
			if calls.Add(1) <= 1 {
				return &domain.Job{
					ID:          "job-dead",
					Type:        "process_url",
					Payload:     payload,
					Status:      "running",
					Attempts:    3,
					MaxAttempts: 3,
				}, nil
			}
			return nil, nil
		},
		failFn: func(_ context.Context, _ string, _ string) error {
			failCalled.Add(1)
			return nil
		},
		retryFn: func(_ context.Context, _ string, _ int) error {
			retryCalled.Add(1)
			return nil
		},
	}

	proc := &mockProcessor{
		processFn: func(_ context.Context, _ *domain.Job) error {
			return errors.New("scrape failed")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewPool(repo, proc, 1, 5*time.Minute)
	go pool.Run(ctx)

	deadline := time.After(2 * time.Second)
	for {
		if failCalled.Load() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out: fail=%d", failCalled.Load())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	cancel()

	// FailJob should be called (attempts == maxAttempts), not RetryJob.
	if retryCalled.Load() != 0 {
		t.Errorf("expected RetryJob not called, got %d calls", retryCalled.Load())
	}
}

func TestPool_GracefulShutdown(t *testing.T) {
	repo := &mockJobRepo{
		claimFn: func(ctx context.Context) (*domain.Job, error) {
			return nil, nil // no jobs
		},
	}

	proc := &mockProcessor{
		processFn: func(_ context.Context, _ *domain.Job) error {
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		pool := NewPool(repo, proc, 2, 5*time.Minute)
		pool.Run(ctx)
		close(done)
	}()

	// Cancel quickly
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK — pool shut down
	case <-time.After(5 * time.Second):
		t.Fatal("pool did not shut down in time")
	}
}

func TestPool_FatalErrorSkipsRetry(t *testing.T) {
	var failCalled atomic.Int32
	var retryCalled atomic.Int32

	payload, _ := json.Marshal(map[string]string{"node_id": "n1"})
	calls := atomic.Int32{}

	repo := &mockJobRepo{
		claimFn: func(_ context.Context) (*domain.Job, error) {
			if calls.Add(1) <= 1 {
				return &domain.Job{
					ID:          "job-fatal",
					Type:        "process_url",
					Payload:     payload,
					Status:      "running",
					Attempts:    1,
					MaxAttempts: 3,
				}, nil
			}
			return nil, nil
		},
		failFn: func(_ context.Context, _ string, _ string) error {
			failCalled.Add(1)
			return nil
		},
		retryFn: func(_ context.Context, _ string, _ int) error {
			retryCalled.Add(1)
			return nil
		},
	}

	proc := &mockProcessor{
		processFn: func(_ context.Context, _ *domain.Job) error {
			return fmt.Errorf("%w: bad payload", domain.ErrFatal)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewPool(repo, proc, 1, 5*time.Minute)
	go pool.Run(ctx)

	deadline := time.After(2 * time.Second)
	for {
		if failCalled.Load() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for FailJob")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	cancel()

	if retryCalled.Load() != 0 {
		t.Errorf("expected no retries for fatal error, got %d", retryCalled.Load())
	}
}

func TestRetryBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		wantMin int
		wantMax int
	}{
		{1, 0, 0},
		{2, 24, 36},  // base=30, ±20%
		{3, 48, 72},  // base=60, ±20%
		{4, 96, 144}, // base=120, ±20%
	}
	for _, tt := range tests {
		got := retryBackoff(tt.attempt)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("retryBackoff(%d) = %d, want [%d, %d]", tt.attempt, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestPool_OnDeadLetterCalled(t *testing.T) {
	var deadLetterCalled atomic.Int32

	payload, _ := json.Marshal(map[string]string{"node_id": "n1"})
	calls := atomic.Int32{}

	repo := &mockJobRepo{
		claimFn: func(_ context.Context) (*domain.Job, error) {
			if calls.Add(1) <= 1 {
				return &domain.Job{
					ID:          "job-dl",
					Type:        "process_url",
					Payload:     payload,
					Status:      "running",
					Attempts:    3,
					MaxAttempts: 3,
				}, nil
			}
			return nil, nil
		},
	}

	proc := &mockProcessor{
		processFn: func(_ context.Context, _ *domain.Job) error {
			return errors.New("permanent failure")
		},
		onDeadLetterFn: func(_ context.Context, job *domain.Job) {
			deadLetterCalled.Add(1)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool := NewPool(repo, proc, 1, 5*time.Minute)
	go pool.Run(ctx)

	deadline := time.After(2 * time.Second)
	for {
		if deadLetterCalled.Load() >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for OnDeadLetter to be called")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	cancel()
}
