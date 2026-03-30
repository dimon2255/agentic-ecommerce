<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm animate-scale-in">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-display font-bold text-[var(--text-primary)]">Create account</h1>
        <p class="text-sm text-muted mt-1">Join FlexShop today</p>
      </div>

      <div class="card-dark p-6">
        <div v-if="errorMsg" class="mb-4 p-3 bg-red-900/20 border border-red-700/30 rounded-lg text-sm text-red-300">
          {{ errorMsg }}
        </div>
        <div v-if="successMsg" class="mb-4 p-3 bg-emerald-900/20 border border-emerald-700/30 rounded-lg text-sm text-emerald-300">
          {{ successMsg }}
        </div>

        <form @submit.prevent="handleRegister" class="space-y-4">
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
              minlength="6"
              autocomplete="new-password"
              class="input-dark"
            />
          </div>
          <div>
            <label for="confirm" class="block text-sm font-medium text-secondary mb-1.5">Confirm password</label>
            <input
              id="confirm"
              v-model="confirmPassword"
              type="password"
              required
              minlength="6"
              autocomplete="new-password"
              class="input-dark"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full btn-accent py-3 rounded-xl text-sm tracking-wide"
          >
            {{ loading ? 'Creating account...' : 'Create account' }}
          </button>
        </form>
      </div>

      <p class="mt-6 text-center text-sm text-muted">
        Already have an account?
        <NuxtLink to="/auth/login" class="text-accent hover:text-accent-hover font-medium transition-colors ml-1">
          Sign in
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
const confirmPassword = ref('')
const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

async function handleRegister() {
  loading.value = true
  errorMsg.value = ''
  successMsg.value = ''

  if (password.value !== confirmPassword.value) {
    errorMsg.value = 'Passwords do not match'
    loading.value = false
    return
  }

  const { error } = await client.auth.signUp({
    email: email.value,
    password: password.value,
  })

  if (error) {
    errorMsg.value = error.message
    loading.value = false
    return
  }

  // In local dev, Supabase auto-confirms. In prod, email confirmation may be required.
  successMsg.value = 'Account created! Redirecting...'
  setTimeout(() => router.push('/auth/login'), 1500)
}
</script>
