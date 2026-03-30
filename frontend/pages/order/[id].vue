<template>
  <div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-12">

    <!-- Payment Failed -->
    <div v-if="redirectStatus === 'failed'" class="text-center py-20 animate-scale-in">
      <div class="w-16 h-16 bg-red-900/30 border border-red-700/30 rounded-full flex items-center justify-center mx-auto mb-5">
        <span class="text-red-400 text-3xl font-bold">&times;</span>
      </div>
      <h1 class="text-2xl font-display font-bold text-[var(--text-primary)] mb-2">Payment Failed</h1>
      <p class="text-secondary mb-6">Your payment could not be processed. Please try again.</p>
      <NuxtLink to="/cart" class="text-accent hover:text-accent-hover font-medium transition-colors">
        Return to Cart
      </NuxtLink>
    </div>

    <!-- Processing / Waiting for webhook -->
    <div v-else-if="!order || order.status === 'draft' || order.status === 'pending'" class="text-center py-20 animate-fade-in">
      <div class="w-12 h-12 border-[3px] border-accent border-t-transparent rounded-full animate-spin mx-auto mb-5"></div>
      <p class="text-secondary">Confirming your payment...</p>
    </div>

    <!-- Order Confirmed -->
    <div v-else-if="order.status === 'paid'" class="animate-scale-in">
      <div class="text-center mb-8">
        <div class="w-16 h-16 bg-emerald-900/30 border border-emerald-700/30 rounded-full flex items-center justify-center mx-auto mb-5">
          <span class="text-emerald-400 text-3xl font-bold">&check;</span>
        </div>
        <h1 class="text-2xl font-display font-bold text-[var(--text-primary)] mb-1">Order Confirmed!</h1>
        <p class="text-secondary">Thank you for your purchase.</p>
      </div>

      <div class="card-dark p-6">
        <div class="flex justify-between text-sm mb-3">
          <span class="text-muted">Order ID</span>
          <span class="font-mono text-secondary">{{ order.id.slice(0, 8) }}...</span>
        </div>
        <div class="flex justify-between text-sm mb-4">
          <span class="text-muted">Confirmation sent to</span>
          <span class="text-secondary">{{ order.email }}</span>
        </div>

        <div class="border-t border-[var(--border-default)] pt-4">
          <div v-for="item in order.items" :key="item.sku_code" class="flex justify-between py-2.5">
            <div>
              <span class="text-[var(--text-primary)] text-sm">{{ item.product_name }}</span>
              <span class="text-xs text-muted ml-2 font-mono">{{ item.sku_code }}</span>
              <span class="text-xs text-muted ml-2">&times;{{ item.quantity }}</span>
            </div>
            <span class="text-sm text-secondary">${{ (item.unit_price * item.quantity).toFixed(2) }}</span>
          </div>
        </div>

        <div class="border-t border-[var(--border-default)] pt-4 mt-2 flex justify-between font-display font-bold">
          <span class="text-[var(--text-primary)]">Total</span>
          <span class="text-accent">${{ order.total.toFixed(2) }}</span>
        </div>
      </div>

      <div class="text-center mt-8">
        <NuxtLink to="/catalog" class="text-accent hover:text-accent-hover font-medium transition-colors">
          Continue Shopping
        </NuxtLink>
      </div>
    </div>

    <!-- Cancelled -->
    <div v-else-if="order.status === 'cancelled'" class="text-center py-20 animate-scale-in">
      <h1 class="text-2xl font-display font-bold text-[var(--text-primary)] mb-2">Order Cancelled</h1>
      <p class="text-secondary mb-6">This order has been cancelled.</p>
      <NuxtLink to="/cart" class="text-accent hover:text-accent-hover font-medium transition-colors">
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
