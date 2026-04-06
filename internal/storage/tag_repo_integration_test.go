//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/neha037/mesh/internal/storage"
)

func TestUpsertTag_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	repo := storage.NewTagRepo(q)
	ctx := context.Background()

	t.Run("insert new tag", func(t *testing.T) {
		id1, err := repo.UpsertTag(ctx, "golang")
		if err != nil {
			t.Fatalf("upsert 1: %v", err)
		}
		if id1 == "" {
			t.Error("expected non-empty ID")
		}

		id2, err := repo.UpsertTag(ctx, "golang")
		if err != nil {
			t.Fatalf("upsert 2: %v", err)
		}
		if id1 != id2 {
			t.Errorf("expected same ID for duplicate tag, got %s and %s", id1, id2)
		}
	})
}

func TestAssociateNodeTag_Integration(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	nodeRepo := storage.NewNodeRepo(q)
	tagRepo := storage.NewTagRepo(q)
	ctx := context.Background()

	// Create test node
	res, err := nodeRepo.UpsertRawNode(ctx, "article", "Test Node", "content", "https://example.com/tag-test")
	if err != nil {
		t.Fatalf("creating node: %v", err)
	}
	nodeID := res.Node.ID

	// Create tag
	tagID, err := tagRepo.UpsertTag(ctx, "integration-test")
	if err != nil {
		t.Fatalf("creating tag: %v", err)
	}

	t.Run("associate and retrieve", func(t *testing.T) {
		if err := tagRepo.AssociateNodeTag(ctx, nodeID, tagID, 0.85); err != nil {
			t.Fatalf("associating: %v", err)
		}

		tags, err := tagRepo.GetNodeTags(ctx, nodeID)
		if err != nil {
			t.Fatalf("getting tags: %v", err)
		}

		if len(tags) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(tags))
		}
		if tags[0].Name != "integration-test" {
			t.Errorf("expected tag name 'integration-test', got %q", tags[0].Name)
		}
		if tags[0].Confidence != 0.85 {
			t.Errorf("expected confidence 0.85, got %f", tags[0].Confidence)
		}
	})

	t.Run("higher confidence wins", func(t *testing.T) {
		if err := tagRepo.AssociateNodeTag(ctx, nodeID, tagID, 0.95); err != nil {
			t.Fatalf("associating higher: %v", err)
		}

		tags, _ := tagRepo.GetNodeTags(ctx, nodeID)
		if tags[0].Confidence != 0.95 {
			t.Errorf("expected updated confidence 0.95, got %f", tags[0].Confidence)
		}

		// Lower confidence should NOT win
		if err := tagRepo.AssociateNodeTag(ctx, nodeID, tagID, 0.70); err != nil {
			t.Fatalf("associating lower: %v", err)
		}
		tags, _ = tagRepo.GetNodeTags(ctx, nodeID)
		if tags[0].Confidence != 0.95 {
			t.Errorf("expected confidence to remain 0.95, got %f", tags[0].Confidence)
		}
	})
}

func TestGetNodeTags_OrderedByConfidence(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	q := storage.New(pool)
	nodeRepo := storage.NewNodeRepo(q)
	tagRepo := storage.NewTagRepo(q)
	ctx := context.Background()

	res, _ := nodeRepo.UpsertRawNode(ctx, "article", "Sorted Tags", "", "https://example.com/sorted")
	nodeID := res.Node.ID

	tagData := []struct {
		name string
		conf float32
	}{
		{"medium", 0.7},
		{"high", 0.9},
		{"low", 0.5},
	}

	for _, d := range tagData {
		tid, _ := tagRepo.UpsertTag(ctx, d.name)
		tagRepo.AssociateNodeTag(ctx, nodeID, tid, d.conf)
	}

	tags, err := tagRepo.GetNodeTags(ctx, nodeID)
	if err != nil {
		t.Fatalf("getting tags: %v", err)
	}

	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[0].Name != "high" || tags[1].Name != "medium" || tags[2].Name != "low" {
		t.Errorf("expected order [high, medium, low], got [%s, %s, %s]", tags[0].Name, tags[1].Name, tags[2].Name)
	}
}
