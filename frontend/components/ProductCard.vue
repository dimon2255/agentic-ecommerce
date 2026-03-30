<template>
  <NuxtLink
    :to="`/product/${product.slug}`"
    class="group block card-dark glow-hover overflow-hidden"
  >
    <div class="aspect-square bg-surface-deep flex items-center justify-center overflow-hidden">
      <img
        v-if="product.images?.length"
        :src="product.images[0]"
        :alt="product.name"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
      />
      <div v-else class="flex items-center justify-center w-full h-full">
        <svg class="w-12 h-12 text-muted/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
          <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
        </svg>
      </div>
    </div>
    <div class="p-4 space-y-2">
      <h3 class="font-display font-semibold text-[var(--text-primary)] group-hover:text-accent transition-colors truncate">
        {{ product.name }}
      </h3>
      <p v-if="product.description" class="text-sm text-secondary line-clamp-2 leading-relaxed">
        {{ product.description }}
      </p>
      <p class="text-lg font-bold text-accent">${{ product.base_price.toFixed(2) }}</p>
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
