# Plan 5: Architecture Overhaul — Progress

## Status: Complete

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
| **4** | **4A:** JWT + Security Headers + Stripe Fixes | Done | feat/plan-5-phase-4-security | 2026-03-31 | iss/aud validation, CSP, HSTS, webhook replay guard, In() fix, error sanitization |
| **4** | **4B:** Rate Limiting + Request Timeout | Done | feat/plan-5-phase-4-security | 2026-03-31 | Token bucket per IP (100/20/50), 30s timeout, config extensions |
| **5** | **5A:** Accessibility (WCAG 2.1 AA) | Done | feat/plan-5-phase-5-frontend-a11y | 2026-03-31 | Skip-to-content, form labels, aria-labels, focus-visible, semantic breadcrumbs, aria-live |
| **5** | **5B:** Design System Token Unification | Done | feat/plan-5-phase-5-frontend-a11y | 2026-03-31 | 12 semantic status CSS vars, Tailwind safelist, Stripe CSS var refs, hardcoded colors replaced |
| **6** | **6A:** useApi Overhaul + Data Fetching | Done | feat/plan-5-phase-6-frontend-polish | 2026-03-31 | 30s timeout, auto-retry, N+1 fix via category_ids, cart fetch-once, polling timeout |
| **6** | **6B:** Skeleton Loaders + Toast + Optimistic Updates | Done | feat/plan-5-phase-6-frontend-polish | 2026-03-31 | SkeletonCard, Toast system, per-item Set loading, optimistic qty updates |
| **6** | **6C:** Remaining UX Polish | Done | feat/plan-5-phase-6-frontend-polish | 2026-03-31 | Lazy images on 3 components, password toggle on login+register, real-time match feedback |

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

---

# Plan 6: AI Shopping Assistant — Progress

## Status: Phase 2 Complete — verified end-to-end 2026-04-02

> Rufus-style conversational AI assistant. Anthropic Claude + Voyage AI embeddings + pgvector RAG.
> See `.claude/plans/validated-crafting-minsky.md` for the Phase 1 detailed plan file.

## Phase Overview

```
Phase 1: Data Foundation + Proof of Life ✓
    |
Phase 2: Tool Use + SSE Streaming ✓
    |
Phase 3: Chat UX + Product Cards ✓
    |
Phase 4: Auth Guards + Rate Limiting + Observability ✓
    |
Phase 5: Personalization + Analytics (future/V2)
```

## Tracking Table

| Phase | PR | Status | Branch | Date | Notes |
|-------|----|--------|--------|------|-------|
| **1** | **1A:** DB Migration + Config | Done | feat/ai-assistant-phase1-db | 2026-04-01 | pgvector, product_embeddings, chat_sessions, chat_messages, match_products RPC, AssistantConfig |
| **1** | **1B:** Embedding Pipeline + Chat Backend | Done | feat/ai-assistant-phase1-db | 2026-04-01 | Voyage AI client, Anthropic client, internal/assistant/ domain, cmd/embed CLI, wiring in main.go |
| **1** | **1C:** Frontend Proof of Life | Done | feat/ai-assistant-phase1-db | 2026-04-01 | useAssistant composable, pages/assistant.vue, nav link |
| **2** | **2A:** Tool Definitions + Execution Loop | Done | feat/ai-assistant-phase2-tools | 2026-04-02 | PR #14. 5 tools → catalog/cart services, agentic loop (max 5 iter), CompleteWithTools, /tools route, 17 tests |
| **2** | **2B:** SSE Streaming + Conversation Persistence | Done | feat/ai-assistant-phase2-tools | 2026-04-02 | PR #14. StreamWithTools, SSE handler, timeout bypass, history helper, frontend fetch+ReadableStream, 29 tests |
| **3** | **3A:** Slide-over Panel + FAB | Done | feat/ai-assistant-phase3-chat-ux | 2026-04-04 | PR #15. 420px panel, FAB trigger, mobile overlay, focus trap, a11y baked in |
| **3** | **3B:** Product Cards + Suggestion Chips | Done | feat/ai-assistant-phase3-chat-ux | 2026-04-04 | PR #15. tool_result SSE event, ChatProductCard, ChatMessage, suggestion chips |
| **4** | **4A:** Config + slog + Rate Limiting | Done | feat/ai-assistant-phase4-guards-observability | 2026-04-04 | Structured JSON logging (slog), per-user/guest sliding-window rate limiter, config extensions |
| **4** | **4B:** Guest Mode + System Prompt Polish | Done | feat/ai-assistant-phase4-guards-observability | 2026-04-04 | GuestTools (no cart), OptionalAuth, frontend auth gate removed, hardened system prompt |
| **4** | **4C:** Cost Tracking + Circuit Breaker | Done | feat/ai-assistant-phase4-guards-observability | 2026-04-04 | Token usage capture, chat_token_usage table, daily budget, circuit breaker on Anthropic client |
| **4** | **4D:** OpenTelemetry + Azure Monitor | Done | feat/ai-assistant-phase4-guards-observability | 2026-04-04 | pkg/telemetry, OTLP exporter, otelhttp middleware, spans on Anthropic/Supabase/Voyage, traced slog, graceful shutdown |
| **5** | TBD | Pending | — | — | Order history-aware recs, conversation analytics, A/B test prompts |

