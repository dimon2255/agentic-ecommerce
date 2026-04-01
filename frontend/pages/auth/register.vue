<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm animate-scale-in">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-display font-bold text-[var(--text-primary)]">Create account</h1>
        <p class="text-sm text-muted mt-1">Join FlexShop today</p>
      </div>

      <div class="card-dark p-6">
        <div v-if="errorMsg" role="alert" class="mb-4 p-3 bg-[var(--color-error-bg)] border border-[var(--color-error-border)] rounded-lg text-sm text-red-300">
          {{ errorMsg }}
        </div>
        <div v-if="successMsg" role="alert" class="mb-4 p-3 bg-[var(--color-success-bg)] border border-[var(--color-success-border)] rounded-lg text-sm text-emerald-300">
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
            <div class="relative">
              <input
                id="password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                required
                minlength="6"
                autocomplete="new-password"
                class="input-dark pr-10"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-muted hover:text-secondary transition-colors text-xs"
                :aria-label="showPassword ? 'Hide password' : 'Show password'"
              >
                {{ showPassword ? 'Hide' : 'Show' }}
              </button>
            </div>
          </div>
          <div>
            <label for="confirm" class="block text-sm font-medium text-secondary mb-1.5">Confirm password</label>
            <input
              id="confirm"
              v-model="confirmPassword"
              :type="showPassword ? 'text' : 'password'"
              required
              minlength="6"
              autocomplete="new-password"
              class="input-dark"
            />
            <p v-if="confirmPassword && password !== confirmPassword" class="text-xs text-[var(--color-error)] mt-1">
              Passwords do not match
            </p>
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
const showPassword = ref(false)
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
