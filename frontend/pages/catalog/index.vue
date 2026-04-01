<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div class="flex items-center gap-4 mb-8 animate-fade-in-up">
      <h1 class="text-3xl font-display font-bold text-[var(--text-primary)]">All Categories</h1>
      <div class="flex-1 h-px bg-gradient-to-r from-[var(--border-strong)] to-transparent"></div>
    </div>
    <div v-if="categories?.length" class="grid grid-cols-2 md:grid-cols-4 gap-5">
      <div
        v-for="(cat, i) in categories"
        :key="cat.id"
        class="animate-fade-in-up"
        :class="`delay-${Math.min(i + 1, 6)}`"
      >
        <CategoryCard :category="cat" />
      </div>
    </div>
    <p v-else class="text-muted">No categories found.</p>
  </div>
</template>

<script setup lang="ts">
const { get } = useApi()

const { data: categories } = await useAsyncData('all-categories', async () => {
  const result = await get<{ items: Array<{ id: string; name: string; slug: string }> }>('/categories')
  return result.items ?? result
})
</script>
