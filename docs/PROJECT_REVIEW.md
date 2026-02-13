# Recipes Parser Project Review

## What the project does

This repository implements an asynchronous recipe ingestion API focused on Reddit:

1. A client calls `POST /api/recipes` with a Reddit URL and optional limit.
2. The API attempts to return existing data from Redis cache first.
3. If cache misses, it looks in MongoDB.
4. If still missing, the API publishes a job to RabbitMQ and responds `202 processing`.
5. A worker consumes the job, fetches Reddit JSON, filters/normalizes recipe posts, persists results to MongoDB, and warms Redis.
6. The client can poll until a `200 found` response is returned.

## Architecture summary

The codebase follows a layered/ports-and-adapters structure:

- **Domain**: entities (`Recipe`), value objects (`Job`), and interfaces (`Cache`, `DataBase`, `Queue`, `JobGuard`).
- **Application**: `RecipeUseCase` orchestrates cache/database reads and async job enqueueing.
- **Infrastructure**:
  - Redis cache + distributed lock (`JobGuard`).
  - Mongo repository with upsert behavior.
  - RabbitMQ publisher/consumer.
  - Reddit parser adapter.
- **Delivery**: Gin HTTP handler + router.
- **Entrypoints**:
  - `cmd/api/main.go` (synchronous query + async dispatch API).
  - `cmd/worker/main.go` (job consumer + parser pipeline).

## End-to-end request flow

### API path

- Request validated by Gin binding (`url` required and URL-typed).
- Default `limit=10` when omitted.
- Use case validates business constraints (`1 <= limit <= 100`).
- Cache key is deterministic: SHA-1 hash of URL + limit.
- Cache hit => `FOUND` with recipes.
- DB hit => returns recipes and backfills cache.
- Full miss => distributed lock attempt:
  - lock acquired => enqueue RabbitMQ job and return `PROCESSING`.
  - lock not acquired => another worker/API already processing, return `PROCESSING`.

### Worker path

- RabbitMQ consumer deserializes `Job`.
- Worker fetches recipes from Reddit parser.
- Error handling policy:
  - 404/invalid URL => permanent, ACK + persist empty list to avoid endless retries.
  - 429/rate limit => retry via NACK requeue.
  - unknown errors => retry via NACK requeue.
- Success => save to Mongo, best-effort set cache.
- Always attempts lock release using token-based Lua compare-delete script.

## Current strengths

- Clean separation between use-case orchestration and adapters.
- Good async API pattern for potentially slow external scraping/parsing.
- Sensible cache-aside strategy with DB fallback.
- Tokenized lock release avoids deleting another process lock accidentally.
- RabbitMQ durable queue and persistent publish mode are configured.

## Key improvement opportunities

### 1) Configuration management (highest priority)

**Issue:** credentials/hosts/TTLs are hardcoded in both API and worker mains.

**Recommendation:** centralize config via env vars + typed config struct + validation (e.g., `envconfig`, `koanf`, or plain stdlib).

**Impact:** safer deployments, easier local/prod parity, no secrets in source.

### 2) Lock key collision between cache and job lock

**Issue:** `JobGuard` currently uses the same key namespace as cached data (`recipes:...`). If lock acquisition writes a string value to a key expected to hold JSON array, it can corrupt cache semantics.

**Recommendation:** isolate lock keys, e.g. `lock:recipes:<hash>:<limit>` and cache keys `cache:recipes:<hash>:<limit>`.

**Impact:** prevents type conflicts and subtle race bugs.

### 3) HTTP error semantics

**Issue:** handler maps all use-case errors to `500`; user-input validation errors from use-case become internal-server responses.

**Recommendation:** return `400` for validation/domain input errors and reserve `500` for infrastructure failures.

**Impact:** clearer client contract and easier observability.

### 4) Parser robustness

**Issues:**
- URL mutation via string formatting risks malformed query strings (`...json?limit=` blindly appended).
- image filtering based only on file extensions can miss CDN URLs without extensions.
- fixed user-agent placeholder should be configurable.

**Recommendations:**
- Build URLs with `net/url` and query merging.
- Add lightweight content-type probing/known-host heuristics.
- Externalize user-agent and timeout config.

### 5) Queue reliability controls

**Issue:** retry strategy currently unbounded with immediate requeue for transient errors.

**Recommendation:** add retry count/dead-letter exchange and exponential backoff queues.

**Impact:** avoids poison-message loops and stabilizes worker throughput.

### 6) Observability

**Issue:** logs are mostly plain text without job correlation IDs/structured fields.

**Recommendation:** structured logger (`zap`/`zerolog`) + request/job IDs + metrics (queue depth, success/failure rates, parser latency, cache hit ratio).

### 7) Testing coverage

**Issue:** no unit/integration tests currently.

**Recommendation:**
- Unit tests for use case branching and key generation.
- Contract tests for parser JSON mapping.
- Integration tests with testcontainers for Redis/Mongo/RabbitMQ.

### 8) API ergonomics

**Recommendation:**
- expose `GET /api/recipes?url=&limit=` as idempotent retrieval endpoint.
- include `job_key` in processing response to support explicit polling.
- standardize response envelope + error codes.

### 9) Security and production hardening

**Recommendation:**
- add request rate limiting and payload size limits.
- sanitize/allowlist Reddit URL path patterns more strictly.
- add context cancellation propagation and graceful shutdown for worker/API.

## Suggested implementation roadmap

1. Config/env migration + secret removal.
2. Separate lock/cache key namespaces.
3. Error mapping improvements in HTTP layer.
4. Add foundational tests for use case and parser.
5. Add dead-letter/retry policy in RabbitMQ.
6. Add structured logs + health endpoints + metrics.

## Quick start hints

- Bring up dependencies with Docker Compose.
- Run API and worker in separate processes.
- Submit a Reddit recipes URL to `/api/recipes` and poll until status changes from `processing` to `found`.
