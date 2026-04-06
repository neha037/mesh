package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/neha037/mesh/internal/config"
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
	scraperSvc := scraper.NewService()
	proc := worker.NewProcessor(scraperSvc, nodeRepo)

	wp := worker.NewPool(jobRepo, proc, cfg.WorkerCount)

	slog.Info("mesh-worker starting", "workers", cfg.WorkerCount)
	wp.Run(ctx)

	slog.Info("mesh-worker stopped")
	return nil
}
