<template>
  <div class="min-h-screen bg-surface-base text-[var(--text-primary)] flex flex-col">
    <header class="fixed top-0 inset-x-0 z-50 glass-strong">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between items-center h-16">
          <NuxtLink to="/" class="flex items-center gap-0.5 group">
            <span class="text-xl font-display font-bold text-accent group-hover:text-accent-hover transition-colors">Flex</span>
            <span class="text-xl font-display font-bold text-[var(--text-primary)]">Shop</span>
          </NuxtLink>
          <nav class="flex items-center gap-7">
            <NuxtLink
              to="/catalog"
              class="text-sm font-medium text-secondary hover:text-[var(--text-primary)] transition-colors relative after:absolute after:bottom-[-4px] after:left-0 after:w-0 after:h-[2px] after:bg-accent after:transition-all after:duration-300 hover:after:w-full"
            >
              Catalog
            </NuxtLink>
            <NuxtLink
              to="/cart"
              class="text-sm font-medium text-secondary hover:text-[var(--text-primary)] transition-colors relative"
            >
              Cart
              <ClientOnly>
                <span
                  v-if="itemCount > 0"
                  class="absolute -top-2.5 -right-5 badge-amber text-[10px] w-5 h-5 flex items-center justify-center rounded-full"
                >
                  {{ itemCount > 9 ? '9+' : itemCount }}
                </span>
              </ClientOnly>
            </NuxtLink>
            <ClientOnly>
              <template v-if="user">
                <span class="text-sm text-muted hidden sm:inline">{{ user.email }}</span>
                <button
                  @click="handleLogout"
                  class="text-sm font-medium text-secondary hover:text-accent transition-colors"
                >
                  Sign out
                </button>
              </template>
              <template v-else>
                <NuxtLink
                  to="/auth/login"
                  class="text-sm font-medium text-secondary hover:text-[var(--text-primary)] transition-colors"
                >
                  Sign in
                </NuxtLink>
              </template>
            </ClientOnly>
          </nav>
        </div>
      </div>
    </header>
    <main class="pt-16 flex-1">
      <slot />
    </main>
    <footer class="border-t border-[var(--border-default)]">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 flex items-center justify-between">
        <p class="text-sm text-muted">
          &copy; {{ new Date().getFullYear() }} FlexShop
        </p>
        <div class="h-px flex-1 mx-6 bg-gradient-to-r from-transparent via-accent/20 to-transparent"></div>
        <p class="text-sm text-muted">All rights reserved</p>
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
