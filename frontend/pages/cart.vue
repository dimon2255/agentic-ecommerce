<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-2xl font-display font-bold text-[var(--text-primary)] mb-8 animate-fade-in-up">Shopping Cart</h1>

    <div v-if="loading" class="text-muted animate-fade-in">Loading cart...</div>

    <div v-else-if="!cart?.items?.length" class="text-center py-20 animate-fade-in">
      <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-surface-elevated border border-[var(--border-default)] flex items-center justify-center">
        <svg class="w-7 h-7 text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" />
        </svg>
      </div>
      <p class="text-muted mb-4">Your cart is empty</p>
      <NuxtLink to="/catalog" class="text-accent hover:text-accent-hover font-medium transition-colors">
        Browse catalog
      </NuxtLink>
    </div>

    <div v-else class="animate-fade-in-up delay-1">
      <div class="card-dark px-6">
        <CartItem
          v-for="item in cart.items"
          :key="item.id"
          :item="item"
          :updating="updatingItems.has(item.id)"
          @update="handleUpdate"
          @remove="handleRemove"
        />
      </div>

      <div class="mt-6 card-dark p-6">
        <div class="flex justify-between items-center text-lg font-display font-bold">
          <span class="text-[var(--text-primary)]">Total</span>
          <span class="text-accent">${{ total.toFixed(2) }}</span>
        </div>
        <NuxtLink
          to="/checkout"
          class="mt-5 w-full btn-accent py-3.5 rounded-xl text-sm tracking-wide text-center block"
        >
          Proceed to Checkout
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { cart, loading, total, refresh, updateItem, removeItem } = useCart()
const updatingItems = ref(new Set<string>())

onMounted(() => {
  refresh(true)
})

async function handleUpdate(itemId: string, quantity: number) {
  updatingItems.value.add(itemId)
  // Optimistic update
  const item = cart.value?.items.find(i => i.id === itemId)
  const prevQty = item?.quantity
  if (item) item.quantity = quantity
  try {
    await updateItem(itemId, quantity)
  } catch {
    // Revert on failure
    if (item && prevQty !== undefined) item.quantity = prevQty
  } finally {
    updatingItems.value.delete(itemId)
  }
}

async function handleRemove(itemId: string) {
  updatingItems.value.add(itemId)
  try {
    await removeItem(itemId)
  } finally {
    updatingItems.value.delete(itemId)
  }
}
</script>
