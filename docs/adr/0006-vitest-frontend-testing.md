# ADR-0006: Vitest + @nuxt/test-utils for Frontend Test Coverage

## Status

Accepted

## Context

The Nuxt 3 frontend has 8 composables and 21 components with zero test coverage. We need a test framework that integrates well with Nuxt's Vite-based build system, handles auto-imports, and supports a path to CI/CD integration.

## Options Considered

### Option A — Jest
- Pro: Largest ecosystem, most community resources
- Con: Requires additional ESM/Vite configuration; not native to the Vite toolchain
- Con: Slower test execution due to transform overhead

### Option B — Vitest + @nuxt/test-utils + @vue/test-utils
- Pro: Native Vite integration — same transform pipeline as the dev server
- Pro: @nuxt/test-utils provides `mountSuspended()` which handles Nuxt auto-imports and context
- Pro: Compatible with Jest API (easy migration if needed)
- Pro: Built-in coverage reporting (v8/istanbul)
- Con: Smaller ecosystem than Jest (fewer blog posts)

### Option C — Cypress Component Testing
- Pro: Visual component testing with real browser
- Con: Heavy setup, slow execution, overkill for unit/integration tests

## Decision

Use **Option B**. Vitest as the test runner, @nuxt/test-utils for composable and Nuxt-aware component testing, @vue/test-utils for lightweight component tests. Playwright added in Phase 3 for E2E only.

Test files are colocated with implementation (e.g., `useCart.test.ts` next to `useCart.ts`) following Go-inspired same-directory convention.

## Consequences

- **Positive:** Fast test execution with native Vite transforms
- **Positive:** Auto-imports (ref, computed, useFetch) work in tests via @nuxt/test-utils
- **Positive:** `npm test` and `npm run test:ci` scripts ready for CI/CD pipeline
- **Negative:** Team members familiar with Jest need minor adjustment (APIs are nearly identical)
- **Risks:** @nuxt/test-utils is newer and may have edge cases with complex SSR composables
