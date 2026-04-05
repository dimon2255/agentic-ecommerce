<template>
  <div class="max-w-3xl space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-display font-bold">Edit Product</h1>
        <p class="text-sm text-secondary mt-1">{{ product?.name }}</p>
      </div>
      <button
        v-if="hasPermission('catalog:write')"
        class="text-sm text-[var(--color-error)] hover:text-[var(--color-error)]/80 transition-colors"
        @click="showDelete = true"
      >
        Delete
      </button>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 border-b border-[var(--border-default)]">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="px-4 py-2.5 text-sm font-medium transition-colors border-b-2 -mb-px"
        :class="activeTab === tab.key
          ? 'border-accent text-accent'
          : 'border-transparent text-secondary hover:text-[var(--text-primary)]'"
        @click="activeTab = tab.key"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Details tab -->
    <div v-if="activeTab === 'details'" class="card-dark p-6">
      <div v-if="loadingProduct" class="space-y-4">
        <div v-for="i in 5" :key="i" class="h-10 bg-[var(--bg-hover)] rounded animate-pulse" />
      </div>
      <ProductForm
        v-else
        :form="form"
        :categories="categories"
        :saving="saving"
        submit-label="Save Changes"
        @submit="handleUpdate"
      />
    </div>

    <!-- SKUs tab -->
    <div v-if="activeTab === 'skus'" class="space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-display font-semibold">SKU Variants</h2>
        <button
          v-if="hasPermission('catalog:write')"
          class="text-sm text-accent hover:text-accent-hover transition-colors"
          @click="showSkuForm = !showSkuForm"
        >
          {{ showSkuForm ? 'Cancel' : '+ Add SKU' }}
        </button>
      </div>

      <!-- New SKU form -->
      <div v-if="showSkuForm" class="card-dark p-5 space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-secondary mb-1.5">SKU Code</label>
            <input v-model="newSku.sku_code" class="input-dark text-sm" placeholder="e.g. PROD-BLK-L" />
          </div>
          <div>
            <label class="block text-sm font-medium text-secondary mb-1.5">Price Override</label>
            <input v-model.number="newSku.price_override" type="number" step="0.01" min="0" class="input-dark text-sm" placeholder="Leave empty for base price" />
          </div>
        </div>
        <button class="btn-accent px-4 py-2 rounded-lg text-sm" @click="handleCreateSku">
          Create SKU
        </button>
      </div>

      <!-- SKU list -->
      <div class="card-dark overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-[var(--border-default)]">
              <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Code</th>
              <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Price</th>
              <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Status</th>
              <th class="px-5 py-3 w-16" />
            </tr>
          </thead>
          <tbody v-if="skus.length === 0">
            <tr><td colspan="4" class="px-5 py-8 text-center text-muted">No SKUs yet</td></tr>
          </tbody>
          <tbody v-else>
            <tr v-for="sku in skus" :key="sku.id" class="border-b border-[var(--border-subtle)]">
              <td class="px-5 py-3 font-mono text-xs">{{ sku.sku_code }}</td>
              <td class="px-5 py-3">{{ sku.price_override != null ? `$${sku.price_override.toFixed(2)}` : 'Base' }}</td>
              <td class="px-5 py-3"><StatusBadge :status="sku.status" /></td>
              <td class="px-5 py-3">
                <button
                  v-if="hasPermission('catalog:write')"
                  class="text-xs text-[var(--color-error)] hover:text-[var(--color-error)]/80"
                  @click="skuToDelete = sku; showDeleteSku = true"
                >
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Delete confirmations -->
    <ConfirmDialog
      :open="showDelete"
      title="Delete Product"
      message="This will permanently delete this product and all its SKUs. This cannot be undone."
      confirm-text="Delete"
      variant="danger"
      @confirm="handleDelete"
      @cancel="showDelete = false"
    />
    <ConfirmDialog
      :open="showDeleteSku"
      title="Delete SKU"
      :message="`Delete SKU ${skuToDelete?.sku_code}?`"
      confirm-text="Delete"
      variant="danger"
      @confirm="handleDeleteSku"
      @cancel="showDeleteSku = false"
    />
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const route = useRoute()
const router = useRouter()
const { get, patch, post, del } = useAdminApi()
const { hasPermission } = useAdminAuth()
const { showToast } = useToast()

const slug = route.params.slug as string
const tabs = [
  { key: 'details', label: 'Details' },
  { key: 'skus', label: 'SKUs' },
]
const activeTab = ref('details')

const product = ref<any>(null)
const loadingProduct = ref(true)
const categories = ref<any[]>([])
const saving = ref(false)
const showDelete = ref(false)

const form = reactive({
  name: '',
  slug: '',
  category_id: '',
  description: '',
  base_price: 0,
  status: 'draft',
  images: [] as string[],
})

// SKU state
const skus = ref<any[]>([])
const showSkuForm = ref(false)
const newSku = reactive({ sku_code: '', price_override: null as number | null, status: 'active' })
const showDeleteSku = ref(false)
const skuToDelete = ref<any>(null)

onMounted(async () => {
  await Promise.all([fetchProduct(), fetchCategories(), fetchSkus()])
})

async function fetchProduct() {
  loadingProduct.value = true
  try {
    const p = await get<any>(`/catalog/products/${slug}`)
    product.value = p
    Object.assign(form, {
      name: p.name,
      slug: p.slug,
      category_id: p.category_id,
      description: p.description || '',
      base_price: p.base_price,
      status: p.status,
      images: p.images || [],
    })
  } catch {
    showToast('Product not found', 'error')
    router.push('/admin/products')
  } finally {
    loadingProduct.value = false
  }
}

async function fetchCategories() {
  try {
    const resp = await get<any>('/catalog/categories?per_page=100')
    categories.value = resp.items
  } catch {}
}

async function fetchSkus() {
  if (!product.value) return
  try {
    skus.value = await get<any[]>(`/catalog/products/${product.value.id}/skus`)
  } catch {
    skus.value = []
  }
}

async function handleUpdate() {
  saving.value = true
  try {
    const payload: any = {
      name: form.name,
      slug: form.slug,
      description: form.description || null,
      base_price: form.base_price,
      status: form.status,
      images: form.images,
    }
    await patch(`/catalog/products/${slug}`, payload)
    showToast('Product updated', 'success')
    if (form.slug !== slug) {
      router.push(`/admin/products/${form.slug}`)
    }
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to update', 'error')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  try {
    await del(`/catalog/products/${slug}`)
    showToast('Product deleted', 'success')
    router.push('/admin/products')
  } catch {
    showToast('Failed to delete product', 'error')
  }
  showDelete.value = false
}

async function handleCreateSku() {
  try {
    const payload: any = { sku_code: newSku.sku_code, status: newSku.status, attribute_values: [] }
    if (newSku.price_override != null && newSku.price_override > 0) {
      payload.price_override = newSku.price_override
    }
    await post(`/catalog/products/${product.value.id}/skus`, payload)
    showToast('SKU created', 'success')
    showSkuForm.value = false
    newSku.sku_code = ''
    newSku.price_override = null
    await fetchSkus()
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to create SKU', 'error')
  }
}

async function handleDeleteSku() {
  if (!skuToDelete.value) return
  try {
    await del(`/catalog/skus/${skuToDelete.value.id}`)
    showToast('SKU deleted', 'success')
    await fetchSkus()
  } catch {
    showToast('Failed to delete SKU', 'error')
  }
  showDeleteSku.value = false
  skuToDelete.value = null
}

// Fetch SKUs when product loads (if tab already open or product loads after mount)
watch(product, (p) => { if (p) fetchSkus() })
</script>
