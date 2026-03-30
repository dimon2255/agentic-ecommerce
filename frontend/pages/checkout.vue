<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <h1 class="text-2xl font-bold text-gray-900 mb-8">Checkout</h1>

    <div v-if="cartLoading" class="text-gray-500">Loading...</div>

    <div v-else-if="!cart?.items?.length" class="text-center py-16">
      <p class="text-gray-500 mb-4">Your cart is empty</p>
      <NuxtLink to="/catalog" class="text-primary-600 hover:text-primary-700 font-medium">
        Browse catalog
      </NuxtLink>
    </div>

    <div v-else>
      <!-- Order Summary -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Order Summary</h2>
        <div v-for="item in cart.items" :key="item.id" class="flex justify-between py-2 border-b border-gray-100 last:border-0">
          <div>
            <span class="font-medium text-gray-900">{{ item.skus.products.name }}</span>
            <span class="text-sm text-gray-500 ml-2">{{ item.skus.sku_code }}</span>
            <span class="text-sm text-gray-500 ml-2">x{{ item.quantity }}</span>
          </div>
          <span class="text-gray-900">${{ (item.unit_price * item.quantity).toFixed(2) }}</span>
        </div>
        <div class="flex justify-between mt-4 pt-3 border-t border-gray-200 text-lg font-bold text-gray-900">
          <span>Total</span>
          <span>${{ cartTotal.toFixed(2) }}</span>
        </div>
      </div>

      <!-- Price Change Warning -->
      <div v-if="priceChanges.length" class="bg-yellow-50 border border-yellow-200 rounded-xl p-4 mb-6">
        <p class="font-medium text-yellow-800">Some prices have been updated:</p>
        <ul class="mt-2 text-sm text-yellow-700">
          <li v-for="change in priceChanges" :key="change.sku_id">
            {{ change.sku_code }}: ${{ change.old_price.toFixed(2) }} &rarr; ${{ change.new_price.toFixed(2) }}
          </li>
        </ul>
        <p class="mt-2 text-sm text-yellow-700">Your cart has been updated. Please review and try again.</p>
      </div>

      <!-- Step 1: Shipping Form -->
      <div v-if="step === 'shipping'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Shipping Information</h2>
        <form @submit.prevent="handleStartCheckout" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input v-model="form.email" type="email" required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Full Name</label>
            <input v-model="form.name" type="text" required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
            <input v-model="form.line1" type="text" required
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Apartment, suite, etc. (optional)</label>
            <input v-model="form.line2" type="text"
              class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">City</label>
              <input v-model="form.city" type="text" required
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">State / Province</label>
              <input v-model="form.state" type="text"
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">ZIP / Postal Code</label>
              <input v-model="form.zip" type="text" required
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Country</label>
              <input v-model="form.country" type="text" required
                class="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-primary-500 focus:border-primary-500" />
            </div>
          </div>

          <button type="submit" :disabled="checkoutLoading"
            class="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed">
            {{ checkoutLoading ? 'Processing...' : 'Continue to Payment' }}
          </button>

          <p v-if="checkoutError" class="text-red-600 text-sm text-center">{{ checkoutError }}</p>
        </form>
      </div>

      <!-- Step 2: Payment -->
      <div v-if="step === 'payment'" class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Payment</h2>
        <div id="payment-element" class="mb-6"></div>
        <button @click="handlePayment" :disabled="paying"
          class="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed">
          {{ paying ? 'Processing payment...' : `Pay $${cartTotal.toFixed(2)}` }}
        </button>
        <p v-if="paymentError" class="text-red-600 text-sm text-center mt-2">{{ paymentError }}</p>
        <button @click="step = 'shipping'" class="w-full text-sm text-gray-500 hover:text-gray-700 mt-3">
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
