<template>
  <div>
    <!-- Hero -->
    <section class="hero-gradient relative overflow-hidden">
      <div class="absolute inset-0 bg-gradient-to-b from-transparent to-surface-base pointer-events-none"></div>
      <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32 text-center">
        <p class="text-accent text-sm font-medium tracking-widest uppercase mb-4 animate-fade-in">
          Curated for the modern lifestyle
        </p>
        <h1 class="text-4xl sm:text-5xl lg:text-6xl font-display font-bold text-[var(--text-primary)] leading-tight animate-fade-in-up">
          Welcome to
          <span class="text-accent">Flex</span>Shop
        </h1>
        <p class="mt-5 text-lg text-secondary max-w-2xl mx-auto leading-relaxed animate-fade-in-up delay-1">
          Quality products, flexible choices. Browse our catalog and find exactly what you need.
        </p>
        <NuxtLink
          to="/catalog"
          class="mt-10 inline-block btn-accent px-10 py-3.5 rounded-xl text-sm tracking-wide animate-fade-in-up delay-2"
        >
          Browse Catalog
        </NuxtLink>
      </div>
    </section>

    <!-- Categories -->
    <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
      <div class="flex items-center gap-4 mb-8 animate-fade-in-up delay-3">
        <h2 class="text-2xl font-display font-bold text-[var(--text-primary)]">Shop by Category</h2>
        <div class="flex-1 h-px bg-gradient-to-r from-[var(--border-strong)] to-transparent"></div>
      </div>
      <div v-if="categories?.length" class="grid grid-cols-2 md:grid-cols-4 gap-5">
        <div
          v-for="(cat, i) in categories"
          :key="cat.id"
          class="animate-fade-in-up"
          :class="`delay-${Math.min(i + 3, 6)}`"
        >
          <CategoryCard :category="cat" />
        </div>
      </div>
      <p v-else class="text-muted">No categories available.</p>
    </section>
  </div>
</template>

<script setup lang="ts">
const { get } = useApi()

const { data: categories } = await useAsyncData('categories', async () => {
  const result = await get<{ items: Array<{ id: string; name: string; slug: string }> }>('/categories?parent_id=null')
  return result.items ?? result
})
</script>
