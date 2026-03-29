export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss', '@nuxtjs/supabase'],

  supabase: {
    redirect: false,
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
    },
  },

  routeRules: {
    '/': { ssr: true },
    '/catalog/**': { ssr: true },
    '/product/**': { ssr: true },
    '/cart': { ssr: false },
    '/checkout': { ssr: false },
    '/account/**': { ssr: false },
  },

  compatibilityDate: '2025-01-01',
})
