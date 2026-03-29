<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div v-if="category">
      <nav class="text-sm text-gray-500 mb-6">
        <NuxtLink to="/catalog" class="hover:text-gray-700">Catalog</NuxtLink>
        <span class="mx-2">/</span>
        <span class="text-gray-900">{{ category.name }}</span>
      </nav>
      <h1 class="text-3xl font-bold text-gray-900 mb-2">{{ category.name }}</h1>
      <div v-if="subcategories?.length" class="mb-8">
        <h2 class="text-lg font-semibold text-gray-700 mb-3">Subcategories</h2>
        <div class="flex flex-wrap gap-3">
          <NuxtLink v-for="sub in subcategories" :key="sub.id" :to="`/catalog/${sub.slug}`"
            class="px-4 py-2 bg-white border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:border-primary-300 hover:text-primary-600 transition-colors">
            {{ sub.name }}
          </NuxtLink>
        </div>
      </div>
      <div v-if="products?.length" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
        <ProductCard v-for="product in products" :key="product.id" :product="product" />
      </div>
      <p v-else class="text-gray-500">No products in this category yet.</p>
    </div>
    <div v-else>
      <p class="text-gray-500">Category not found.</p>
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

const { data: subcategories } = await useAsyncData(`subcats-${slug}`, () =>
  category.value
    ? get<Array<{ id: string; name: string; slug: string }>>(`/categories?parent_id=${category.value.id}`)
    : Promise.resolve([])
)

const { data: products } = await useAsyncData(`products-${slug}`, () =>
  category.value
    ? get<Array<any>>(`/products?category_id=${category.value.id}`)
    : Promise.resolve([])
)
</script>
