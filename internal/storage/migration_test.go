//go:build integration

package storage_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/neha037/mesh/migrations"
)

// newTestMigrator creates a migrate.Migrate instance for testing.
func newTestMigrator(t *testing.T, connStr string) *migrate.Migrate {
	t.Helper()

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		t.Fatalf("creating migration source: %v", err)
	}

	dbURL := strings.Replace(connStr, "postgresql://", "pgx5://", 1)
	dbURL = strings.Replace(dbURL, "postgres://", "pgx5://", 1)

	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		t.Fatalf("creating migrator: %v", err)
	}

	return m
}

// startPostgres spins up a pgvector container and returns its connection string.
func startPostgres(t *testing.T) (connStr string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("mesh_migration_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("starting postgres container: %v", err)
	}

	cs, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("getting connection string: %v", err)
	}

	return cs, func() {
		if err := ctr.Terminate(ctx); err != nil {
			t.Logf("terminating container: %v", err)
		}
	}
}

// TestMigrations_RoundTrip applies all migrations up, then all down,
// then all up again. This catches syntax errors, non-immutable functions
// in indexes, and broken down migrations.
func TestMigrations_RoundTrip(t *testing.T) {
	connStr, cleanup := startPostgres(t)
	defer cleanup()

	m := newTestMigrator(t, connStr)
	defer m.Close()

	// Up: apply all migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrating up: %v", err)
	}

	// Down: roll back all migrations
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrating down: %v", err)
	}

	// Up again: re-apply all migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrating up (second pass): %v", err)
	}
}

// TestMigrations_StepByStep applies each migration individually, then
// rolls it back and re-applies it. This pinpoints exactly which migration
// is broken if the round-trip test fails.
func TestMigrations_StepByStep(t *testing.T) {
	connStr, cleanup := startPostgres(t)
	defer cleanup()

	m := newTestMigrator(t, connStr)
	defer m.Close()

	// Discover how many migrations exist by migrating all the way up.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("initial up: %v", err)
	}
	version, _, err := m.Version()
	if err != nil {
		t.Fatalf("getting version: %v", err)
	}
	maxVersion := version

	// Reset to a clean slate.
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("resetting down: %v", err)
	}

	// Need a fresh migrator after Down since the source may be exhausted.
	m.Close()
	m = newTestMigrator(t, connStr)
	defer m.Close()

	// Step through each version.
	for v := uint(1); v <= maxVersion; v++ {
		t.Run("migration_"+itoa(v), func(t *testing.T) {
			// Up to version v
			if err := m.Migrate(v); err != nil {
				t.Fatalf("up to version %d: %v", v, err)
			}

			// Down one step
			if v == 1 {
				if err := m.Down(); err != nil {
					t.Fatalf("down from version %d: %v", v, err)
				}
			} else {
				if err := m.Migrate(v - 1); err != nil {
					t.Fatalf("down to version %d: %v", v-1, err)
				}
			}

			// Re-apply
			if err := m.Migrate(v); err != nil {
				t.Fatalf("re-up to version %d: %v", v, err)
			}
		})
	}
}

func itoa(v uint) string {
	return fmt.Sprintf("%d", v)
}
