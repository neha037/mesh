---
layout: default
title: API Reference
nav_order: 5
---

# API Reference

The Mesh API runs at `http://localhost:8080`. All endpoints use JSON.

---

## Save a Page

```
POST /api/v1/ingest/raw
```

Saves a web page to the knowledge base. If a page with the same URL already exists, it is updated instead.

**Request:**

```bash
curl -X POST http://localhost:8080/api/v1/ingest/raw \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/article",
    "title": "Example Article",
    "content": "The full text content of the page..."
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
| `url` | Yes | The page URL |
| `title` | Yes | The page title |
| `content` | No | The extracted text content |

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
    "title": "Example Article",
    "source_url": "https://example.com/article",
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
| `per_page` | 20 | Number of results per page (max 100) |
| `cursor` | _(none)_ | ISO 8601 timestamp from `next_cursor` of previous response |

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
      "title": "Example Article",
      "source_url": "https://example.com/article",
      "created_at": "2026-04-02T12:43:09Z"
    }
  ],
  "next_cursor": "2026-04-02T12:43:09Z",
  "has_more": true
}
```

**Request (next page):**

```bash
curl "http://localhost:8080/api/v1/nodes?per_page=10&cursor=2026-04-02T12:43:09Z"
```

When `has_more` is `false`, there are no more pages to load.

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
| 500 | Internal server error |
