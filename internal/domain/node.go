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
}
