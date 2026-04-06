//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/pgvector/pgvector-go"

	"github.com/neha037/mesh/internal/storage"
)

func TestBuildTagSharedEdges_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	nodeRepo := storage.NewNodeRepo(q)
	tagRepo := storage.NewTagRepo(q)
	edgeRepo := storage.NewEdgeRepo(q)
	ctx := context.Background()

	// Create nodes A and B
	resA, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node A", "content A", "https://example.com/node-a")
	resB, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node B", "content B", "https://example.com/node-b")

	nodeA := resA.Node.ID
	nodeB := resB.Node.ID

	// Create 3 tags and associate all with both A and B
	tagNames := []string{"golang", "posgres", "pgvector"}
	for _, name := range tagNames {
		tagID, err := tagRepo.UpsertTag(ctx, name)
		if err != nil {
			t.Fatalf("upserting tag: %v", err)
		}
		if err := tagRepo.AssociateNodeTag(ctx, nodeA, tagID, 0.9); err != nil {
			t.Fatalf("associating tag A: %v", err)
		}
		if err := tagRepo.AssociateNodeTag(ctx, nodeB, tagID, 0.9); err != nil {
			t.Fatalf("associating tag B: %v", err)
		}
	}

	t.Run("build edge when 3 tags shared", func(t *testing.T) {
		if err := edgeRepo.BuildTagSharedEdges(ctx, nodeA); err != nil {
			t.Fatalf("building shared edges: %v", err)
		}

		// Verify edge exists from A to B
		rows, err := pool.Query(ctx, "SELECT source_id, target_id, rel_type, weight FROM edges WHERE source_id = $1 AND target_id = $2 AND rel_type = 'tag_shared'", nodeA, nodeB)
		if err != nil {
			t.Fatalf("querying edges: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected tag_shared edge to exist")
		}
	})
}

func TestBuildTagSharedEdges_NoEdgeForSingleSharedTag(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	nodeRepo := storage.NewNodeRepo(q)
	tagRepo := storage.NewTagRepo(q)
	edgeRepo := storage.NewEdgeRepo(q)
	ctx := context.Background()

	resA, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node A", "", "https://example.com/a")
	resB, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node B", "", "https://example.com/b")

	tid, _ := tagRepo.UpsertTag(ctx, "shared")
	tagRepo.AssociateNodeTag(ctx, resA.Node.ID, tid, 0.9)
	tagRepo.AssociateNodeTag(ctx, resB.Node.ID, tid, 0.9)

	if err := edgeRepo.BuildTagSharedEdges(ctx, resA.Node.ID); err != nil {
		t.Fatalf("building shared edges: %v", err)
	}

	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM edges WHERE source_id = $1 AND target_id = $2 AND rel_type = 'tag_shared')", resA.Node.ID, resB.Node.ID).Scan(&exists)
	if err != nil {
		t.Fatalf("querying edges: %v", err)
	}

	if exists {
		t.Error("did not expect edge for single shared tag (min shared tags = 2)")
	}
}

func TestUpsertSemanticEdge_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	nodeRepo := storage.NewNodeRepo(q)
	edgeRepo := storage.NewEdgeRepo(q)
	ctx := context.Background()

	resA, _ := nodeRepo.UpsertRawNode(ctx, "article", "A", "", "https://example.com/a")
	resB, _ := nodeRepo.UpsertRawNode(ctx, "article", "B", "", "https://example.com/b")

	nodeA := resA.Node.ID
	nodeB := resB.Node.ID

	if err := edgeRepo.UpsertSemanticEdge(ctx, nodeA, nodeB, 0.82); err != nil {
		t.Fatalf("upserting edge: %v", err)
	}

	var weight float32
	err := pool.QueryRow(ctx, "SELECT weight FROM edges WHERE source_id = $1 AND target_id = $2 AND rel_type = 'semantic'", nodeA, nodeB).Scan(&weight)
	if err != nil {
		t.Fatalf("querying edges: %v", err)
	}

	if weight != 0.82 {
		t.Errorf("expected weight 0.82, got %f", weight)
	}
}

func TestFindSimilarNodes_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	nodeRepo := storage.NewNodeRepo(q)
	edgeRepo := storage.NewEdgeRepo(q)
	ctx := context.Background()

	// Helper to set embedding
	setEmbedding := func(nodeID string, vec []float32, ver int32) {
		res, err := pool.Exec(ctx, "UPDATE nodes SET embedding = $1 WHERE id = $2", pgvector.NewVector(vec), nodeID)
		if err != nil || res.RowsAffected() == 0 {
			t.Fatalf("failed to set embedding for node %s: %v", nodeID, err)
		}
	}

	resA, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node A", "", "https://example.com/a")
	resB, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node B", "", "https://example.com/b")
	resC, _ := nodeRepo.UpsertRawNode(ctx, "article", "Node C", "", "https://example.com/c")

	nodeA := resA.Node.ID
	nodeB := resB.Node.ID
	nodeC := resC.Node.ID

	vecA := make([]float32, 768)
	vecA[0] = 1.0

	vecB := make([]float32, 768)
	vecB[0] = 0.95
	vecB[1] = 0.05

	vecC := make([]float32, 768)
	vecC[767] = 1.0

	setEmbedding(nodeA, vecA, resA.Node.Version)
	setEmbedding(nodeB, vecB, resB.Node.Version)
	setEmbedding(nodeC, vecC, resC.Node.Version)

	similar, err := edgeRepo.FindSimilarNodes(ctx, vecA, nodeA, 5)
	if err != nil {
		t.Fatalf("finding similar nodes: %v", err)
	}

	if len(similar) != 2 {
		t.Fatalf("expected 2 similar nodes, got %d", len(similar))
	}
	// B should be more similar than C
	if similar[0].ID != nodeB {
		t.Errorf("expected node B to be most similar, got %s", similar[0].ID)
	}
}