## Phase 1 Details

**Goal:** Prove end-to-end RAG chat works — user asks a product question, vector search retrieves relevant products, Claude generates a product-aware answer.

**Architecture:**
- **Embedding model:** Voyage AI `voyage-3-large` (1024 dims)
- **Chat model:** Anthropic `claude-sonnet-4-5` (non-streaming for Phase 1)
- **Vector store:** pgvector in Supabase PostgreSQL (HNSW index)
- **New Go package:** `internal/assistant/` (handler → service → repository)
- **New clients:** `pkg/voyage/`, `pkg/anthropic/`
- **Frontend:** Minimal `pages/assistant.vue` + `useAssistant()` composable

**Key files:**
- `supabase/migrations/00008_ai_assistant.sql` — pgvector + tables + RPC
- `api/internal/config/config.go` — AssistantConfig (API keys, model names)
- `api/internal/assistant/` — handler, service, repository, embedding, prompts
- `api/pkg/voyage/client.go` — embedding client
- `api/pkg/anthropic/client.go` — chat completion client
- `api/cmd/embed/main.go` — batch embedding CLI
- `frontend/composables/useAssistant.ts` — chat state
- `frontend/pages/assistant.vue` — chat page

**NOT in Phase 1:** SSE streaming, tool use, slide-over panel, product cards, guest mode, rate limiting

**Bugs fixed during Phase 1:**
- Auth middleware only accepted HS256; Supabase now issues ES256 JWTs → added JWKS-based ES256 verification
- `useSupabaseSession()` / `client.auth.getSession()` not resolving token → dual-path auth header resolution
- Model ID `claude-sonnet-4-6-20250514` not found → corrected to `claude-sonnet-4-5` (no date suffix)
- JWKS URL not derived when `JWTIssuer` config empty → fallback to Supabase URL

### Phase 2: Tool Use + SSE Streaming
- **Goal:** Make the assistant agentic — Claude calls tools to search products, view details, manage the cart — and streams responses via SSE for real-time UX.
- **5 tools:** `search_products`, `get_product_details`, `get_categories`, `get_cart`, `add_to_cart` — all dispatch to existing catalog/cart service methods
- **Agentic loop:** max 5 iterations, 2048 max tokens, tool results not persisted to DB
- **SSE streaming:** `StreamWithTools()` parses Anthropic SSE events, `POST /assistant/stream` endpoint
- **Conversation history:** `buildConversationMessages()` caps at 20, enforces role alternation
- **Timeout restructure:** moved from global to per-group so SSE bypasses `http.TimeoutHandler`
- **Frontend:** `fetch()` + `ReadableStream`, message status states, markdown rendering via `marked`
- **Test:** `go test ./...` (29 tests), `npm run build`

**Bugs fixed during Phase 2:**
- SSE `eventType` lost across TCP chunk boundaries → moved variable outside while loop
- `tool_use.input` field omitted when nil/empty → `len(block.Input) == 0` check + fallback to `{}`
- Claude passing `sku_code` instead of UUID `id` for `add_to_cart` → clarified tool description + system prompt
- Cart creation fails with empty sessionID → generate deterministic `assistant-{userID}` session
- Streaming cursor persists after completion → force `status: 'complete'` in finally block
- `http.TimeoutHandler` kills `http.Flusher` → timeout middleware per-group, not global
