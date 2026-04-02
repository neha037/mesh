---
layout: default
title: Roadmap
nav_order: 7
---

# Roadmap

Mesh is developed in 7 phases over approximately 30 weeks.

## Phase Overview

| Phase | Name | Duration | Status |
|-------|------|----------|--------|
| 1 | Foundation and Ingestion -- "The Senses" | Weeks 1-3 | **In Progress** |
| 2 | Processing and Intelligence -- "The Brain" | Weeks 4-6 | Not Started |
| 3 | Graph Traversal and Query API -- "The Memory" | Weeks 7-9 | Not Started |
| 4 | Frontend Visualization -- "The Eyes" | Weeks 10-14 | Not Started |
| 5 | Multi-Modal and Journaling -- "The Human Element" | Weeks 15-18 | Not Started |
| 6 | Anti-Echo Chamber Engine -- "Discovery" | Weeks 19-24 | Not Started |
| 7 | Spaced Repetition and Semantic Depth -- "The Slow Burn" | Weeks 25-30 | Not Started |

---

## Phase 1: Foundation and Ingestion

Build the core infrastructure and basic page ingestion.

- [x] Go backend with chi router
- [x] PostgreSQL 16 + pgvector database
- [x] MinIO object storage
- [x] Docker Compose orchestration
- [x] Browser extension (one-click save)
- [x] URL deduplication (upsert)
- [x] Cursor-based pagination
- [x] System tray with service controls
- [x] Systemd integration with autostart
- [ ] Web scraper with circuit breaker
- [ ] Background job queue

## Phase 2: Processing and Intelligence

Automatic content processing with AI.

- Tag extraction using local LLM (Ollama + Mistral 7B)
- Embedding generation (nomic-embed-text, 384 dimensions)
- Automatic edge building between related content
- Fallback NLP when Ollama is unavailable

## Phase 3: Graph Traversal and Query API

Search and explore your knowledge.

- Full-text search (PostgreSQL trigram)
- Semantic search (pgvector cosine similarity)
- Hybrid search (Reciprocal Rank Fusion)
- Graph traversal API (BFS with depth limits)
- Node CRUD and filtering

## Phase 4: Frontend Visualization

Interactive knowledge graph in the browser.

- React + TypeScript + Cytoscape.js
- Interactive graph with color-coded nodes
- Click-to-explore local subgraphs
- Search bar with multiple modes
- Filter by type, date, and tags

## Phase 5: Multi-Modal and Journaling

Beyond web pages.

- Image upload and storage (MinIO)
- Journal entries with rich text
- Vision model descriptions for images
- Timeline view

## Phase 6: Anti-Echo Chamber Engine

Fight intellectual stagnation.

- Cluster density analysis
- Knowledge gap detection
- Bridge detection between isolated clusters
- Wildcard injection from external sources (Wikipedia, HN, arXiv)
- Serendipity metrics

## Phase 7: Spaced Repetition and Semantic Depth

Long-term retention.

- FSRS v5 spaced repetition algorithm
- Daily review cards
- "Surprisingly similar" content suggestions
- Semantic edge building (nightly batch)
