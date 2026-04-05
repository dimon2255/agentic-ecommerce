<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">Sales Report</h1>
      <p class="text-sm text-secondary mt-1">Revenue and order trends</p>
    </div>

    <!-- Date range filter -->
    <div class="flex items-center gap-3">
      <input v-model="dateFrom" type="date" class="input-dark max-w-[160px] text-sm" @change="fetchData" />
      <span class="text-muted text-xs">to</span>
      <input v-model="dateTo" type="date" class="input-dark max-w-[160px] text-sm" @change="fetchData" />
    </div>

    <!-- Summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
      <KpiCard label="Total Revenue" :value="totalRevenue" prefix="$" format="currency" :loading="loading" />
      <KpiCard label="Total Orders" :value="totalOrders" :loading="loading" />
      <KpiCard label="Avg Order Value" :value="avgOrder" prefix="$" format="currency" :loading="loading" />
    </div>

    <!-- Revenue chart -->
    <div class="card-dark p-6 space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-wider text-muted">Revenue by Day</h2>
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 5" :key="i" class="h-6 bg-[var(--bg-hover)] rounded animate-pulse" />
      </div>
      <SimpleChart v-else :bars="revenueBars" color="accent" />
    </div>

    <!-- Data table -->
    <div class="card-dark overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[var(--border-default)]">
            <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Date</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Orders</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Revenue</th>
          </tr>
        </thead>
        <tbody v-if="sales.length === 0">
          <tr><td colspan="3" class="px-6 py-8 text-center text-muted">No sales data for this period</td></tr>
        </tbody>
        <tbody v-else>
          <tr v-for="row in sales" :key="row.day" class="border-b border-[var(--border-subtle)]">
            <td class="px-6 py-3">{{ row.day }}</td>
            <td class="px-6 py-3 text-right">{{ row.order_count }}</td>
            <td class="px-6 py-3 text-right font-medium">${{ Number(row.revenue).toFixed(2) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get } = useAdminApi()

interface SalesDay {
  day: string
  order_count: number
  revenue: number
}

const sales = ref<SalesDay[]>([])
const loading = ref(true)

// Default to last 30 days
const now = new Date()
const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
const dateFrom = ref(thirtyDaysAgo.toISOString().slice(0, 10))
const dateTo = ref(now.toISOString().slice(0, 10))

const totalRevenue = computed(() => sales.value.reduce((sum, s) => sum + Number(s.revenue), 0))
const totalOrders = computed(() => sales.value.reduce((sum, s) => sum + s.order_count, 0))
const avgOrder = computed(() => totalOrders.value > 0 ? totalRevenue.value / totalOrders.value : 0)

const revenueBars = computed(() =>
  sales.value.slice(0, 14).reverse().map(s => ({
    label: s.day.slice(5), // MM-DD
    value: Number(s.revenue),
    formatted: `$${Number(s.revenue).toFixed(0)}`,
  }))
)

async function fetchData() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (dateFrom.value) params.set('date_from', dateFrom.value)
    if (dateTo.value) params.set('date_to', dateTo.value)
    sales.value = await get<SalesDay[]>(`/reports/sales?${params}`)
  } catch {
    sales.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>
