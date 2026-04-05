<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">Token Usage</h1>
      <p class="text-sm text-secondary mt-1">AI assistant token consumption</p>
    </div>

    <!-- Date range filter -->
    <div class="flex items-center gap-3">
      <input v-model="dateFrom" type="date" class="input-dark max-w-[160px] text-sm" @change="fetchData" />
      <span class="text-muted text-xs">to</span>
      <input v-model="dateTo" type="date" class="input-dark max-w-[160px] text-sm" @change="fetchData" />
    </div>

    <!-- Summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
      <AdminKpiCard label="Total Input Tokens" :value="totalInput" :loading="loading" />
      <AdminKpiCard label="Total Output Tokens" :value="totalOutput" :loading="loading" />
      <AdminKpiCard label="Total Requests" :value="totalRequests" :loading="loading" />
    </div>

    <!-- Usage chart -->
    <div class="card-dark p-6 space-y-4">
      <h2 class="text-sm font-semibold uppercase tracking-wider text-muted">Requests by Day</h2>
      <div v-if="loading" class="space-y-3">
        <div v-for="i in 5" :key="i" class="h-6 bg-[var(--bg-hover)] rounded animate-pulse" />
      </div>
      <AdminSimpleChart v-else :bars="requestBars" color="info" />
    </div>

    <!-- Data table -->
    <div class="card-dark overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[var(--border-default)]">
            <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Date</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Input Tokens</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Output Tokens</th>
            <th class="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-muted">Requests</th>
          </tr>
        </thead>
        <tbody v-if="usage.length === 0">
          <tr><td colspan="4" class="px-6 py-8 text-center text-muted">No usage data for this period</td></tr>
        </tbody>
        <tbody v-else>
          <tr v-for="row in usage" :key="row.day" class="border-b border-[var(--border-subtle)]">
            <td class="px-6 py-3">{{ row.day }}</td>
            <td class="px-6 py-3 text-right">{{ Number(row.input_tokens).toLocaleString() }}</td>
            <td class="px-6 py-3 text-right">{{ Number(row.output_tokens).toLocaleString() }}</td>
            <td class="px-6 py-3 text-right">{{ row.request_count }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get } = useAdminApi()

interface UsageDay {
  day: string
  input_tokens: number
  output_tokens: number
  request_count: number
}

const usage = ref<UsageDay[]>([])
const loading = ref(true)

const now = new Date()
const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
const dateFrom = ref(thirtyDaysAgo.toISOString().slice(0, 10))
const dateTo = ref(now.toISOString().slice(0, 10))

const totalInput = computed(() => usage.value.reduce((sum, u) => sum + Number(u.input_tokens), 0))
const totalOutput = computed(() => usage.value.reduce((sum, u) => sum + Number(u.output_tokens), 0))
const totalRequests = computed(() => usage.value.reduce((sum, u) => sum + u.request_count, 0))

const requestBars = computed(() =>
  usage.value.slice(0, 14).reverse().map(u => ({
    label: u.day.slice(5),
    value: u.request_count,
    formatted: String(u.request_count),
  }))
)

async function fetchData() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (dateFrom.value) params.set('date_from', dateFrom.value)
    if (dateTo.value) params.set('date_to', dateTo.value)
    usage.value = await get<UsageDay[]>(`/reports/token-usage?${params}`)
  } catch {
    usage.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>
