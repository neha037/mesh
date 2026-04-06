# Mesh API Endpoint Specification

This document details the REST API endpoints for the Mesh project.

## Ingestion Endpoints

| Method | Path | Description | Request Body |
|--------|------|-------------|-------------|
| `POST` | `/api/v1/ingest/url` | Submit a URL for scraping and processing | `{ "url": "https://...", "type": "article" }` |
| `POST` | `/api/v1/ingest/text` | Submit raw text (thought, note) | `{ "title": "...", "content": "...", "type": "thought" }` |
| `POST` | `/api/v1/ingest/image` | Upload an image file | `multipart/form-data` with image + metadata JSON |
| `POST` | `/api/v1/ingest/journal` | Submit a journal entry | `{ "content": "...", "mood": "..." }` |

## Query Endpoints

| Method | Path | Description | Query Params |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/graph` | Full graph or subgraph | `?center=<uuid>&depth=<int>&types=<csv>` |
| `GET` | `/api/v1/nodes` | List nodes with pagination | `?page=&limit=&type=&tag=&from=&to=` |
| `GET` | `/api/v1/nodes/:id` | Get single node with edges | — |
| `GET` | `/api/v1/nodes/:id/similar` | Semantic similarity search | `?limit=<int>&threshold=<float>` |
| `GET` | `/api/v1/search` | Full-text + semantic search | `?q=<text>&mode=text|semantic|hybrid` |
| `GET` | `/api/v1/tags` | List all tags with counts | `?sort=count|alpha` |
| `GET` | `/api/v1/clusters` | Cluster density report | — |

## FSRS / Review Endpoints

| Method | Path | Description | Request Body |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/review/today` | Get today's review node | — |
| `POST` | `/api/v1/review/:node_id` | Submit review rating | `{ "rating": 1-4 }` |
| `GET` | `/api/v1/review/stats` | Review history and metrics | — |

## Discovery Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/discovery/bridges` | Get suggested bridge readings |
| `GET` | `/api/v1/discovery/wildcards` | Get recent wildcard injections |
| `POST` | `/api/v1/discovery/trigger` | Manually trigger discovery run |

## Node CRUD

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/api/v1/nodes/:id` | Update node (title, content, tags) |
| `DELETE` | `/api/v1/nodes/:id` | Delete node and associated edges |
| `POST` | `/api/v1/nodes/:id/edges` | Manually create an edge |
| `DELETE` | `/api/v1/edges/:id` | Delete an edge |
