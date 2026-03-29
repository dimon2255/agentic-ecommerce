<template>
  <span class="text-2xl font-bold text-gray-900">${{ displayPrice.toFixed(2) }}</span>
</template>

<script setup lang="ts">
const props = defineProps<{
  basePrice: number
  priceOverride?: number | null
}>()

const displayPrice = computed(() => props.priceOverride ?? props.basePrice)
</script>
