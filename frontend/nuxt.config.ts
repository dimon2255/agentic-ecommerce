export default defineNuxtConfig({
  // Use flat directory structure (Nuxt 3 convention) instead of Nuxt 4's default app/ subdirectory
  srcDir: '.',
  dir: {
    app: 'app',
  },

  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/supabase'],

  css: ['~/assets/css/main.css'],

  supabase: {
    redirect: false,
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:9090',
      stripeKey: process.env.NUXT_PUBLIC_STRIPE_KEY || '',
    },
  },

  routeRules: {
    '/': { ssr: true },
    '/catalog/**': { ssr: true },
    '/product/**': { ssr: true },
    '/cart': { ssr: false },
    '/checkout': { ssr: false },
    '/auth/**': { ssr: false },
    '/account/**': { ssr: false },
    '/order/**': { ssr: false },
  },

  compatibilityDate: '2025-01-01',
})
