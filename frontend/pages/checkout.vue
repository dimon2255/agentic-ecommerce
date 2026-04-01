<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-2xl font-display font-bold text-[var(--text-primary)] mb-8 animate-fade-in-up">Checkout</h1>

    <div v-if="cartLoading" class="text-muted animate-fade-in">Loading...</div>

    <div v-else-if="!cart?.items?.length" class="text-center py-20 animate-fade-in">
      <p class="text-muted mb-4">Your cart is empty</p>
      <NuxtLink to="/catalog" class="text-accent hover:text-accent-hover font-medium transition-colors">
        Browse catalog
      </NuxtLink>
    </div>

    <div v-else class="space-y-6 animate-fade-in-up delay-1">
      <!-- Step Indicator -->
      <div role="list" aria-label="Checkout steps" class="flex items-center gap-3 mb-2">
        <div role="listitem" :aria-current="step === 'shipping' ? 'step' : undefined" class="flex items-center gap-2">
          <div :class="['w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold', step === 'shipping' ? 'btn-accent' : 'bg-accent/20 text-accent']">1</div>
          <span :class="['text-sm font-medium', step === 'shipping' ? 'text-[var(--text-primary)]' : 'text-secondary']">Shipping</span>
        </div>
        <div class="flex-1 h-px bg-[var(--border-default)]"></div>
        <div role="listitem" :aria-current="step === 'payment' ? 'step' : undefined" class="flex items-center gap-2">
          <div :class="['w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold', step === 'payment' ? 'btn-accent' : 'bg-surface-elevated border border-[var(--border-default)] text-muted']">2</div>
          <span :class="['text-sm font-medium', step === 'payment' ? 'text-[var(--text-primary)]' : 'text-muted']">Payment</span>
        </div>
      </div>

      <!-- Order Summary -->
      <div class="card-dark p-6">
        <h2 class="text-sm font-medium text-muted uppercase tracking-wider mb-4">Order Summary</h2>
        <div v-for="item in cart.items" :key="item.id" class="flex justify-between py-2.5 border-b border-[var(--border-subtle)] last:border-0">
          <div>
            <span class="font-medium text-[var(--text-primary)] text-sm">{{ item.skus.products.name }}</span>
            <span class="text-xs text-muted ml-2 font-mono">{{ item.skus.sku_code }}</span>
            <span class="text-xs text-muted ml-2">&times;{{ item.quantity }}</span>
          </div>
          <span class="text-sm text-secondary">${{ (item.unit_price * item.quantity).toFixed(2) }}</span>
        </div>
        <div class="flex justify-between mt-4 pt-3 border-t border-[var(--border-default)] font-display font-bold">
          <span class="text-[var(--text-primary)]">Total</span>
          <span class="text-accent">${{ cartTotal.toFixed(2) }}</span>
        </div>
      </div>

      <!-- Price Change Warning -->
      <div v-if="priceChanges.length" class="bg-[var(--color-warning-bg)] border border-[var(--color-warning-border)] rounded-xl p-4">
        <p class="font-medium text-amber-300 text-sm">Some prices have been updated:</p>
        <ul class="mt-2 text-sm text-amber-400/80">
          <li v-for="change in priceChanges" :key="change.sku_id">
            {{ change.sku_code }}: ${{ change.old_price.toFixed(2) }} &rarr; ${{ change.new_price.toFixed(2) }}
          </li>
        </ul>
        <p class="mt-2 text-sm text-amber-400/80">Your cart has been updated. Please review and try again.</p>
      </div>

      <!-- Step 1: Shipping Form -->
      <div v-if="step === 'shipping'" class="card-dark p-6">
        <h2 class="text-sm font-medium text-muted uppercase tracking-wider mb-5">Shipping Information</h2>
        <form @submit.prevent="handleStartCheckout" class="space-y-4">
          <div>
            <label for="email" class="block text-sm font-medium text-secondary mb-1.5">Email</label>
            <input id="email" v-model="form.email" type="email" required class="input-dark" />
          </div>
          <div>
            <label for="shipping-name" class="block text-sm font-medium text-secondary mb-1.5">Full Name</label>
            <input id="shipping-name" v-model="form.name" type="text" required class="input-dark" />
          </div>
          <div>
            <label for="shipping-line1" class="block text-sm font-medium text-secondary mb-1.5">Address</label>
            <input id="shipping-line1" v-model="form.line1" type="text" required class="input-dark" />
          </div>
          <div>
            <label for="shipping-line2" class="block text-sm font-medium text-secondary mb-1.5">Apartment, suite, etc. (optional)</label>
            <input id="shipping-line2" v-model="form.line2" type="text" class="input-dark" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="shipping-city" class="block text-sm font-medium text-secondary mb-1.5">City</label>
              <input id="shipping-city" v-model="form.city" type="text" required class="input-dark" />
            </div>
            <div>
              <label for="shipping-state" class="block text-sm font-medium text-secondary mb-1.5">State / Province</label>
              <input id="shipping-state" v-model="form.state" type="text" class="input-dark" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="shipping-zip" class="block text-sm font-medium text-secondary mb-1.5">ZIP / Postal Code</label>
              <input id="shipping-zip" v-model="form.zip" type="text" required class="input-dark" />
            </div>
            <div>
              <label for="shipping-country" class="block text-sm font-medium text-secondary mb-1.5">Country</label>
              <input id="shipping-country" v-model="form.country" type="text" required class="input-dark" />
            </div>
          </div>

          <button type="submit" :disabled="checkoutLoading" class="w-full btn-accent py-3.5 rounded-xl text-sm tracking-wide mt-2">
            {{ checkoutLoading ? 'Processing...' : 'Continue to Payment' }}
          </button>

          <p v-if="checkoutError" role="alert" class="text-[var(--color-error)] text-sm text-center">{{ checkoutError }}</p>
        </form>
      </div>

      <!-- Step 2: Payment -->
      <div v-if="step === 'payment'" class="card-dark p-6">
        <h2 class="text-sm font-medium text-muted uppercase tracking-wider mb-5">Payment</h2>
        <div id="payment-element" class="mb-6"></div>
        <button @click="handlePayment" :disabled="paying" class="w-full btn-accent py-3.5 rounded-xl text-sm tracking-wide">
          {{ paying ? 'Processing payment...' : `Pay $${cartTotal.toFixed(2)}` }}
        </button>
        <p v-if="paymentError" role="alert" class="text-[var(--color-error)] text-sm text-center mt-3">{{ paymentError }}</p>
        <button @click="step = 'shipping'" class="w-full text-sm text-muted hover:text-secondary mt-4 transition-colors">
          Back to shipping
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { cart, loading: cartLoading, total: cartTotal, refresh: refreshCart } = useCart()
const { loading: checkoutLoading, error: checkoutError, priceChanges, startCheckout, initStripe, confirmPayment } = useCheckout()

const user = useSupabaseUser()
const step = ref<'shipping' | 'payment'>('shipping')
const orderId = ref('')
const paying = ref(false)
const paymentError = ref('')

const form = reactive({
  email: '',
  name: '',
  line1: '',
  line2: '',
  city: '',
  state: '',
  zip: '',
  country: 'US',
})

onMounted(async () => {
  await refreshCart()
  if (user.value?.email) {
    form.email = user.value.email
  }
})

async function handleStartCheckout() {
  const result = await startCheckout(form.email, {
    name: form.name,
    line1: form.line1,
    line2: form.line2,
    city: form.city,
    state: form.state,
    zip: form.zip,
    country: form.country,
  })

  if (!result) {
    if (priceChanges.value.length) {
      await refreshCart()
    }
    return
  }

  orderId.value = result.order_id
  step.value = 'payment'

  await nextTick()
  await initStripe(result.client_secret)
}

async function handlePayment() {
  paying.value = true
  paymentError.value = ''

  const errMsg = await confirmPayment(orderId.value)
  if (errMsg) {
    paymentError.value = errMsg
  }
  paying.value = false
}
</script>
