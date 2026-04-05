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
| 2.1 | Test `PriceDisplay` (formatting, currency, sale prices) | Pending |
| 2.2 | Test `SkuSelector` (variant selection, availability, price update) | Pending |
| 2.3 | Test `CartItem` (quantity change, remove, price calculation) | Pending |
| 2.4 | Test `ProductCard` (render, link, image fallback) | Pending |
| 2.5 | Test `CategoryCard` (render, navigation) | Pending |
| 2.6 | Test `Toast` (show/dismiss, types) | Pending |
| 2.7 | Test `AssistantPanel` (open/close, message rendering) | Pending |
| 2.8 | Test `ChatProductCard` (product data display) | Pending |
| 2.9 | Test admin `DataTable` (sorting, pagination, empty state) | Pending |
| 2.10 | Test admin `ProductForm` (validation, submit, image upload) | Pending |
| 2.11 | Test admin `CategoryForm` (validation, submit) | Pending |
| 2.12 | Test admin `ConfirmDialog` (confirm/cancel actions) | Pending |
| 2.13 | Test admin `StatusBadge` (status variants rendering) | Pending |
| 2.14 | Test admin `ImageUploader` (file select, preview, upload) | Pending |

## Phase 3 — E2E Tests (Playwright)

Full user journey validation. 5-10 critical flows through the real UI.
Will run in CI/CD pipeline.

| # | Task | Status |
|---|------|--------|
| 3.1 | Install Playwright and configure for Nuxt | Pending |
| 3.2 | Create `playwright.config.ts` with CI-friendly settings | Pending |
| 3.3 | E2E: Browse catalog and view product detail | Pending |
| 3.4 | E2E: Add to cart (guest) and view cart | Pending |
| 3.5 | E2E: Sign up / sign in flow | Pending |
| 3.6 | E2E: Checkout flow (cart to payment to order confirmation) | Pending |
| 3.7 | E2E: Cart merge on login (guest cart + auth cart) | Pending |
| 3.8 | E2E: Admin login and catalog CRUD | Pending |
| 3.9 | E2E: Admin order management | Pending |
| 3.10 | E2E: AI assistant interaction | Pending |
| 3.11 | Add `test:e2e` and `test:e2e:ci` scripts to `package.json` | Pending |

---

## CI/CD Notes

- `npm test` — runs Vitest (unit + component) with coverage report
- `npm run test:ci` — same, with CI-optimized reporter (junit XML output)
- `npm run test:e2e` — runs Playwright E2E suite
- `npm run test:e2e:ci` — headless, with retries and CI reporter
- Future: add to GitHub Actions pipeline as required check on PRs
