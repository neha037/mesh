package main

import (
	"log"

	"github.com/neha037/mesh/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("mesh-worker starting (workers=%d, log_level=%s)", cfg.WorkerCount, cfg.LogLevel)
}
