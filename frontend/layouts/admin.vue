<template>
  <div class="min-h-screen bg-surface-base text-[var(--text-primary)]">
    <AdminSidebar />

    <!-- Main content area (offset by sidebar width) -->
    <div class="ml-64 min-h-screen flex flex-col">
      <!-- Top bar -->
      <header class="h-16 flex items-center justify-end px-8 border-b border-[var(--border-default)] bg-[var(--bg-base)]">
        <ClientOnly>
          <div class="flex items-center gap-4">
            <span v-if="roles?.length" class="text-xs font-medium px-2 py-0.5 rounded-full bg-accent/10 text-accent">
              {{ roles[0] }}
            </span>
            <span class="text-sm text-secondary truncate max-w-[200px]">{{ user?.email }}</span>
          </div>
        </ClientOnly>
      </header>

      <!-- Page content -->
      <main class="flex-1 p-8">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const user = useSupabaseUser()
const { roles } = useAdminAuth()
</script>
