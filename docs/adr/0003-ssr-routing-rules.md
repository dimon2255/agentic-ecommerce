# ADR-0003: SSR for Catalog Pages, SPA for Cart/Checkout/Auth

## Status

Accepted

## Context

Nuxt 3 supports hybrid rendering — SSR, SPA, and static generation can be configured per route. We need to decide which pages are server-rendered for SEO and performance, and which run client-only for interactivity and auth-gated content.

## Options Considered

### Option A — Full SSR everywhere
- Pro: Consistent rendering model, best SEO across all pages
- Con: Cart/checkout/auth pages don't benefit from SEO; SSR adds complexity for auth-gated content and client-side state (Stripe, session IDs)

### Option B — Full SPA (disable SSR)
- Pro: Simplest setup, no hydration issues
- Con: Catalog and product pages lose SEO benefits and have slower first paint

### Option C — Hybrid: SSR for public catalog, SPA for authenticated/interactive routes
- Pro: SEO where it matters (product pages, categories), simpler client-side logic where auth and payment state dominate
- Con: Two rendering modes to reason about; must be careful with composables that assume one mode

## Decision

Use **Option C** via `routeRules` in `nuxt.config.ts`. SSR is enabled for catalog and product pages. SPA mode (`ssr: false`) is used for `/cart`, `/checkout`, `/auth`, `/account`, and `/admin` routes.

## Consequences

- **Positive:** Product and category pages are SEO-friendly with fast server-rendered first paint
- **Positive:** Cart/checkout/auth avoid SSR hydration issues with Stripe, session state, and Supabase auth
- **Negative:** Developers must be aware of which mode a page runs in when using composables
- **Risks:** Composables that use browser-only APIs (localStorage, window) must be guarded on SSR pages
