import { defineVitestConfig } from '@nuxt/test-utils/config'

export default defineVitestConfig({
  test: {
    environment: 'nuxt',
    globals: true,

    include: [
      'composables/**/*.test.ts',
      'components/**/*.test.ts',
    ],

    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['composables/**/*.ts', 'components/**/*.vue'],
      exclude: ['**/*.test.ts', '**/*.spec.ts', '**/node_modules/**', '**/__mocks__/**'],
      thresholds: {
        'composables/**/*.ts': {
          statements: 90,
          branches: 70,
          functions: 95,
          lines: 90,
        },
      },
    },

    testTimeout: 15000,
    hookTimeout: 15000,

    clearMocks: true,
    restoreMocks: true,

    environmentOptions: {
      nuxt: {
        mock: {
          intersectionObserver: true,
          indexedDb: true,
        },
      },
    },
  },
})
