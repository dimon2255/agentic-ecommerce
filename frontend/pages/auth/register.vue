<template>
  <div class="min-h-[60vh] flex items-center justify-center px-4">
    <div class="w-full max-w-sm">
      <h1 class="text-2xl font-bold text-gray-900 text-center mb-8">Create account</h1>

      <div v-if="errorMsg" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-700">
        {{ errorMsg }}
      </div>
      <div v-if="successMsg" class="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-700">
        {{ successMsg }}
      </div>

      <form @submit.prevent="handleRegister" class="space-y-4">
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
            minlength="6"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <div>
          <label for="confirm" class="block text-sm font-medium text-gray-700 mb-1">Confirm password</label>
          <input
            id="confirm"
            v-model="confirmPassword"
            type="password"
            required
            minlength="6"
            autocomplete="new-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
          />
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-primary-600 text-white py-2.5 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
        >
          {{ loading ? 'Creating account...' : 'Create account' }}
        </button>
      </form>

      <p class="mt-6 text-center text-sm text-gray-500">
        Already have an account?
        <NuxtLink to="/auth/login" class="text-primary-600 hover:text-primary-700 font-medium">
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
