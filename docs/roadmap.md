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
| 1 | Foundation and Ingestion -- "The Senses" | Weeks 1-3 | **Complete** |
| 2 | Processing and Intelligence -- "The Brain" | Weeks 4-6 | **In Progress** |
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
- [x] Web scraper with circuit breaker
- [x] Background job queue

## Phase 2: Processing and Intelligence

Automatic content processing with AI.

- Tag extraction using local LLM (Ollama + Gemma 4 E4B, structured JSON output)
- Embedding generation (EmbeddingGemma-300M, 768 dimensions, Matryoshka)
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

Beyond web pages — images, PDFs, voice notes, and export.

- Image upload and AI description (Gemma 4 native vision)
- PDF ingestion with native parsing and OCR (Gemma 4)
- Voice note ingestion with native ASR transcription (Gemma 4)
- Journal entries with rich text editor
- Subgraph export (Markdown, JSON-LD, PNG, Obsidian-compatible)
- Gallery and timeline views

## Phase 6: Anti-Echo Chamber Engine

Fight intellectual stagnation.

- Cluster density analysis
- Knowledge gap detection
- Bridge detection between isolated clusters
- Wildcard injection from external sources (Wikipedia, HN, arXiv)
- Automatic de-duplication (cosine similarity > 0.90, merge suggestions)
- Serendipity metrics

## Phase 7: Spaced Repetition and Semantic Depth

Long-term retention.

- FSRS v5 spaced repetition algorithm
- Daily review cards
- Knowledge decay visualization (node opacity maps to retrievability)
- "Surprisingly similar" content suggestions
- Semantic edge building (nightly batch)

## Future Enhancements (Post-Phase 7)

- LoRA fine-tuning for personalized tagging (learns user's taxonomy)
- Mobile companion app for voice note capture
- RSS feed ingestion
- Plugin system for custom sources (Kindle, Twitter, Pocket)
