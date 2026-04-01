<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm animate-scale-in">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-display font-bold text-[var(--text-primary)]">Welcome back</h1>
        <p class="text-sm text-muted mt-1">Sign in to your account</p>
      </div>

      <div class="card-dark p-6">
        <div v-if="errorMsg" role="alert" class="mb-4 p-3 bg-[var(--color-error-bg)] border border-[var(--color-error-border)] rounded-lg text-sm text-red-300">
          {{ errorMsg }}
        </div>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label for="email" class="block text-sm font-medium text-secondary mb-1.5">Email</label>
            <input
              id="email"
              v-model="email"
              type="email"
              required
              autocomplete="email"
              class="input-dark"
            />
          </div>
          <div>
            <label for="password" class="block text-sm font-medium text-secondary mb-1.5">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              required
              autocomplete="current-password"
              class="input-dark"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full btn-accent py-3 rounded-xl text-sm tracking-wide"
          >
            {{ loading ? 'Signing in...' : 'Sign in' }}
          </button>
        </form>
      </div>

      <p class="mt-6 text-center text-sm text-muted">
        Don't have an account?
        <NuxtLink to="/auth/register" class="text-accent hover:text-accent-hover font-medium transition-colors ml-1">
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
