package domain

import (
	"context"
	"time"
)

// Node represents a knowledge entity in the graph.
type Node struct {
	ID        string
	Type      string
	Title     string
	Content   string
	Summary   string
	SourceURL string
	ImageKey  string
	Status    string
	Version   int32
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UpsertResult holds the result of an upsert operation.
type UpsertResult struct {
	Node    Node
	Created bool
}

// ListNodesParams holds parameters for paginated node listing.
type ListNodesParams struct {
	Limit    int32
	CursorAt *time.Time // nil means first page
	CursorID *string    // nil means first page
}

// ListNodesResult holds a page of nodes with pagination info.
type ListNodesResult struct {
	Nodes   []Node
	HasMore bool
}

// NodeRepository defines the interface for node storage operations.
// Handlers depend on this interface, not on concrete storage implementations.
type NodeRepository interface {
	UpsertRawNode(ctx context.Context, nodeType, title, content, sourceURL string) (UpsertResult, error)
	GetNode(ctx context.Context, id string) (Node, error)
	ListRecentNodes(ctx context.Context, limit int32) ([]Node, error)
	ListNodes(ctx context.Context, params ListNodesParams) (ListNodesResult, error)
	DeleteNode(ctx context.Context, id string) error
	UpdateNodeContent(ctx context.Context, id string, content string) error
	UpdateNodeStatus(ctx context.Context, id string, status string) error
	UpdateNodeEmbedding(ctx context.Context, id string, embedding []float32, expectedVersion int32) (bool, error)
	GetNodeContent(ctx context.Context, id string) (Node, error)
}

// Tag represents a concept tag extracted from content.
type Tag struct {
	ID         string
	Name       string
	Confidence float32
}

// TagRepository defines the interface for tag storage operations.
type TagRepository interface {
	UpsertTag(ctx context.Context, name string) (string, error)
	AssociateNodeTag(ctx context.Context, nodeID, tagID string, confidence float32) error
	GetNodeTags(ctx context.Context, nodeID string) ([]Tag, error)
}

// SimilarNode holds a node ID and its similarity score from vector search.
type SimilarNode struct {
	ID         string
	Title      string
	Similarity float32
}

// EdgeRepository defines the interface for edge storage operations.
type EdgeRepository interface {
	BuildTagSharedEdges(ctx context.Context, nodeID string) error
	UpsertSemanticEdge(ctx context.Context, sourceID, targetID string, weight float32) error
	FindSimilarNodes(ctx context.Context, embedding []float32, excludeID string, limit int32) ([]SimilarNode, error)
}
