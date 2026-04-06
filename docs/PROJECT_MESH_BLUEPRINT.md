# Project Mesh: Architectural Blueprint (Summary)

**Version:** 1.3  
**Status:** Phase 2 Complete (Intelligence & Processing)  
**Docs Ref:** [Implementation Roadmap](arch/roadmap.md) | [Data Model](arch/data_model.md) | [API Spec](arch/api_spec.md)

---

## 1. Executive Summary

Project Mesh is a **localized, private Personal Growth Engine** that maps both structured knowledge and fluid creative pursuits into a unified topological space. It acts as an **active cognitive partner**, surface connections using the "Adjacent Possible" concept.

### Design Constraints
- **Zero cloud costs** (Local compute/storage).
- **Absolute data sovereignty** (No data leaves the machine).
- **Scalable for single developers** (Sustainable 6-8 hours/week).

## 2. Technical Stack

| Category | Choice | Justification |
|----------|--------|---------------|
| **Backend** | Go (Golang) | Compiled binary, concurrency (goroutines), type safety. |
| **Database** | PostgreSQL 16 + pgvector | Recursive CTEs (graphs) + Vector similarity. |
| **Object Storage** | MinIO | S3-compatible local storage for images/PDFs. |
| **AI/NLP** | Ollama (Gemma 4) | High-quality local inference, structured JSON output. |
| **Frontend** | React + Cytoscape.js | Robust ecosystem + graph visualization algorithms. |

## 3. High-Level Architecture

Mesh uses an asynchronous, job-based ingestion pipeline:
1. **Ingestion**: URL or text submitted via API/Extension.
2. **Processing**: Worker claims job, scrapes content, uses LLM (via Ollama) to extract tags and embeddings.
3. **Graphing**: Nodes are connected via tags and semantic similarity.
4. **Discovery**: Engine analyzes clusters to suggest new links ("Bridges") or inject "Wildcards."

Detailed diagrams and specifications across modules:
- [Data Model & SQL Schema](arch/data_model.md)
- [API Spec & Interaction Flows](arch/api_spec.md)
- [Key Algorithms (FSRS, Kotkov)](arch/algorithms.md)

## 4. Security & Privacy
Mesh is designed for **maximum privacy**. All containers bind to `127.0.0.1`. No external APIs are used for inference. Data remains on local volumes.

Detailed analysis: [Risk Assessment & Mitigation](arch/risks.md)
