//go:build integration

package storage_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/neha037/mesh/internal/domain"
	"github.com/neha037/mesh/internal/storage"
	"github.com/neha037/mesh/migrations"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	ctr, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("mesh_test"),
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

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("getting connection string: %v", err)
	}

	if err := storage.RunMigrations(migrations.FS, connStr); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("creating pool: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := ctr.Terminate(ctx); err != nil {
			t.Logf("terminating container: %v", err)
		}
	}

	return pool, cleanup
}

func TestUpsertRawNode_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := storage.NewNodeRepo(storage.New(pool))
	ctx := context.Background()

	t.Run("insert new node", func(t *testing.T) {
		result, err := repo.UpsertRawNode(ctx, "article", "Test Article", "Some content", "https://example.com/1")
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}
		if !result.Created {
			t.Error("expected Created=true for new node")
		}
		if result.Node.ID == "" {
			t.Error("expected non-empty ID")
		}
		if result.Node.Title != "Test Article" {
			t.Errorf("title = %q, want %q", result.Node.Title, "Test Article")
		}
	})

	t.Run("upsert updates existing node", func(t *testing.T) {
		result, err := repo.UpsertRawNode(ctx, "article", "Updated Title", "Updated content", "https://example.com/1")
		if err != nil {
			t.Fatalf("upsert: %v", err)
		}
		if result.Created {
			t.Error("expected Created=false for upsert")
		}
		if result.Node.Title != "Updated Title" {
			t.Errorf("title = %q, want %q", result.Node.Title, "Updated Title")
		}
	})
}

