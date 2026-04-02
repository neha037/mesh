---
layout: default
title: Home
nav_order: 1
---

# Mesh

A local-first **Personal Growth Engine** that maps your knowledge into an interactive graph — with automatic connections, intelligent discovery, and spaced repetition. Everything runs on your machine. Zero cloud costs, total data sovereignty.

---

## Features

| Feature | Status |
|---------|--------|
| One-click page saving (browser extension) | Available |
| Automatic URL deduplication | Available |
| View and manage all saved pages | Available |
| System tray with service controls | Available |
| Autostart on login (systemd) | Available |
| REST API for ingestion and retrieval | Available |
| Cursor-based pagination (scalable) | Available |
| Background workers (scraping, NLP) | Coming Soon |
| AI-powered tag extraction (Ollama) | Coming Soon |
| Semantic search (pgvector) | Coming Soon |
| Interactive knowledge graph (Cytoscape.js) | Coming Soon |
| Spaced repetition (FSRS) | Coming Soon |
| Discovery engine (anti-echo chamber) | Coming Soon |

---

## How It Works

```
  You browse the web
        |
        v
  [Browser Extension] ---> [Go API Server :8080]
   one-click save              |
                               v
                        [PostgreSQL + pgvector]
                         knowledge storage
                               |
                               v
                        [Background Workers]
                         tagging, embeddings,
                         edge building
                               |
                               v
                        [React Dashboard]
                         graph visualization,
                         search, review
```

The extension saves pages with a single click. The API stores them in PostgreSQL. Workers (coming soon) will automatically extract tags, generate embeddings, and build connections between your saved knowledge.

---

## Quick Links

- [Getting Started](getting-started) -- Install and run Mesh
- [Browser Extension](browser-extension) -- Save and manage pages
- [System Tray](system-tray) -- Control services from your desktop
- [API Reference](api-reference) -- Use the REST API directly
- [Roadmap](roadmap) -- What's coming next
