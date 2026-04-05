<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-display font-bold">Products</h1>
        <p class="text-sm text-secondary mt-1">Manage your product catalog</p>
      </div>
      <NuxtLink
        v-if="hasPermission('catalog:write')"
        to="/admin/products/new"
        class="btn-accent px-4 py-2 rounded-lg text-sm"
      >
        New Product
      </NuxtLink>
    </div>

    <AdminDataTable
      :columns="columns"
      :rows="products"
      :loading="loading"
      :total="total"
      :page="page"
      :per-page="perPage"
      :total-pages="totalPages"
      @update:page="page = $event; fetchProducts()"
      @sort="onSort"
      @row-click="(row: any) => navigateTo(`/admin/products/${row.slug}`)"
    >
      <template #toolbar>
        <div class="flex items-center gap-3">
          <input
            v-model="search"
            class="input-dark max-w-xs text-sm"
            placeholder="Search products..."
            @input="debouncedFetch"
          />
          <select v-model="statusFilter" class="input-dark max-w-[140px] text-sm" @change="page = 1; fetchProducts()">
            <option value="">All statuses</option>
            <option value="draft">Draft</option>
            <option value="active">Active</option>
            <option value="archived">Archived</option>
          </select>
        </div>
      </template>

      <template #cell-status="{ value }">
        <AdminStatusBadge :status="value" />
      </template>

      <template #cell-base_price="{ value }">
        ${{ Number(value).toFixed(2) }}
      </template>

      <template #cell-created_at="{ value }">
        {{ new Date(value).toLocaleDateString() }}
      </template>
    </AdminDataTable>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get } = useAdminApi()
const { hasPermission } = useAdminAuth()

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'status', label: 'Status' },
  { key: 'base_price', label: 'Price', sortable: true },
  { key: 'created_at', label: 'Created', sortable: true },
]

const products = ref<any[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const perPage = 20
const totalPages = ref(0)
const search = ref('')
const statusFilter = ref('')
const sortBy = ref('created_at')
const sortDir = ref('desc')

let debounceTimer: ReturnType<typeof setTimeout>
function debouncedFetch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { page.value = 1; fetchProducts() }, 300)
}

function onSort(key: string, dir: string) {
  sortBy.value = key
  sortDir.value = dir
  fetchProducts()
}

async function fetchProducts() {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      per_page: String(perPage),
      sort_by: sortBy.value,
      sort_dir: sortDir.value,
    })
    if (search.value) params.set('search', search.value)
    if (statusFilter.value) params.set('status', statusFilter.value)

    const resp = await get<any>(`/catalog/products?${params}`)
    products.value = resp.items
    total.value = resp.total
    totalPages.value = resp.total_pages
  } catch {
    products.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchProducts)
</script>
