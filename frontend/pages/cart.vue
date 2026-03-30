<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Shopping Cart</h1>

    <div v-if="loading" class="text-gray-500">Loading cart...</div>

    <div v-else-if="!cart?.items?.length" class="text-center py-16">
      <p class="text-gray-500 mb-4">Your cart is empty</p>
      <NuxtLink to="/catalog" class="text-primary-600 hover:text-primary-700 font-medium">
        Browse catalog
      </NuxtLink>
    </div>

    <div v-else>
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 px-6">
        <CartItem
          v-for="item in cart.items"
          :key="item.id"
          :item="item"
          :updating="updating"
          @update="handleUpdate"
          @remove="handleRemove"
        />
      </div>

      <div class="mt-6 bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div class="flex justify-between items-center text-lg font-bold text-gray-900">
          <span>Total</span>
          <span>${{ total.toFixed(2) }}</span>
        </div>
        <NuxtLink
          to="/checkout"
          class="mt-4 w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors text-center block"
        >
          Proceed to Checkout
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { cart, loading, total, refresh, updateItem, removeItem } = useCart()
const updating = ref(false)

onMounted(() => {
  refresh()
})

async function handleUpdate(itemId: string, quantity: number) {
  updating.value = true
  try {
    await updateItem(itemId, quantity)
  } finally {
    updating.value = false
  }
}

async function handleRemove(itemId: string) {
  updating.value = true
  try {
    await removeItem(itemId)
  } finally {
    updating.value = false
  }
}
</script>
