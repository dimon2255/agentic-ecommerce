<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-3xl font-bold text-gray-900 mb-8">All Categories</h1>
    <div v-if="categories?.length" class="grid grid-cols-2 md:grid-cols-4 gap-6">
      <CategoryCard v-for="cat in categories" :key="cat.id" :category="cat" />
    </div>
    <p v-else class="text-gray-500">No categories found.</p>
  </div>
</template>

<script setup lang="ts">
const { get } = useApi()

const { data: categories } = await useAsyncData('all-categories', () =>
  get<Array<{ id: string; name: string; slug: string }>>('/categories')
)
</script>
