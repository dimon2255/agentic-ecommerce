<template>
  <div>
    <section class="bg-white">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 text-center">
        <h1 class="text-4xl font-bold text-gray-900 sm:text-5xl">
          Welcome to FlexShop
        </h1>
        <p class="mt-4 text-lg text-gray-600 max-w-2xl mx-auto">
          Quality products, flexible choices. Browse our catalog and find exactly what you need.
        </p>
        <NuxtLink
          to="/catalog"
          class="mt-8 inline-block bg-primary-600 text-white px-8 py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors"
        >
          Browse Catalog
        </NuxtLink>
      </div>
    </section>
    <section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <h2 class="text-2xl font-bold text-gray-900 mb-6">Shop by Category</h2>
      <div v-if="categories?.length" class="grid grid-cols-2 md:grid-cols-4 gap-6">
        <CategoryCard v-for="cat in categories" :key="cat.id" :category="cat" />
      </div>
      <p v-else class="text-gray-500">No categories available.</p>
    </section>
  </div>
</template>

<script setup lang="ts">
const { get } = useApi()

const { data: categories } = await useAsyncData('categories', () =>
  get<Array<{ id: string; name: string; slug: string }>>('/categories?parent_id=null')
)
</script>
