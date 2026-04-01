# Plan 5: Architecture Overhaul — Progress

## Status: In Progress

> Full audit of system design, backend architecture, and frontend UI/UX. 6 phases, 15 PRs, ~50 files.
> See `.claude/plans/jazzy-watching-yeti.md` for the detailed plan file.

## Phase Dependency Graph

```
Phase 1: Backend Infrastructure ──────────────────┐
    |                                              |
Phase 2: Service Layer (catalog, cart, checkout)   |
    |                                              |
Phase 3: Database + Performance                 Phase 4: Security
    |                                              |
Phase 6: Frontend Data Layer + Polish              |
                                                   |
Phase 5: Frontend Accessibility + Design System ───┘
         (independent — can run in parallel)

All 6 phases complete --> Plan 4: Admin Dashboard
```

## Tracking Table

| Phase | PR | Status | Branch | Date | Notes |
|-------|----|--------|--------|------|-------|
| **1** | **1A:** Error Types + Response Envelope + Request ID | Done | feat/plan-5-phase-1-infrastructure | 2026-03-31 | PR #5. `apperror`, `requestid` (own package to avoid import cycle), `response/json.go` |
| **1** | **1B:** Input Validation Framework | Done | feat/plan-5-phase-1-infrastructure | 2026-03-31 | PR #5. `validate` package, `Validate()` on 11 request structs |
| **2** | **2A:** Catalog Service + Repository | Done | feat/plan-5-phase-2-service-layer | 2026-03-31 | N+1 fixed via embedded selects. 5 handlers refactored |
| **2** | **2B:** Cart Service + Repository | Done | feat/plan-5-phase-2-service-layer | 2026-03-31 | Handler 418→90 lines. Auth error now 401 |
| **2** | **2C:** Checkout Service + Repository | Done | feat/plan-5-phase-2-service-layer | 2026-03-31 | Handler 275→100 lines. Price conflict now structured |
| **3** | **3A:** Indexes, Constraints, Soft Deletes | Done | feat/plan-5-phase-3-db-performance | 2026-03-31 | Migration 00005. Composite indexes, active cart uniqueness, order soft delete, shipping CHECK |
| **3** | **3B:** Atomic Operations via RPC | Done | feat/plan-5-phase-3-db-performance | 2026-03-31 | Migration 00006. Cart upsert RPC, atomic order creation, idempotency key |
| **3** | **3C:** Pagination + Search + Filtering | Done | feat/plan-5-phase-3-db-performance | 2026-03-31 | Migration 00007. FTS, pagination pkg, Supabase Ilike/Fts/CountExact, multi-category filter |
| **4** | **4A:** JWT + Security Headers + Stripe Fixes | Not Started | | | `iss`/`aud` validation, security headers, webhook replay, fix `In()` quoting |
| **4** | **4B:** Rate Limiting + Request Timeout | Not Started | | | Token bucket per IP, request timeout, context-aware Supabase client |
| **5** | **5A:** Accessibility (WCAG 2.1 AA) | Not Started | | | Skip-to-content, form labels, aria-labels, focus-visible, semantic breadcrumbs |
| **5** | **5B:** Design System Token Unification | Not Started | | | Semantic status colors, Tailwind safelist fix, single source of truth for tokens |
| **6** | **6A:** useApi Overhaul + Data Fetching | Not Started | | | Timeout, retry, error interceptors. Fix frontend N+1. Cart fetch-once. Polling timeout |
| **6** | **6B:** Skeleton Loaders + Toast + Optimistic Updates | Not Started | | | SkeletonCard, Toast system, per-item cart loading, optimistic updates |
| **6** | **6C:** Remaining UX Polish | Not Started | | | Lazy images, form validation UX, password toggle, breadcrumbs |

## Phase Details (Quick Reference)

### Phase 1: Backend Infrastructure
- **Goal:** Error types, structured response envelope, request IDs, validation framework
- **New packages:** `api/internal/apperror`, `api/internal/validate`, `api/internal/middleware/request_id`
- **Key change:** Responses become `{"data": ...}` / `{"error": {"code", "message", "request_id"}}`
- **Test:** `go test ./internal/apperror/... ./internal/validate/... ./internal/middleware/...`

### Phase 2: Service Layer + Repository Pattern
- **Goal:** Extract business logic from handlers into services, data access into repositories
- **Pattern:** Handler (thin HTTP adapter) → Service (validation + business logic) → Repository (data access)
- **Key fix:** N+1 queries eliminated via PostgREST embedded selects in repository layer
- **Wiring:** `main.go` builds: repo → service → handler
- **Test:** `go test ./internal/catalog/... ./internal/cart/... ./internal/checkout/...`

### Phase 3: Database Hardening + Performance
- **Goal:** Missing indexes, atomic operations via RPC, pagination, full-text search
- **Migrations:** `00005` (indexes/constraints), `00006` (RPC functions), `00007` (search)
- **Key fix:** Cart and checkout race conditions eliminated via stored procedures
- **Test:** `supabase db reset && go test ./...`

### Phase 4: Security Hardening
- **Goal:** JWT claim validation, security headers, rate limiting, request timeouts
- **Key fix:** `iss`/`aud` JWT validation, webhook replay protection, Supabase `In()` quoting fix
- **Test:** `curl -I` for headers, rapid requests for rate limiting, expired JWT tokens

### Phase 5: Frontend Accessibility + Design System (parallel with backend)
- **Goal:** WCAG 2.1 AA compliance, unified design tokens
- **Key fixes:** Skip-to-content, form label association, aria-labels, focus-visible, semantic colors
- **Test:** Keyboard-only navigation, screen reader, color contrast checker

### Phase 6: Frontend Data Layer + UX Polish (depends on Phase 3C)
- **Goal:** useApi overhaul, skeleton loaders, toast notifications, optimistic updates, form validation
- **Key fixes:** Frontend N+1 → single API call, cart fetch-once, per-item loading states
- **Test:** `npm run dev` — full user flow verification

## Deferred Items

| Item | Target Plan |
|------|-------------|
| Bulk operations | Plan 4 (Admin Dashboard) |
| Product soft deletes | Plan 4 (Admin Dashboard) |
| Inventory management | Plan 5+ |
| Audit trail tables | Plan 5+ |
| Page views TTL/partitioning | Plan 5+ |
| Images text[] → JSONB | Plan 5+ |
