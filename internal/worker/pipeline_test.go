package worker_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/neha037/mesh/internal/domain"
	"github.com/neha037/mesh/internal/nlp"
	"github.com/neha037/mesh/internal/ollama"
	"github.com/neha037/mesh/internal/worker"
)

type mockTagRepo struct {
	upsertFn    func(ctx context.Context, name string) (string, error)
	associateFn func(ctx context.Context, nodeID, tagID string, confidence float32) error
}

func (m *mockTagRepo) UpsertTag(ctx context.Context, name string) (string, error) {
	return m.upsertFn(ctx, name)
}
func (m *mockTagRepo) AssociateNodeTag(ctx context.Context, nodeID, tagID string, confidence float32) error {
	return m.associateFn(ctx, nodeID, tagID, confidence)
}
func (m *mockTagRepo) GetNodeTags(ctx context.Context, nodeID string) ([]domain.Tag, error) {
	return nil, nil
}

type mockEdgeRepo struct {
	buildFn   func(ctx context.Context, nodeID string) error
	upsertFn  func(ctx context.Context, sourceID, targetID string, weight float32) error
	similarFn func(ctx context.Context, embedding []float32, excludeID string, limit int32) ([]domain.SimilarNode, error)
}

func (m *mockEdgeRepo) BuildTagSharedEdges(ctx context.Context, nodeID string) error {
	return m.buildFn(ctx, nodeID)
}
func (m *mockEdgeRepo) UpsertSemanticEdge(ctx context.Context, sourceID, targetID string, weight float32) error {
	return m.upsertFn(ctx, sourceID, targetID, weight)
}
func (m *mockEdgeRepo) FindSimilarNodes(ctx context.Context, embedding []float32, excludeID string, limit int32) ([]domain.SimilarNode, error) {
	return m.similarFn(ctx, embedding, excludeID, limit)
}

type mockJobRepo struct {
	createFn func(ctx context.Context, jobType string, payload any, maxAttempts int32) (string, error)
}

func (m *mockJobRepo) CreateJob(ctx context.Context, jobType string, payload any, maxAttempts int32) (string, error) {
	return m.createFn(ctx, jobType, payload, maxAttempts)
}
func (m *mockJobRepo) ClaimJob(_ context.Context) (*domain.Job, error)                     { return nil, nil }
func (m *mockJobRepo) CompleteJob(_ context.Context, _ string) error                      { return nil }
func (m *mockJobRepo) FailJob(_ context.Context, _, _ string) error                       { return nil }
func (m *mockJobRepo) RetryJob(_ context.Context, _ string, _ int) error                  { return nil }

type mockNodeRepo struct {
	getStatusFn  func(ctx context.Context, id string) (domain.Node, error)
	updateStatFn func(ctx context.Context, id, status string) error
	updateEmbFn  func(ctx context.Context, id string, emb []float32, ver int32) (bool, error)
}

func (m *mockNodeRepo) GetNodeContent(ctx context.Context, id string) (domain.Node, error) {
	return m.getStatusFn(ctx, id)
}
func (m *mockNodeRepo) UpdateNodeStatus(ctx context.Context, id, status string) error {
	return m.updateStatFn(ctx, id, status)
}
func (m *mockNodeRepo) UpdateNodeEmbedding(ctx context.Context, id string, emb []float32, ver int32) (bool, error) {
	return m.updateEmbFn(ctx, id, emb, ver)
}
func (m *mockNodeRepo) UpsertRawNode(ctx context.Context, nodeType, title, content, sourceURL string) (domain.UpsertResult, error) {
	return domain.UpsertResult{}, nil
}
func (m *mockNodeRepo) GetNode(ctx context.Context, id string) (domain.Node, error) {
	return domain.Node{}, nil
}
func (m *mockNodeRepo) ListRecentNodes(ctx context.Context, limit int32) ([]domain.Node, error) {
	return nil, nil
}
func (m *mockNodeRepo) ListNodes(ctx context.Context, params domain.ListNodesParams) (domain.ListNodesResult, error) {
	return domain.ListNodesResult{}, nil
}
func (m *mockNodeRepo) DeleteNode(ctx context.Context, id string) error { return nil }
func (m *mockNodeRepo) UpdateNodeContent(ctx context.Context, id, content string) error {
	return nil
}

type mockOllama struct {
	extractFn func(ctx context.Context, content string) (ollama.TagResult, error)
	embedFn   func(ctx context.Context, text string) ([]float32, error)
}

func (m *mockOllama) ExtractTags(ctx context.Context, content string) (ollama.TagResult, error) {
	return m.extractFn(ctx, content)
}
func (m *mockOllama) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return m.embedFn(ctx, text)
}
func (m *mockOllama) Healthy(ctx context.Context) bool { return true }

func TestPipeline_ChainsJobs(t *testing.T) {
	ctx := context.Background()
	nodeID := "node-1"

	nodeRepo := &mockNodeRepo{
		getStatusFn: func(_ context.Context, id string) (domain.Node, error) {
			return domain.Node{ID: id, Content: "AI is cool", Version: 1}, nil
		},
		updateStatFn: func(_ context.Context, _, _ string) error { return nil },
		updateEmbFn:  func(_ context.Context, _ string, _ []float32, _ int32) (bool, error) { return true, nil },
	}

	tagRepo := &mockTagRepo{
		upsertFn:    func(_ context.Context, _ string) (string, error) { return "tag-1", nil },
		associateFn: func(_ context.Context, _ string, _ string, _ float32) error { return nil },
	}

	edgeRepo := &mockEdgeRepo{
		buildFn:   func(_ context.Context, _ string) error { return nil },
		upsertFn:  func(_ context.Context, _, _ string, _ float32) error { return nil },
		similarFn: func(_ context.Context, _ []float32, _ string, _ int32) ([]domain.SimilarNode, error) { return nil, nil },
	}

	var nextJobType string
	jobRepo := &mockJobRepo{
		createFn: func(_ context.Context, jobType string, _ any, _ int32) (string, error) {
			nextJobType = jobType
			return "job-2", nil
		},
	}

	ollamaMock := &mockOllama{
		extractFn: func(_ context.Context, _ string) (ollama.TagResult, error) {
			return ollama.TagResult{Tags: []string{"ai"}, Confidence: 0.9}, nil
		},
		embedFn: func(_ context.Context, _ string) ([]float32, error) {
			return make([]float32, 768), nil
		},
	}

	nlpSvc := nlp.NewService(ollamaMock)
	proc := worker.NewProcessor(nil, nodeRepo, tagRepo, edgeRepo, jobRepo, nlpSvc)

	t.Run("process_text enqueues generate_embedding", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{"node_id": nodeID})
		err := proc.Process(ctx, &domain.Job{Type: "process_text", Payload: payload})
		if err != nil {
			t.Fatalf("processText failed: %v", err)
		}
		if nextJobType != "generate_embedding" {
			t.Errorf("expected generate_embedding job, got %s", nextJobType)
		}
	})

	t.Run("generate_embedding enqueues build_edges", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{"node_id": nodeID})
		err := proc.Process(ctx, &domain.Job{Type: "generate_embedding", Payload: payload})
		if err != nil {
			t.Fatalf("generateEmbedding failed: %v", err)
		}
		if nextJobType != "build_edges" {
			t.Errorf("expected build_edges job, got %s", nextJobType)
		}
	})
}
