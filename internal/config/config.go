package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	OllamaHost     string
	WorkerCount    int
	LogLevel       string
}

// Load reads configuration from environment variables and returns a Config.
// Returns an error if required variables are missing.
func Load() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	workerCount := 4
	if v := os.Getenv("WORKER_COUNT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("WORKER_COUNT must be an integer: %w", err)
		}
		workerCount = n
	}

	return &Config{
		DatabaseURL:    dbURL,
		MinioEndpoint:  getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnvOrDefault("MINIO_ACCESS_KEY", "meshadmin"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    getEnvOrDefault("MINIO_BUCKET", "mesh-images"),
		OllamaHost:     getEnvOrDefault("OLLAMA_HOST", "http://localhost:11434"),
		WorkerCount:    workerCount,
		LogLevel:       getEnvOrDefault("LOG_LEVEL", "info"),
	}, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
