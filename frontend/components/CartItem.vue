<template>
  <div class="flex items-center gap-4 py-4 border-b border-gray-100">
    <div class="w-16 h-16 bg-gray-100 rounded-lg flex-shrink-0 overflow-hidden">
      <img
        v-if="item.skus.products.images?.length"
        :src="item.skus.products.images[0]"
        :alt="item.skus.products.name"
        class="w-full h-full object-cover"
      />
    </div>
    <div class="flex-1 min-w-0">
      <NuxtLink :to="`/product/${item.skus.products.slug}`" class="text-sm font-medium text-gray-900 hover:text-primary-600 truncate block">
        {{ item.skus.products.name }}
      </NuxtLink>
      <p class="text-xs text-gray-500 mt-0.5">{{ item.skus.sku_code }}</p>
      <p class="text-sm font-medium text-gray-900 mt-1">${{ item.unit_price.toFixed(2) }}</p>
    </div>
    <div class="flex items-center gap-2">
      <button
        @click="$emit('update', item.id, item.quantity - 1)"
        :disabled="item.quantity <= 1 || updating"
        class="w-8 h-8 flex items-center justify-center rounded-md border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        -
      </button>
      <span class="w-8 text-center text-sm font-medium">{{ item.quantity }}</span>
      <button
        @click="$emit('update', item.id, item.quantity + 1)"
        :disabled="updating"
        class="w-8 h-8 flex items-center justify-center rounded-md border border-gray-300 text-gray-600 hover:bg-gray-50 disabled:opacity-40"
      >
        +
      </button>
    </div>
    <div class="text-right w-20">
      <p class="text-sm font-medium text-gray-900">${{ (item.unit_price * item.quantity).toFixed(2) }}</p>
    </div>
    <button
      @click="$emit('remove', item.id)"
      :disabled="updating"
      class="text-gray-400 hover:text-red-500 transition-colors disabled:opacity-40"
      title="Remove item"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import type { CartItem as CartItemType } from '~/composables/useCart'

defineProps<{
  item: CartItemType
  updating: boolean
}>()

defineEmits<{
  update: [itemId: string, quantity: number]
  remove: [itemId: string]
}>()
</script>
