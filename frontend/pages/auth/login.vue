<template>
  <div class="min-h-[60vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-bold text-gray-900 text-center mb-8">Sign in</h1>

      <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
        {{ errorMsg }}
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            autocomplete="email"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            autocomplete="current-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-primary-600 text-white py-2.5 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          {{ loading ? 'Signing in...' : 'Sign in' }}
        </button>
      </form>

      <p class="mt-6 text-center text-sm text-gray-500">
        Don't have an account?
        <NuxtLink to="/auth/register" class="text-primary-600 hover:text-primary-700 font-medium">
          Create one
        </NuxtLink>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
const client = useSupabaseClient()
const router = useRouter()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin() {
  loading.value = true
  errorMsg.value = ''

  const { error } = await client.auth.signInWithPassword({
    email: email.value,
    password: password.value,
  })

  if (error) {
    errorMsg.value = error.message
    loading.value = false
    return
  }

  // Merge guest cart if session exists
  try {
    const sessionId = localStorage.getItem('session_id')
    if (sessionId) {
      const { post } = useApi()
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        await post('/cart/merge', { session_id: sessionId }, {
          'Authorization': `Bearer ${session.access_token}`,
          'X-Session-ID': sessionId,
        })
      }
    }
  } catch {
    // Cart merge is best-effort
  }

  router.push('/')
}
</script>
