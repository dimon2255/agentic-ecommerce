# ADR-0007: Admin Embedding Integration

## Status

Accepted

## Context

Product embeddings for the AI shopping assistant were generated via a standalone CLI tool (`cmd/embed/main.go`) that had to be run manually after products were created or updated. This required shell access and created a gap between catalog changes and search quality. The platform needed a way for admins to manage embeddings through the existing admin portal and for embeddings to stay fresh automatically.

## Options Considered

### Option A — Background Job System
- Pro: Scales to large catalogs, provides progress tracking, retry logic
- Con: Over-engineered for ~50-100 products, adds infrastructure complexity (job queue, worker)

### Option B — Admin API Endpoint + Auto-embed Hook
- Pro: Simple, no new infrastructure, reuses existing Voyage and Supabase clients
- Pro: Single-product regeneration completes in ~1s (synchronous), batch runs as goroutine
- Con: No built-in progress tracking for batch operations (uses server logs)

### Option C — Trigger-based (Database)
- Pro: Fully automatic, no application code changes
- Con: Requires Supabase Edge Functions or pg_cron, harder to debug, Voyage API key in DB layer

## Decision

Option B: Admin API endpoint with goroutine-based batch processing + auto-embed hook on product create/update.

- `POST /api/v1/admin/embeddings/regenerate` — batch re-embed all active products (goroutine, returns 202)
- `POST /api/v1/admin/embeddings/regenerate/{productId}` — single product (synchronous, returns 200)
- `GET /api/v1/admin/embeddings/status` — coverage counts
- Auto-hook: `CatalogHandler.onProductChange` fires a goroutine to re-embed after product create/update
- Admin UI: `pages/admin/embeddings.vue` with status cards, bulk regeneration, and per-product triggers
- `cmd/embed` CLI preserved as a convenience tool

## Consequences

- **Positive:** Embeddings stay fresh automatically on product changes. Admins can trigger bulk regeneration without CLI access. Simple architecture with no new infrastructure.
- **Negative:** Batch progress not visible in UI (only server logs). If Voyage API is down, auto-embed fails silently (logged, not retried).
- **Risks:** At larger catalog sizes (500+ products), batch Voyage API call may need chunking. The goroutine-based approach doesn't survive server restarts.
