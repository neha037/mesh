package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	OllamaHost     string
	OllamaModel    string
	EmbeddingModel string
	ServerPort     string
	AllowedOrigins []string
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
		if n < 1 {
			return nil, fmt.Errorf("WORKER_COUNT must be >= 1, got %d", n)
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
		OllamaModel:    getEnvOrDefault("OLLAMA_MODEL", "gemma4:e4b"),
		EmbeddingModel: getEnvOrDefault("EMBEDDING_MODEL", "embeddinggemma:300m-qat-q8_0"),
		ServerPort:     getEnvOrDefault("SERVER_PORT", "8080"),
		AllowedOrigins: loadOrigins(),
		WorkerCount:    workerCount,
		LogLevel:       getEnvOrDefault("LOG_LEVEL", "info"),
	}, nil
}

func loadOrigins() []string {
	v := getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000,chrome-extension://*")
	var origins []string
	for _, o := range strings.Split(v, ",") {
		origins = append(origins, strings.TrimSpace(o))
	}
	return origins
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
