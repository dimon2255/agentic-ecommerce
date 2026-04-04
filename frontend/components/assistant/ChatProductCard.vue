<template>
  <NuxtLink
    :to="`/product/${product.slug}`"
    class="flex-shrink-0 w-[200px] rounded-lg bg-surface-deep border border-[var(--border-default)] hover:border-accent/50 transition-all group overflow-hidden"
  >
    <div class="aspect-[3/2] bg-surface-deep overflow-hidden">
      <img
        v-if="product.images?.length"
        :src="product.images[0]"
        :alt="product.name"
        loading="lazy"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
      />
      <div v-else class="w-full h-full flex items-center justify-center">
        <svg class="w-8 h-8 text-muted/30" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
          <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.41a2.25 2.25 0 013.182 0l2.909 2.91M3.75 21h16.5a1.5 1.5 0 001.5-1.5V5.25a1.5 1.5 0 00-1.5-1.5H3.75a1.5 1.5 0 00-1.5 1.5v14.25c0 .828.672 1.5 1.5 1.5z" />
        </svg>
      </div>
    </div>
    <div class="px-2.5 py-2">
      <p class="text-xs font-medium text-[var(--text-primary)] truncate group-hover:text-accent transition-colors">{{ product.name }}</p>
      <p class="text-xs text-accent font-semibold mt-0.5">${{ product.base_price.toFixed(2) }}</p>
    </div>
  </NuxtLink>
</template>

<script setup lang="ts">
defineProps<{
  product: {
    id: string
    name: string
    slug: string
    base_price: number
    images?: string[]
  }
}>()
</script>
