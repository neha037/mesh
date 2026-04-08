package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neha037/mesh/internal/config"
	"github.com/neha037/mesh/internal/nlp"
	"github.com/neha037/mesh/internal/ollama"
	"github.com/neha037/mesh/internal/scraper"
	"github.com/neha037/mesh/internal/storage"
	"github.com/neha037/mesh/internal/worker"
	"github.com/neha037/mesh/migrations"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))

	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", cfg.LogLevel, err)
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if err := storage.RunMigrations(migrations.FS, cfg.DatabaseURL); err != nil {
		return err
	}
	slog.Info("migrations applied")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := storage.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	queries := storage.New(pool)
	jobRepo := storage.NewJobRepo(queries)
	nodeRepo := storage.NewNodeRepo(queries)
	tagRepo := storage.NewTagRepo(queries)
	edgeRepo := storage.NewEdgeRepo(queries)

	ollamaClient := ollama.NewClient(cfg.OllamaHost, cfg.OllamaModel, cfg.EmbeddingModel, cfg.EmbeddingDim)
	if ollamaClient.Healthy(ctx) {
		slog.Info("ollama connection verified", "host", cfg.OllamaHost)
	} else {
		slog.Warn("ollama not reachable at startup — workers will use fallback NLP until Ollama recovers", "host", cfg.OllamaHost)
	}
	nlpSvc := nlp.NewService(ollamaClient)

	scraperSvc := scraper.NewService()
	proc := worker.NewProcessor(scraperSvc, nodeRepo, tagRepo, edgeRepo, jobRepo, nlpSvc)

	staleThreshold := time.Now().Add(-1 * time.Hour)
	count, err := nodeRepo.ResetStaleProcessingNodes(ctx, staleThreshold)
	if err != nil {
		slog.Error("failed to reset stale nodes", "error", err)
	} else if count > 0 {
		slog.Info("reset stale processing nodes", "count", count)
	}

	// Re-enqueue embeddings for nodes that completed without embeddings (Ollama recovery)
	if ollamaClient.Healthy(ctx) {
		nodeIDs, err := nodeRepo.ListNodesWithoutEmbedding(ctx, 100)
		if err != nil {
			slog.Error("failed to list nodes without embedding", "error", err)
		} else if len(nodeIDs) > 0 {
			for _, id := range nodeIDs {
				if _, err := jobRepo.CreateJob(ctx, "generate_embedding", map[string]string{"node_id": id}, 5); err != nil {
					slog.Error("failed to create reembed job", "node_id", id, "error", err)
				}
			}
			slog.Info("enqueued embedding recovery jobs", "count", len(nodeIDs))
		}
	}

	wp := worker.NewPool(jobRepo, proc, cfg.WorkerCount, cfg.JobTimeout)

	slog.Info("mesh-worker starting", "workers", cfg.WorkerCount)
	wp.Run(ctx)

	slog.Info("mesh-worker stopped")
	return nil
}