func TestListRecentNodes_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := storage.NewNodeRepo(storage.New(pool))
	ctx := context.Background()

	// Insert 3 nodes
	for i, title := range []string{"First", "Second", "Third"} {
		_, err := repo.UpsertRawNode(ctx, "article", title, "", "https://example.com/list-"+title)
		if err != nil {
			t.Fatalf("inserting node %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond) // ensure distinct created_at
	}

	nodes, err := repo.ListRecentNodes(ctx, 2)
	if err != nil {
		t.Fatalf("list recent: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("len = %d, want 2", len(nodes))
	}
	// Most recent first
	if nodes[0].Title != "Third" {
		t.Errorf("first node title = %q, want %q", nodes[0].Title, "Third")
	}
	if nodes[1].Title != "Second" {
		t.Errorf("second node title = %q, want %q", nodes[1].Title, "Second")
	}
}

func TestListNodes_Pagination_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := storage.NewNodeRepo(storage.New(pool))
	ctx := context.Background()

	// Insert 5 nodes
	for i := range 5 {
		_, err := repo.UpsertRawNode(ctx, "article", "Node "+string(rune('A'+i)), "", "https://example.com/page-"+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("inserting node %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Page 1: limit 2
	page1, err := repo.ListNodes(ctx, domain.ListNodesParams{Limit: 2})
	if err != nil {
		t.Fatalf("page 1: %v", err)
	}
	if len(page1.Nodes) != 2 {
		t.Fatalf("page 1 len = %d, want 2", len(page1.Nodes))
	}
	if !page1.HasMore {
		t.Error("expected HasMore=true for page 1")
	}

	// Page 2: use cursor from last node of page 1
	last := page1.Nodes[len(page1.Nodes)-1]
	page2, err := repo.ListNodes(ctx, domain.ListNodesParams{
		Limit:    2,
		CursorAt: &last.CreatedAt,
		CursorID: &last.ID,
	})
	if err != nil {
		t.Fatalf("page 2: %v", err)
	}
	if len(page2.Nodes) != 2 {
		t.Fatalf("page 2 len = %d, want 2", len(page2.Nodes))
	}
	if !page2.HasMore {
		t.Error("expected HasMore=true for page 2")
	}

	// Page 3: last page
	last = page2.Nodes[len(page2.Nodes)-1]
	page3, err := repo.ListNodes(ctx, domain.ListNodesParams{
		Limit:    2,
		CursorAt: &last.CreatedAt,
		CursorID: &last.ID,
	})
	if err != nil {
		t.Fatalf("page 3: %v", err)
	}
	if len(page3.Nodes) != 1 {
		t.Fatalf("page 3 len = %d, want 1", len(page3.Nodes))
	}
	if page3.HasMore {
		t.Error("expected HasMore=false for last page")
	}
}

func TestGetNode_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := storage.NewNodeRepo(storage.New(pool))
	ctx := context.Background()

	t.Run("get existing node", func(t *testing.T) {
		result, err := repo.UpsertRawNode(ctx, "article", "Get Me", "some content", "https://example.com/get-me")
		if err != nil {
			t.Fatalf("inserting node: %v", err)
		}

		node, err := repo.GetNode(ctx, result.Node.ID)
		if err != nil {
			t.Fatalf("get node: %v", err)
		}
		if node.Title != "Get Me" {
			t.Errorf("title = %q, want %q", node.Title, "Get Me")
		}
		if node.Type != "article" {
			t.Errorf("type = %q, want %q", node.Type, "article")
		}
		if node.Status != "pending" {
			t.Errorf("status = %q, want %q", node.Status, "pending")
		}
	})

	t.Run("get non-existent returns ErrNotFound", func(t *testing.T) {
		_, err := repo.GetNode(ctx, "00000000-0000-0000-0000-000000000000")
		if err != domain.ErrNotFound {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("get invalid UUID returns error", func(t *testing.T) {
		_, err := repo.GetNode(ctx, "not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid UUID")
		}
	})
}

func TestListNodes_EdgeCases_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := storage.NewNodeRepo(storage.New(pool))
	ctx := context.Background()

	t.Run("empty database returns empty list", func(t *testing.T) {
		result, err := repo.ListNodes(ctx, domain.ListNodesParams{Limit: 10})
		if err != nil {
			t.Fatalf("list nodes: %v", err)
		}
		if len(result.Nodes) != 0 {
			t.Errorf("len = %d, want 0", len(result.Nodes))
		}
		if result.HasMore {
			t.Error("expected HasMore=false for empty DB")
		}
	})

	t.Run("single node no more pages", func(t *testing.T) {
		_, err := repo.UpsertRawNode(ctx, "article", "Only One", "", "https://example.com/only-one")
		if err != nil {
			t.Fatalf("inserting: %v", err)
		}

		result, err := repo.ListNodes(ctx, domain.ListNodesParams{Limit: 10})
		if err != nil {
			t.Fatalf("list nodes: %v", err)
		}
		if len(result.Nodes) != 1 {
			t.Errorf("len = %d, want 1", len(result.Nodes))
		}
		if result.HasMore {
			t.Error("expected HasMore=false for single node")
		}
	})
}

// TestGetNode_CoversAllSchemaColumns_Integration verifies that GetNodeRow includes
// every column in the nodes table, or explicitly documents why a column is excluded.
// This prevents schema drift where a migration adds a column but queries are not updated.
func TestGetNode_CoversAllSchemaColumns_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Columns intentionally excluded from GetNodeRow with reasons.
	// When a new column is added via migration, this test will fail unless
	// the column is either added to the GetNode query or listed here.
	excluded := map[string]string{
		"embedding": "large vector column, fetched separately in Phase 2",
	}

	// Get all columns from the nodes table via information_schema.
	rows, err := pool.Query(ctx,
		`SELECT column_name FROM information_schema.columns
		 WHERE table_name = 'nodes' ORDER BY ordinal_position`)
	if err != nil {
		t.Fatalf("querying schema: %v", err)
	}
	defer rows.Close()

	var schemaCols []string
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			t.Fatalf("scanning column: %v", err)
		}
		schemaCols = append(schemaCols, col)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterating columns: %v", err)
	}

	if len(schemaCols) == 0 {
		t.Fatal("no columns found in nodes table — migrations may not have run")
	}

	// Build a set of JSON tag names from GetNodeRow struct fields.
	rt := reflect.TypeOf(storage.GetNodeRow{})
	structTags := make(map[string]bool, rt.NumField())
	for i := range rt.NumField() {
		tag := rt.Field(i).Tag.Get("json")
		if tag != "" && tag != "-" {
			structTags[tag] = true
		}
	}

	// Every schema column must be in the struct or in the exclusion list.
	for _, col := range schemaCols {
		if _, ok := excluded[col]; ok {
			continue
		}
		if !structTags[col] {
			t.Errorf("schema column %q exists in nodes table but is missing from GetNodeRow struct and not in exclusion list — "+
				"either add it to the GetNode query in queries/nodes.sql and run 'make sqlc', "+
				"or add it to the excluded map in this test with a reason", col)
		}
	}
}

func TestDeleteNode_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := storage.NewNodeRepo(storage.New(pool))
	ctx := context.Background()

	result, err := repo.UpsertRawNode(ctx, "article", "To Delete", "", "https://example.com/delete")
	if err != nil {
		t.Fatalf("inserting node: %v", err)
	}

	t.Run("delete existing node", func(t *testing.T) {
		if err := repo.DeleteNode(ctx, result.Node.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
	})

	t.Run("delete non-existent returns ErrNotFound", func(t *testing.T) {
		err := repo.DeleteNode(ctx, result.Node.ID)
		if err != domain.ErrNotFound {
			t.Errorf("err = %v, want ErrNotFound", err)
		}
	})

	t.Run("delete invalid UUID returns error", func(t *testing.T) {
		err := repo.DeleteNode(ctx, "not-a-uuid")
		if err == nil {
			t.Error("expected error for invalid UUID")
		}
	})
}
