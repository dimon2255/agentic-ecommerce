<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">Audit Log</h1>
      <p class="text-sm text-secondary mt-1">Admin action history</p>
    </div>

    <AdminDataTable
      :columns="columns"
      :rows="entries"
      :loading="loading"
      :total="total"
      :page="page"
      :per-page="perPage"
      :total-pages="totalPages"
      empty-text="No audit entries found"
      @update:page="page = $event; fetchLog()"
      @row-click="toggleExpand"
    >
      <template #toolbar>
        <div class="flex flex-wrap items-center gap-3">
          <select v-model="actionFilter" class="input-dark max-w-[180px] text-sm" @change="page = 1; fetchLog()">
            <option value="">All actions</option>
            <option value="product:create">Product Create</option>
            <option value="product:update">Product Update</option>
            <option value="product:delete">Product Delete</option>
            <option value="category:create">Category Create</option>
            <option value="category:update">Category Update</option>
            <option value="category:delete">Category Delete</option>
            <option value="order:update_status">Order Status Update</option>
            <option value="sku:create">SKU Create</option>
            <option value="sku:delete">SKU Delete</option>
          </select>
          <select v-model="resourceFilter" class="input-dark max-w-[150px] text-sm" @change="page = 1; fetchLog()">
            <option value="">All resources</option>
            <option value="product">Product</option>
            <option value="category">Category</option>
            <option value="order">Order</option>
            <option value="sku">SKU</option>
            <option value="attribute">Attribute</option>
          </select>
          <input
            v-model="dateFrom"
            type="date"
            class="input-dark max-w-[160px] text-sm"
            @change="page = 1; fetchLog()"
          />
        </div>
      </template>

      <template #cell-created_at="{ value }">
        {{ new Date(value).toLocaleString() }}
      </template>

      <template #cell-action="{ value }">
        <span class="font-mono text-xs px-2 py-0.5 rounded bg-[var(--bg-surface)] text-secondary">
          {{ value }}
        </span>
      </template>

      <template #cell-resource_id="{ value }">
        <span v-if="value" class="font-mono text-xs text-muted">{{ value.slice(0, 8) }}...</span>
        <span v-else class="text-muted">-</span>
      </template>

      <template #cell-changes="{ row }">
        <button
          v-if="row.changes"
          class="text-xs text-accent hover:text-accent-hover transition-colors"
          @click.stop="toggleExpand(row)"
        >
          {{ expandedId === row.id ? 'Hide' : 'View' }}
        </button>
        <span v-else class="text-muted text-xs">-</span>
      </template>
    </AdminDataTable>

    <!-- Expanded changes viewer -->
    <Teleport to="body">
      <Transition name="dialog">
        <div v-if="expandedEntry" class="fixed inset-0 z-50 flex items-center justify-center">
          <div class="absolute inset-0 bg-black/60" @click="expandedId = ''" />
          <div class="relative glass-strong rounded-xl p-6 w-full max-w-lg max-h-[70vh] overflow-y-auto animate-scale-in">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-display font-semibold">Changes</h3>
              <button class="text-secondary hover:text-[var(--text-primary)]" @click="expandedId = ''">Close</button>
            </div>
            <pre class="text-xs text-secondary bg-[var(--bg-deep)] rounded-lg p-4 overflow-x-auto">{{ JSON.stringify(expandedEntry.changes, null, 2) }}</pre>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get } = useAdminApi()

const columns = [
  { key: 'created_at', label: 'Time' },
  { key: 'action', label: 'Action' },
  { key: 'resource_type', label: 'Resource' },
  { key: 'resource_id', label: 'ID' },
  { key: 'changes', label: 'Details' },
]

const entries = ref<any[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const perPage = 20
const totalPages = ref(0)
const actionFilter = ref('')
const resourceFilter = ref('')
const dateFrom = ref('')

const expandedId = ref('')
const expandedEntry = computed(() => entries.value.find(e => e.id === expandedId.value))

function toggleExpand(row: any) {
  if (!row.changes) return
  expandedId.value = expandedId.value === row.id ? '' : row.id
}

async function fetchLog() {
  loading.value = true
  try {
    const params = new URLSearchParams({
      page: String(page.value),
      per_page: String(perPage),
    })
    if (actionFilter.value) params.set('action', actionFilter.value)
    if (resourceFilter.value) params.set('resource_type', resourceFilter.value)
    if (dateFrom.value) params.set('date_from', dateFrom.value)

    const resp = await get<any>(`/audit-log?${params}`)
    entries.value = resp.items
    total.value = resp.total
    totalPages.value = resp.total_pages
  } catch {
    entries.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchLog)
</script>

<style scoped>
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.2s ease;
}
.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}
</style>
