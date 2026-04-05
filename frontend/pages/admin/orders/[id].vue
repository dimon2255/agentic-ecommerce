<template>
  <div class="max-w-3xl space-y-6">
    <NuxtLink to="/admin/orders" class="inline-flex items-center gap-1 text-sm text-secondary hover:text-accent transition-colors">
      <span>&larr;</span> Back to orders
    </NuxtLink>

    <div v-if="loading" class="space-y-4">
      <div v-for="i in 4" :key="i" class="h-12 bg-[var(--bg-hover)] rounded animate-pulse" />
    </div>

    <template v-else-if="order">
      <!-- Header -->
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-display font-bold">Order Details</h1>
          <p class="text-sm text-secondary mt-1 font-mono">{{ order.id }}</p>
        </div>
        <AdminStatusBadge :status="order.status" />
      </div>

      <!-- Order info -->
      <div class="card-dark p-6 grid grid-cols-2 gap-6">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Customer</p>
          <p>{{ order.email }}</p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Date</p>
          <p>{{ new Date(order.created_at).toLocaleString() }}</p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Subtotal</p>
          <p>${{ order.subtotal.toFixed(2) }}</p>
        </div>
        <div>
          <p class="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Total</p>
          <p class="text-lg font-bold text-accent">${{ order.total.toFixed(2) }}</p>
        </div>
        <div v-if="order.shipping_address" class="col-span-2">
          <p class="text-xs font-semibold uppercase tracking-wider text-muted mb-1">Shipping Address</p>
          <p class="text-sm text-secondary">
            {{ formatAddress(order.shipping_address) }}
          </p>
        </div>
      </div>

      <!-- Status update -->
      <div v-if="hasPermission('orders:write')" class="card-dark p-6">
        <p class="text-xs font-semibold uppercase tracking-wider text-muted mb-3">Update Status</p>
        <div class="flex items-center gap-3">
          <select v-model="newStatus" class="input-dark max-w-[200px] text-sm">
            <option v-for="s in statusOptions" :key="s" :value="s" :disabled="s === order.status">
              {{ s }}
            </option>
          </select>
          <button
            class="btn-accent px-4 py-2 rounded-lg text-sm"
            :disabled="newStatus === order.status || updating"
            @click="handleUpdateStatus"
          >
            {{ updating ? 'Updating...' : 'Update' }}
          </button>
        </div>
      </div>

      <!-- Order items -->
      <div class="card-dark overflow-hidden">
        <div class="px-6 py-4 border-b border-[var(--border-default)]">
          <h2 class="text-sm font-semibold uppercase tracking-wider text-muted">Items</h2>
        </div>
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-[var(--border-default)]">
              <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Product</th>
              <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">SKU</th>
              <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Qty</th>
              <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Price</th>
              <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Line Total</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in order.items" :key="item.id" class="border-b border-[var(--border-subtle)]">
              <td class="px-6 py-3">{{ item.product_name }}</td>
              <td class="px-6 py-3 font-mono text-xs text-secondary">{{ item.sku_code }}</td>
              <td class="px-6 py-3 text-right">{{ item.quantity }}</td>
              <td class="px-6 py-3 text-right">${{ item.unit_price.toFixed(2) }}</td>
              <td class="px-6 py-3 text-right font-medium">${{ (item.quantity * item.unit_price).toFixed(2) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const route = useRoute()
const { get, patch } = useAdminApi()
const { hasPermission } = useAdminAuth()
const { showToast } = useToast()

const orderId = route.params.id as string
const statusOptions = ['pending', 'paid', 'shipped', 'completed', 'cancelled']

const order = ref<any>(null)
const loading = ref(true)
const newStatus = ref('')
const updating = ref(false)

function formatAddress(addr: any): string {
  if (!addr) return ''
  const parts = [addr.name, addr.line1, addr.line2, addr.city, addr.state, addr.zip, addr.country]
  return parts.filter(Boolean).join(', ')
}

async function fetchOrder() {
  loading.value = true
  try {
    order.value = await get<any>(`/orders/${orderId}`)
    newStatus.value = order.value.status
  } catch {
    showToast('Order not found', 'error')
  } finally {
    loading.value = false
  }
}

async function handleUpdateStatus() {
  updating.value = true
  try {
    await patch(`/orders/${orderId}/status`, { status: newStatus.value })
    order.value.status = newStatus.value
    showToast('Status updated', 'success')
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to update status', 'error')
  } finally {
    updating.value = false
  }
}

onMounted(fetchOrder)
</script>
