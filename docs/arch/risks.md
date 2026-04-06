# Mesh: Risk Assessment and Mitigation

## 1. Local Hardware Constraints (RAM/CPU)
- **Problem**: Running heavy LLMs locally can strain systems.
- **Mitigation**: 
  - Use quantized models (Gemma 4 E4B fits in 6GB).
  - Use Docker profiles to start Ollama only when needed.
  - Implement memory limits in Docker.

## 2. Scraping Failures
- **Problem**: Targets may block requests or change structures.
- **Mitigation**:
  - Circuit breakers (`gobreaker`) to prevent runaway failures.
  - User-Agent rotation and respect for `robots.txt`.
  - Exponential backoff for job retries.

## 3. Data Privacy and Sovereignty
- **Problem**: Knowledge graph data is highly sensitive.
- **Mitigation**:
  - **Zero data leaves the machine**. All inference is local via Ollama.
  - No cloud telemetry or third-party tracking.
  - All services bind to `127.0.0.1` only.

## 4. Concurrency and Race Conditions
- **Problem**: Multiple workers accessing the same resources.
- **Mitigation**:
  - PostgreSQL Atomic UPSERTs (`ON CONFLICT`).
  - `FOR UPDATE SKIP LOCKED` for job queue processing.
  - Optimistic Concurrency Control (`version` column) on nodes.

## 5. Scope Creep
- **Problem**: Side projects often stall due to over-ambition.
- **Mitigation**:
  - Modular phases where each delivers standalone value.
  - MVP-first focus (Phases 1-4).
  - No hard deadlines; sustainable 6-8 hours/week pace.
