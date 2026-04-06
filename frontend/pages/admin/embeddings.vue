<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">Embeddings</h1>
      <p class="text-sm text-secondary mt-1">Manage AI search embeddings for products</p>
    </div>

    <!-- Status cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
      <AdminKpiCard label="Active Products" :value="status?.active_products ?? 0" :loading="statusLoading" />
      <AdminKpiCard label="Embedded Products" :value="status?.embedded_products ?? 0" :loading="statusLoading" />
      <AdminKpiCard
        label="Coverage"
        :value="coverageLabel"
        :loading="statusLoading"
        format="none"
      />
    </div>

    <!-- Actions -->
    <div class="card-dark p-6 space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-wider text-muted">Actions</h2>
      <div class="flex items-center gap-4">
        <button
          @click="showConfirmAll = true"
          :disabled="regenerating"
          class="btn-accent px-4 py-2 rounded-lg text-sm font-medium disabled:opacity-40"
        >
          {{ regenerating ? 'Regenerating...' : 'Regenerate All Embeddings' }}
        </button>
        <p v-if="message" :class="messageClass" class="text-sm">{{ message }}</p>
      </div>
      <p class="text-xs text-muted">
        Regenerates Voyage AI embeddings for all active products. Runs in the background.
      </p>
    </div>

    <!-- Product list with per-product regeneration -->
    <div class="card-dark overflow-hidden">
      <div class="px-6 py-4 border-b border-[var(--border-default)] flex items-center justify-between">
        <h2 class="text-sm font-semibold uppercase tracking-wider text-muted">Products</h2>
        <span class="text-xs text-muted">{{ products.length }} active</span>
      </div>
      <div v-if="productsLoading" class="p-6 space-y-3">
        <div v-for="i in 5" :key="i" class="h-10 bg-[var(--bg-hover)] rounded animate-pulse" />
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b border-[var(--border-default)]">
            <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Product</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Price</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Action</th>
          </tr>
        </thead>
        <tbody v-if="products.length === 0">
          <tr><td colspan="3" class="px-6 py-8 text-center text-muted">No active products</td></tr>
        </tbody>
        <tbody v-else>
          <tr v-for="p in products" :key="p.id" class="border-b border-[var(--border-subtle)]">
            <td class="px-6 py-3">{{ p.name }}</td>
            <td class="px-6 py-3 text-right">${{ Number(p.base_price).toFixed(2) }}</td>
            <td class="px-6 py-3 text-right">
              <button
                @click="confirmProduct = { id: p.id, name: p.name }"
                :disabled="regeneratingIds.has(p.id)"
                class="text-xs text-accent hover:text-accent-hover transition-colors disabled:opacity-40"
              >
                {{ regeneratingIds.has(p.id) ? 'Embedding...' : 'Regenerate' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- Confirm dialogs -->
    <AdminConfirmDialog
      :open="showConfirmAll"
      title="Regenerate All Embeddings"
      :message="`This will regenerate Voyage AI embeddings for all ${status?.active_products ?? 0} active products. Continue?`"
      confirm-text="Regenerate All"
      @confirm="doRegenerateAll"
      @cancel="showConfirmAll = false"
    />
    <AdminConfirmDialog
      :open="!!confirmProduct"
      title="Regenerate Embedding"
      :message="`Regenerate embedding for '${confirmProduct?.name}'?`"
      confirm-text="Regenerate"
      @confirm="doRegenerateProduct"
      @cancel="confirmProduct = null"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get, post } = useAdminApi()

interface EmbeddingStatus {
  active_products: number
  embedded_products: number
  coverage_complete: boolean
}

interface ProductRow {
  id: string
  name: string
  base_price: number
}

const status = ref<EmbeddingStatus | null>(null)
const statusLoading = ref(true)
const products = ref<ProductRow[]>([])
const productsLoading = ref(true)
const regenerating = ref(false)
const regeneratingIds = ref(new Set<string>())
const message = ref('')
const messageType = ref<'success' | 'error'>('success')
const showConfirmAll = ref(false)
const confirmProduct = ref<{ id: string; name: string } | null>(null)

const coverageLabel = computed(() => {
  if (!status.value) return '-'
  const { active_products, embedded_products } = status.value
  if (active_products === 0) return '100%'
  return `${Math.round((embedded_products / active_products) * 100)}%`
})

const messageClass = computed(() =>
  messageType.value === 'success' ? 'text-green-400' : 'text-[var(--color-error-border)]'
)

async function fetchStatus() {
  statusLoading.value = true
  try {
    status.value = await get<EmbeddingStatus>('/embeddings/status')
  } catch {
    status.value = null
  } finally {
    statusLoading.value = false
  }
}

async function fetchProducts() {
  productsLoading.value = true
  try {
    const res = await get<{ items: ProductRow[] }>('/catalog/products?limit=100&sort_by=name&sort_dir=asc')
    products.value = res.items ?? []
  } catch {
    products.value = []
  } finally {
    productsLoading.value = false
  }
}

async function doRegenerateAll() {
  showConfirmAll.value = false
  await regenerateAll()
}

async function doRegenerateProduct() {
  if (!confirmProduct.value) return
  const { id, name } = confirmProduct.value
  confirmProduct.value = null
  await regenerateProduct(id, name)
}

async function regenerateAll() {
  regenerating.value = true
  message.value = ''
  try {
    await post('/embeddings/regenerate', {})
    message.value = 'Regeneration started in background'
    messageType.value = 'success'
    // Refresh status after a short delay to show progress
    setTimeout(fetchStatus, 3000)
  } catch {
    message.value = 'Failed to start regeneration'
    messageType.value = 'error'
  } finally {
    regenerating.value = false
  }
}

async function regenerateProduct(productId: string, name: string) {
  regeneratingIds.value.add(productId)
  message.value = ''
  try {
    await post(`/embeddings/regenerate/${productId}`, {})
    message.value = `Embedded: ${name}`
    messageType.value = 'success'
    fetchStatus()
  } catch {
    message.value = `Failed to embed: ${name}`
    messageType.value = 'error'
  } finally {
    regeneratingIds.value.delete(productId)
  }
}

onMounted(() => {
  fetchStatus()
  fetchProducts()
})
</script>
