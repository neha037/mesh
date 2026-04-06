# Mesh Implementation Roadmap

The project is divided into **7 phases** spanning approximately **30 weeks**. Each phase delivers a working, standalone increment.

```
Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6 ──► Phase 7
 Senses      Brain      Memory       Eyes       Human     Discovery    Slow Burn
 Wk 1-3     Wk 4-6     Wk 7-9     Wk 10-14   Wk 15-18   Wk 19-24    Wk 25-30
```

## Phase 1: Foundation and Ingestion — "The Senses"
**Timeline:** Weeks 1-3
**Goal:** A working Go API that accepts data and persists it.
- Project Scaffolding
- PostgreSQL 16 + pgvector setup
- Basic REST Ingestion API
- Web Scraper Integration (Week 3)

## Phase 2: Processing and Intelligence — "The Brain"
**Timeline:** Weeks 4-6
**Goal:** Background workers for content extraction, tagging, and embeddings.
- Worker Pool Implementation
- HTML Stripping Pipeline
- Ollama (Gemma 4) Integration
- Automatic Edge Building

## Phase 3: Graph Traversal and Query API — "The Memory"
**Timeline:** Weeks 7-9
**Goal:** Rich query API for graph traversal, search, and similarity.
- Recursive CTE Traversal
- Full-text and Semantic Search
- Hybrid Search (RRF)
- Node/Edge CRUD

## Phase 4: Frontend Visualization — "The Eyes"
**Timeline:** Weeks 10-14
**Goal:** Interactive knowledge graph web UI.
- React + TypeScript + Vite Scaffolding
- Cytoscape.js Rendering
- Search/Filter Dashboard
- Node Detail Side Panels

## Phase 5: Multi-Modal and Journaling — "The Human Element"
**Timeline:** Weeks 15-18
- MinIO Object Storage for images/PDFs
- Image Understanding (Gemma 4 Vision)
- PDF and Voice Note Ingestion
- Rich Text Journaling

## Phase 6: Anti-Echo Chamber Engine — "Discovery"
**Timeline:** Weeks 19-24
- Cluster Density Analysis
- Bridge Detection (Adjacent Possible)
- Wildcard Topic Injection
- De-duplication Logic

## Phase 7: Spaced Repetition and Semantic Depth — "The Slow Burn"
**Timeline:** Weeks 25-30
- FSRS Spaced Repetition Implementation
- Knowledge Decay Visualization
- Nightly Semantic Edge Builder
- Serendipity Metrics Tracking
