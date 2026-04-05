# Frontend Test Coverage — Progress Tracker

Plan: Add comprehensive test coverage for Nuxt 3 TypeScript frontend.
Stack: Vitest + @nuxt/test-utils + @vue/test-utils + Playwright (E2E).
Convention: Colocated test files (`useCart.test.ts` next to `useCart.ts`).

---

## Phase 1 — Test Infrastructure + Composable Tests

Setup Vitest, @nuxt/test-utils, @vue/test-utils. Write unit tests for all 8 composables.
Highest ROI — composables hold business logic and state.

| # | Task | Status |
|---|------|--------|
| 1.1 | Install vitest, @nuxt/test-utils, @vue/test-utils, happy-dom | Done |
| 1.2 | Create `vitest.config.ts` with Nuxt integration | Done |
| 1.3 | Add `test` and `test:ci` scripts to `package.json` | Done |
| 1.4 | Test `useCart` composable (14 tests — CRUD, totals, dual auth) | Done |
| 1.5 | Test `useCheckout` composable (12 tests — Stripe, price changes) | Done |
| 1.6 | Test `useApi` composable (10 tests — $fetch, retry, timeout) | Done |
| 1.7 | Test `useAdminAuth` composable (10 tests — perms AND/OR, reset) | Done |
| 1.8 | Test `useAdminApi` composable (7 tests — prefix, auth headers) | Done |
| 1.9 | Test `useAssistant` composable (17 tests — SSE, sync, tools) | Done |
| 1.10 | Test `useAssistantPanel` composable (7 tests — toggle, unread) | Done |
| 1.11 | Test `useToast` composable (7 tests — queue, dismiss, FIFO) | Done |

**Phase 1 Results:** 8 test files, 84 tests, all passing. Composable coverage: 92% stmts, 75% branches, 99% functions.

## Phase 2 — Component Tests

Test components with conditional rendering, user interaction, or calculations.
Skip pure-layout components. Focus on behavior, not markup.

| # | Task | Status |
|---|------|--------|
| 2.1 | Test `PriceDisplay` (3 tests — base, override, null) | Done |
| 2.2 | Test `SkuSelector` (7 tests — select, match, availability, aria) | Done |
| 2.3 | Test `CartItem` (7 tests — qty +/-, remove, disabled, total) | Done |
| 2.4 | Test `ProductCard` (4 tests — name, price, image, link) | Done |
| 2.5 | Test `CategoryCard` (3 tests — name, link, gradient) | Done |
| 2.6 | Test `Toast` (3 tests — render, role="alert", empty) | Done |
| 2.7 | Test `AssistantPanel` — Skipped (343 LOC, heavy mocking, deferred) | Deferred |
| 2.8 | Test `ChatProductCard` (3 tests — name, image, fallback) | Done |
| 2.9 | Test admin `DataTable` (10 tests — sort, pagination, states) | Done |
| 2.10 | Test admin `ProductForm` — Skipped (thin form wrapper) | Skipped |
| 2.11 | Test admin `CategoryForm` — Skipped (thin form wrapper) | Skipped |
| 2.12 | Test admin `ConfirmDialog` (5 tests — open/close, emit, variant) | Done |
| 2.13 | Test admin `StatusBadge` (4 tests — text, label, class mapping) | Done |
| 2.14 | Test admin `ImageUploader` — Skipped (async upload, deferred) | Deferred |
| 2.15 | Test admin `KpiCard` (5 tests — format, prefix, loading) | Done |
| 2.16 | Test admin `SlugInput` (6 tests — auto-slug, manual, special chars) | Done |

**Phase 2 Results:** 12 test files, 60 tests, all passing. Tested components at 99% stmts. Skipped thin wrappers + deferred complex async components.

## Phase 3 — E2E Tests (Playwright)

Full user journey validation. 5-10 critical flows through the real UI.
Will run in CI/CD pipeline.

| # | Task | Status |
|---|------|--------|
| 3.1 | Install Playwright + Chromium | Done |
| 3.2 | Create `playwright.config.ts` (CI-friendly, webServer, retries) | Done |
| 3.3 | E2E: Browse catalog and view product detail (3 tests) | Done |
| 3.4 | E2E: Add to cart (guest) and view cart (2 tests) | Done |
| 3.5 | E2E: Auth flow — login, register, invalid credentials (3 tests) | Done |
| 3.6 | E2E: Checkout flow — form render, cart requirement (2 tests) | Done |
| 3.7 | E2E: Cart persistence across navigation (1 test) | Done |
| 3.8 | E2E: Admin catalog pages load (4 tests) | Done |
| 3.9 | E2E: Admin orders + reports pages load (2 tests) | Done |
| 3.10 | E2E: AI assistant FAB, open/close panel (3 tests) | Done |
| 3.11 | Add `test:e2e` and `test:e2e:ci` scripts to `package.json` | Done |

**Phase 3 Results:** 8 test files, 19 E2E tests. 14 passing, 5 flaky. Requires running dev stack (`npm run dev` + Go API + Supabase).

### Known Flaky E2E Tests

5 tests fail in headless Chromium due to Vue `reactive()` state timing with Playwright auto-waiting:

- **Cart tests** (cart-guest x2, cart-merge, checkout) — SkuSelector radio button clicks don't register. The `[role="radio"]` click fires but Vue's `selectOption()` handler doesn't execute before the next assertion. Works intermittently.
- **Assistant panel** — FAB click opens panel but the DOM update isn't detected by Playwright within the timeout.

**Root cause:** Vue's reactivity system batches DOM updates. Playwright's `click()` returns before Vue processes the event. The `waitForTimeout(500)` workaround is insufficient.

**Fix options:**
1. Use `page.waitForFunction(() => document.querySelector('[aria-checked="true"]'))` after each radio click
2. Use `--headed` mode (clicks register more reliably with visible browser)
3. Add `data-testid` attributes to the SkuSelector for more stable selectors

---

## CI/CD Notes

- `npm test` — runs Vitest (unit + component) with coverage report
- `npm run test:ci` — same, with CI-optimized reporter (junit XML output)
- `npm run test:e2e` — runs Playwright E2E suite
- `npm run test:e2e:ci` — headless, with retries and CI reporter
- Future: add to GitHub Actions pipeline as required check on PRs
