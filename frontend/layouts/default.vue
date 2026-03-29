<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <NuxtLink to="/" class="text-xl font-bold text-gray-900 tracking-tight">
            FlexShop
          </NuxtLink>
          <nav class="flex items-center gap-6">
            <NuxtLink to="/catalog" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors">
              Catalog
            </NuxtLink>
            <NuxtLink to="/cart" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors relative">
              Cart
              <ClientOnly>
                <span
                  v-if="itemCount > 0"
                  class="absolute -top-2 -right-4 bg-primary-600 text-white text-xs font-bold w-5 h-5 flex items-center justify-center rounded-full"
                >
                  {{ itemCount > 9 ? '9+' : itemCount }}
                </span>
              </ClientOnly>
            </NuxtLink>
            <ClientOnly>
              <template v-if="user">
                <span class="text-sm text-gray-500">{{ user.email }}</span>
                <button
                  @click="handleLogout"
                  class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors"
                >
                  Sign out
                </button>
              </template>
              <template v-else>
                <NuxtLink to="/auth/login" class="text-sm font-medium text-gray-600 hover:text-gray-900 transition-colors">
                  Sign in
                </NuxtLink>
              </template>
            </ClientOnly>
          </nav>
        </div>
      </div>
    </header>
    <main>
      <slot />
    </main>
    <footer class="bg-white border-t border-gray-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <p class="text-sm text-gray-400 text-center">
          &copy; {{ new Date().getFullYear() }} FlexShop. All rights reserved.
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
const user = useSupabaseUser()
const client = useSupabaseClient()
const router = useRouter()
const { itemCount, refresh } = useCart()

onMounted(() => {
  refresh()
})

async function handleLogout() {
  await client.auth.signOut()
  router.push('/')
}
</script>
