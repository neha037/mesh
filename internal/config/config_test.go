package config

import (
	"testing"
)

// configEnvVars lists all environment variables that Load() reads.
var configEnvVars = []string{
	"DATABASE_URL", "WORKER_COUNT", "MINIO_ENDPOINT", "MINIO_ACCESS_KEY",
	"MINIO_SECRET_KEY", "MINIO_BUCKET", "OLLAMA_HOST", "OLLAMA_MODEL",
	"EMBEDDING_MODEL", "SERVER_PORT", "ALLOWED_ORIGINS", "LOG_LEVEL",
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		expectError bool
		checkConfig func(*testing.T, *Config)
	}{
		{
			name: "missing DATABASE_URL",
			env: map[string]string{
				"WORKER_COUNT": "4",
			},
			expectError: true,
		},
		{
			name: "valid configuration",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost",
			},
			expectError: false,
			checkConfig: func(t *testing.T, c *Config) {
				if c.DatabaseURL != "postgres://localhost" {
					t.Errorf("Expected DatabaseURL postgres://localhost, got %v", c.DatabaseURL)
				}
				if c.WorkerCount != 4 {
					t.Errorf("Expected WorkerCount 4, got %v", c.WorkerCount)
				}
			},
		},
		{
			name: "invalid WORKER_COUNT",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost",
				"WORKER_COUNT": "abc",
			},
			expectError: true,
		},
		{
			name: "custom worker count",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost",
				"WORKER_COUNT": "10",
			},
			expectError: false,
			checkConfig: func(t *testing.T, c *Config) {
				if c.WorkerCount != 10 {
					t.Errorf("Expected WorkerCount 10, got %v", c.WorkerCount)
				}
			},
		},
		{
			name: "zero WORKER_COUNT rejected",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost",
				"WORKER_COUNT": "0",
			},
			expectError: true,
		},
		{
			name: "negative WORKER_COUNT rejected",
			env: map[string]string{
				"DATABASE_URL": "postgres://localhost",
				"WORKER_COUNT": "-1",
			},
			expectError: true,
		},
		{
			name: "multiple origins",
			env: map[string]string{
				"DATABASE_URL":    "postgres://localhost",
				"ALLOWED_ORIGINS": "http://a.com , http://b.com",
			},
			expectError: false,
			checkConfig: func(t *testing.T, c *Config) {
				if len(c.AllowedOrigins) != 2 {
					t.Fatalf("Expected 2 origins, got %d", len(c.AllowedOrigins))
				}
				if c.AllowedOrigins[0] != "http://a.com" || c.AllowedOrigins[1] != "http://b.com" {
					t.Errorf("Origins parsed incorrectly: %v", c.AllowedOrigins)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all config env vars (auto-restored after test)
			for _, key := range configEnvVars {
				t.Setenv(key, "")
			}
			// Set test-specific env vars
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if (err != nil) != tt.expectError {
				t.Errorf("Load() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if err == nil && tt.checkConfig != nil {
				tt.checkConfig(t, cfg)
			}
		})
	}
}
