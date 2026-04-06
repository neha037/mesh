---
layout: default
title: API Reference
nav_order: 5
---

# API Reference

The Mesh API runs at `http://localhost:8080`. All endpoints use JSON.

---

## Health Check

```
GET /healthz
```

Returns the health status of the API and its database connection.

**Request:**

```bash
curl http://localhost:8080/healthz
```

**Response (200 OK):**

```json
{
  "status": "ok"
}
```

**Response (503 Service Unavailable):**

```json
{
  "status": "unhealthy",
  "error": "database unreachable"
}
```

---

## Save a Page (with content)

```
POST /api/v1/ingest/raw
```

Saves a web page to the knowledge base with its content included in the request. If a page with the same URL already exists, it is updated instead.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/ingest/raw \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/article",
    "title": "Example Article",
    "content": "The full text content of the page...",
    "type": "article"
  }'
```

**Response (new page -- 201 Created):**

```json
{
  "id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
  "title": "Example Article",
  "created_at": "2026-04-02T12:43:09Z"
}
```

**Response (existing URL updated -- 200 OK):**

```json
{
  "id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
  "title": "Example Article",
  "created_at": "2026-04-02T12:43:09Z",
  "updated": true
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `url` | Yes | The page URL (must be http or https) |
| `title` | Yes | The page title |
| `content` | No | The extracted text content (max 500 KB, HTML-sanitized) |
| `type` | No | Node type (default: `article`). See [Node Types](#node-types) |

---

## Save a URL (async scraping)

```
POST /api/v1/ingest/url
```

Queues a URL for background scraping and processing. The page content is fetched asynchronously by a worker.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/ingest/url \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/article",
    "type": "article"
  }'
```

**Response (202 Accepted):**

```json
{
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "node_id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
  "status": "pending"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `url` | Yes | The page URL to scrape (must be http or https) |
| `type` | No | Node type (default: `article`). See [Node Types](#node-types) |

The worker will scrape the page in the background and update the node with the extracted content.

---

## Save Text (no URL)

```
POST /api/v1/ingest/text
```

Saves a text note or thought that doesn't have a URL. The content is queued for background processing.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/ingest/text \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My thought on distributed systems",
    "content": "Consistency and availability are fundamentally at odds...",
    "type": "thought"
  }'
```

**Response (201 Created):**

```json
{
  "node_id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
  "job_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "pending"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | The note title |
| `content` | Yes | The text content (HTML-sanitized) |
| `type` | No | Node type (default: `thought`). See [Node Types](#node-types) |

---

## List Recent Saves

```
GET /api/v1/nodes/recent
```

Returns the 20 most recently saved pages.

**Request:**

```bash
curl http://localhost:8080/api/v1/nodes/recent
```

**Response (200 OK):**

```json
[
  {
    "id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
    "type": "article",
    "title": "Example Article",
    "source_url": "https://example.com/article",
    "status": "active",
    "created_at": "2026-04-02T12:43:09Z"
  }
]
```

---

## List All Pages (Paginated)

```
GET /api/v1/nodes
```

Returns pages with cursor-based pagination.

**Parameters:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `per_page` | 20 | Number of results per page (1--100) |
| `cursor` | _(none)_ | Cursor string from `next_cursor` of previous response |

**Request (first page):**

```bash
curl "http://localhost:8080/api/v1/nodes?per_page=10"
```

**Response (200 OK):**

```json
{
  "nodes": [
    {
      "id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
      "type": "article",
      "title": "Example Article",
      "source_url": "https://example.com/article",
      "status": "active",
      "created_at": "2026-04-02T12:43:09Z"
    }
  ],
  "next_cursor": "eyJ0IjoiMjAyNi0wNC0wMlQxMjo0MzowOVoiLCJpIjoiNmFiZWZhYjYifQ==",
  "has_more": true
}
```

**Request (next page):**

```bash
curl "http://localhost:8080/api/v1/nodes?per_page=10&cursor=eyJ0IjoiMjAyNi0wNC0wMlQxMjo0MzowOVoiLCJpIjoiNmFiZWZhYjYifQ=="
```

When `has_more` is `false`, there are no more pages to load.

---

## Get a Page

```
GET /api/v1/nodes/{id}
```

Retrieves a single saved page by its ID.

**Request:**

```bash
curl http://localhost:8080/api/v1/nodes/6abefab6-e713-4ae7-8fbb-c0704ab33640
```

**Response (200 OK):**

```json
{
  "id": "6abefab6-e713-4ae7-8fbb-c0704ab33640",
  "type": "article",
  "title": "Example Article",
  "source_url": "https://example.com/article",
  "status": "active",
  "created_at": "2026-04-02T12:43:09Z"
}
```

**Response (404 Not Found):**

```json
{
  "error": "node not found"
}
```

---

## Delete a Page

```
DELETE /api/v1/nodes/{id}
```

Permanently deletes a saved page.

**Request:**

```bash
curl -X DELETE http://localhost:8080/api/v1/nodes/6abefab6-e713-4ae7-8fbb-c0704ab33640
```

**Response:** `204 No Content` (empty body on success)

**Response (404 Not Found):**

```json
{
  "error": "node not found"
}
```

---

## Node Types

All ingest endpoints accept an optional `type` field. Valid values:

| Type | Description |
|------|-------------|
| `article` | Web article or blog post (default for `/ingest/raw` and `/ingest/url`) |
| `book` | Book or long-form content |
| `hobby` | Hobby-related content |
| `thought` | Personal thought or note (default for `/ingest/text`) |
| `journal` | Journal entry |
| `image` | Image content |
| `wildcard` | Uncategorized or discovery content |

---

## Error Responses

All errors return a JSON object with an `error` field:

```json
{
  "error": "url is required"
}
```

| Status Code | Meaning |
|-------------|---------|
| 400 | Bad request (missing fields, invalid format) |
| 404 | Resource not found |
| 500 | Internal server error |
