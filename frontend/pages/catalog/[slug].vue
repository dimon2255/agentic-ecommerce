<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div v-if="category" class="animate-fade-in">
      <nav aria-label="Breadcrumb" class="text-sm text-muted mb-6">
        <ol class="flex items-center gap-2">
          <li><NuxtLink to="/catalog" class="hover:text-secondary transition-colors">Catalog</NuxtLink></li>
          <li class="text-muted/50">/</li>
          <li aria-current="page" class="text-secondary">{{ category.name }}</li>
        </ol>
      </nav>

      <h1 class="text-3xl font-display font-bold text-[var(--text-primary)] mb-2 animate-fade-in-up">
        {{ category.name }}
      </h1>

      <div v-if="subcategories?.length" class="mb-10 animate-fade-in-up delay-1">
        <h2 class="text-sm font-medium text-muted uppercase tracking-wider mb-3">Subcategories</h2>
        <div class="flex flex-wrap gap-2">
          <NuxtLink
            v-for="sub in subcategories"
            :key="sub.id"
            :to="`/catalog/${sub.slug}`"
            class="px-4 py-2 bg-surface-elevated border border-[var(--border-default)] rounded-lg text-sm font-medium text-secondary hover:border-accent/30 hover:text-accent transition-all duration-200"
          >
            {{ sub.name }}
          </NuxtLink>
        </div>
      </div>

      <div v-if="productsLoading" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
        <SkeletonCard v-for="n in 8" :key="n" />
      </div>
      <div v-else-if="products?.length" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-5">
        <div
          v-for="(product, i) in products"
          :key="product.id"
          class="animate-fade-in-up"
          :class="`delay-${Math.min(i + 1, 6)}`"
        >
          <ProductCard :product="product" />
        </div>
      </div>
      <p v-else class="text-muted">No products in this category yet.</p>
    </div>
    <div v-else>
      <p class="text-muted">Category not found.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { get } = useApi()
const slug = route.params.slug as string

const { data: category } = await useAsyncData(`category-${slug}`, () =>
  get<{ id: string; name: string; slug: string }>(`/categories/${slug}`)
)

const { data: subcategories } = await useAsyncData(`subcats-${slug}`, async () => {
  if (!category.value) return []
  const result = await get<{ items: Array<{ id: string; name: string; slug: string }> }>(`/categories?parent_id=${category.value.id}`)
  return result.items ?? result
})

const { data: products, pending: productsLoading } = await useAsyncData(`products-${slug}`, async () => {
  if (!category.value) return []

  // Single API call with all category IDs (fixes N+1)
  const ids = [category.value.id]
  if (subcategories.value?.length) {
    ids.push(...subcategories.value.map(s => s.id))
  }
  const result = await get<{ items: Array<any> }>(`/products?category_ids=${ids.join(',')}`)
  return result.items ?? result
})
</script>
