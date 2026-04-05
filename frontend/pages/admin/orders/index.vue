<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">Orders</h1>
      <p class="text-sm text-secondary mt-1">View and manage customer orders</p>
    </div>

    <DataTable
      :columns="columns"
      :rows="orders"
      :loading="loading"
      :total="total"
      :page="page"
      :per-page="perPage"
      :total-pages="totalPages"
      @update:page="page = $event; fetchOrders()"
      @sort="onSort"
      @row-click="(row: any) => navigateTo(`/admin/orders/${row.id}`)"
    >
      <template #toolbar>
        <div class="flex flex-wrap items-center gap-3">
          <input
            v-model="search"
            class="input-dark max-w-xs text-sm"
            placeholder="Search by email..."
            @input="debouncedFetch"
          />
          <select v-model="statusFilter" class="input-dark max-w-[150px] text-sm" @change="page = 1; fetchOrders()">
            <option value="">All statuses</option>
            <option value="pending">Pending</option>
            <option value="paid">Paid</option>
            <option value="shipped">Shipped</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <input
            v-model="dateFrom"
            type="date"
            class="input-dark max-w-[160px] text-sm"
            @change="page = 1; fetchOrders()"
          />
          <span class="text-muted text-xs">to</span>
          <input
            v-model="dateTo"
            type="date"
            class="input-dark max-w-[160px] text-sm"
            @change="page = 1; fetchOrders()"
          />
        </div>
      </template>

      <template #cell-id="{ value }">
        <span class="font-mono text-xs">{{ value.slice(0, 8) }}...</span>
      </template>

      <template #cell-status="{ value }">
        <StatusBadge :status="value" />
      </template>

      <template #cell-total="{ value }">
        ${{ Number(value).toFixed(2) }}
      </template>

      <template #cell-created_at="{ value }">
        {{ new Date(value).toLocaleDateString() }}
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get } = useAdminApi()

const columns = [
  { key: 'id', label: 'Order ID' },
  { key: 'email', label: 'Email', sortable: true },
  { key: 'status', label: 'Status' },
  { key: 'total', label: 'Total', sortable: true },
  { key: 'created_at', label: 'Date', sortable: true },
]

const orders = ref<any[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const perPage = 20
const totalPages = ref(0)
const search = ref('')
const statusFilter = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const sortBy = ref('created_at')
const sortDir = ref('desc')

let debounceTimer: ReturnType<typeof setTimeout>
function debouncedFetch() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { page.value = 1; fetchOrders() }, 300)
}

function onSort(key: string, dir: string) {
  sortBy.value = key
  sortDir.value = dir
  fetchOrders()
}

async function fetchOrders() {
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
    if (dateFrom.value) params.set('date_from', dateFrom.value)
    if (dateTo.value) params.set('date_to', dateTo.value)

    const resp = await get<any>(`/orders?${params}`)
    orders.value = resp.items
    total.value = resp.total
    totalPages.value = resp.total_pages
  } catch {
    orders.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchOrders)
</script>
