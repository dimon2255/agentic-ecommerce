<template>
  <NuxtLink
    :to="`/product/${product.slug}`"
    class="group block bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden hover:shadow-md transition-shadow"
  >
    <div class="aspect-square bg-gray-100 flex items-center justify-center">
      <img v-if="product.images?.length" :src="product.images[0]" :alt="product.name" class="w-full h-full object-cover" />
      <span v-else class="text-gray-300 text-5xl">&#9744;</span>
    </div>
    <div class="p-4">
      <h3 class="font-semibold text-gray-900 group-hover:text-primary-600 transition-colors truncate">{{ product.name }}</h3>
      <p v-if="product.description" class="mt-1 text-sm text-gray-500 line-clamp-2">{{ product.description }}</p>
      <p class="mt-2 text-lg font-bold text-gray-900">${{ product.base_price.toFixed(2) }}</p>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{
  product: {
    id: string; name: string; slug: string; description?: string; base_price: number; images?: string[]
  }
}>()
</script>
