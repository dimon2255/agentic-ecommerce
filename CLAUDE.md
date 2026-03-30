# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Single-tenant e-commerce platform: Nuxt 3 frontend, Go API (Chi router), Supabase PostgreSQL with RLS, Stripe payments. No Docker or CI/CD — all services run locally.

## Commands

### Go API (from `api/`)
```bash
go run ./cmd/server/          # Start API on :9090
go test ./...                 # Run all tests
go test ./internal/cart/...   # Run tests for one package
go test -run TestName ./...   # Run a single test by name
```

### Frontend (from `frontend/`)
```bash
npm install                   # Install dependencies
npm run dev                   # Start dev server on :3000
npm run build                 # Production build
```

### Supabase
```bash
supabase start                # Start local instance (ports: API 54321, DB 54322, Studio 54323)
supabase db reset             # Reset DB and apply seed + migrations
```

### Stripe
```bash
stripe listen --forward-to localhost:9090/stripe/webhook   # Forward webhooks locally
```

## Architecture

```
Nuxt 3 (SSR/SPA) --HTTP+JWT--> Go Chi API (:9090) --PostgREST--> Supabase PostgreSQL
                                     |
                                     +--> Stripe API (PaymentIntent + webhooks)
```

**Data flow:** Frontend uses Supabase Auth for login, gets JWT. Go API validates JWT via middleware, talks to Supabase using service role key for all mutations. Stripe webhooks hit `/stripe/webhook` (outside `/api/v1`, no auth, signature-verified).

**Cart identity:** Guest carts use `X-Session-ID` header; authenticated carts use JWT user ID. Merging happens at login via `POST /cart/merge`.

### Go API structure (`api/`)
- `cmd/server/main.go` — entry point, route wiring
- `internal/config/` — Viper-based config (YAML + env vars prefixed `ESHOP_`)
- `internal/catalog/` — category, product, SKU, attribute, custom-field handlers
- `internal/cart/` — cart CRUD handlers
- `internal/checkout/` — order creation, Stripe PaymentIntent, webhook handler
- `internal/middleware/` — JWT auth (required + optional variants)
- `pkg/supabase/` — PostgREST HTTP client
- `pkg/stripe/` — Stripe client wrapper
- `pkg/response/` — JSON response helpers

### Frontend structure (`frontend/`)
- `pages/` — file-based routing (catalog, product, cart, checkout, order, auth)
- `composables/` — `useApi` (HTTP client), `useCart` (cart state), `useCheckout` (payment flow)
- `components/` — ProductCard, SkuSelector, CartItem, PriceDisplay, CategoryCard

### Database (`supabase/`)
- Migrations in `supabase/migrations/` (catalog, cart, orders — each with RLS policies)
- Seed data in `supabase/seed.sql`

## Configuration

API config loads via Viper: defaults in `api/config.yaml`, overridden by env vars.

**Required env vars** (no defaults — must be set):
- `ESHOP_SUPABASE_SERVICE_ROLE_KEY`
- `ESHOP_SUPABASE_JWT_SECRET`
- `ESHOP_STRIPE_SECRET_KEY`
- `ESHOP_STRIPE_WEBHOOK_SECRET`

Frontend env vars: `NUXT_PUBLIC_API_BASE`, `NUXT_PUBLIC_SUPABASE_URL`, `NUXT_PUBLIC_SUPABASE_KEY`

See `.env.example` for the full list.

## Testing Patterns

Go tests use `httptest` servers to mock Supabase PostgREST responses. Pattern:
1. Create `httptest.NewServer` with mock handler matching expected PostgREST queries
2. Build handler with mock Supabase client pointing to test server
3. Use `httptest.NewRequest` / `httptest.NewRecorder` to test handlers
4. Inject user ID and route params via context

## Key Design Decisions

- **Service role key pattern:** Go API bypasses RLS using service role key; frontend anon key gets public read only
- **Price snapshots:** Cart items and order items store `unit_price` at creation time
- **SSR routing:** SSR enabled for catalog pages, disabled for cart/checkout/auth/account (`nuxt.config.ts` `routeRules`)
- **Stripe flow:** PaymentIntent API with client-side confirmation (supports 3D Secure)
