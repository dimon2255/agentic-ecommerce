<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-12">

    <!-- Payment Failed -->
    <div v-if="redirectStatus === 'failed'" class="text-center py-16">
      <div class="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <span class="text-red-600 text-3xl font-bold">&times;</span>
      </div>
      <h1 class="text-2xl font-bold text-gray-900 mb-2">Payment Failed</h1>
      <p class="text-gray-600 mb-6">Your payment could not be processed. Please try again.</p>
      <NuxtLink to="/cart" class="text-primary-600 hover:text-primary-700 font-medium">
        Return to Cart
      </NuxtLink>
    </div>

    <!-- Processing / Waiting for webhook -->
    <div v-else-if="!order || order.status === 'draft'" class="text-center py-16">
      <div class="w-10 h-10 border-4 border-primary-600 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
      <p class="text-gray-600">Confirming your payment...</p>
    </div>

    <!-- Order Confirmed -->
    <div v-else-if="order.status === 'paid'">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
          <span class="text-green-600 text-3xl font-bold">&check;</span>
        </div>
        <h1 class="text-2xl font-bold text-gray-900 mb-1">Order Confirmed!</h1>
        <p class="text-gray-600">Thank you for your purchase.</p>
      </div>

      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div class="flex justify-between text-sm mb-3">
          <span class="text-gray-500">Order ID</span>
          <span class="font-mono text-gray-700">{{ order.id.slice(0, 8) }}...</span>
        </div>
        <div class="flex justify-between text-sm mb-4">
          <span class="text-gray-500">Confirmation sent to</span>
          <span class="text-gray-700">{{ order.email }}</span>
        </div>

        <div class="border-t border-gray-200 pt-4">
          <div v-for="item in order.items" :key="item.sku_code" class="flex justify-between py-2">
            <div>
              <span class="text-gray-900">{{ item.product_name }}</span>
              <span class="text-sm text-gray-500 ml-2">{{ item.sku_code }}</span>
              <span class="text-sm text-gray-500 ml-2">&times;{{ item.quantity }}</span>
            </div>
            <span class="text-gray-900">${{ (item.unit_price * item.quantity).toFixed(2) }}</span>
          </div>
        </div>

        <div class="border-t border-gray-200 pt-4 mt-2 flex justify-between font-bold text-gray-900">
          <span>Total</span>
          <span>${{ order.total.toFixed(2) }}</span>
        </div>
      </div>

      <div class="text-center mt-8">
        <NuxtLink to="/catalog" class="text-primary-600 hover:text-primary-700 font-medium">
          Continue Shopping
        </NuxtLink>
      </div>
    </div>

    <!-- Cancelled -->
    <div v-else-if="order.status === 'cancelled'" class="text-center py-16">
      <h1 class="text-2xl font-bold text-gray-900 mb-2">Order Cancelled</h1>
      <p class="text-gray-600 mb-6">This order has been cancelled.</p>
      <NuxtLink to="/cart" class="text-primary-600 hover:text-primary-700 font-medium">
        Return to Cart
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { OrderResponse } from '~/composables/useCheckout'

const route = useRoute()
const { getOrder } = useCheckout()

const orderId = route.params.id as string
const redirectStatus = (route.query.redirect_status as string) || ''

const order = ref<OrderResponse | null>(null)

onMounted(() => {
  if (redirectStatus === 'failed') return
  pollOrderStatus()
})

async function pollOrderStatus() {
  for (let i = 0; i < 15; i++) {
    try {
      order.value = await getOrder(orderId)
      if (order.value?.status === 'paid' || order.value?.status === 'cancelled') return
    } catch {}
    await new Promise(resolve => setTimeout(resolve, 2000))
  }
}
</script>
