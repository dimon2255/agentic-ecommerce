<template>
  <div class="space-y-8">
    <div>
      <h1 class="text-2xl font-display font-bold">Dashboard</h1>
      <p class="text-sm text-secondary mt-1">Overview of your store</p>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <AdminKpiCard
        label="Total Revenue"
        :value="kpis?.total_revenue ?? 0"
        prefix="$"
        format="currency"
        :loading="loading"
      />
      <AdminKpiCard
        label="Total Orders"
        :value="kpis?.total_orders ?? 0"
        :loading="loading"
      />
      <AdminKpiCard
        label="Active Products"
        :value="kpis?.active_products ?? 0"
        :loading="loading"
      />
      <AdminKpiCard
        label="Customers"
        :value="kpis?.total_customers ?? 0"
        :loading="loading"
      />
    </div>

    <!-- Quick links -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6">
      <NuxtLink
        v-if="hasPermission('catalog:read')"
        to="/admin/products"
        class="card-dark glow-hover p-6 space-y-2"
      >
        <p class="font-display font-semibold text-accent">Products</p>
        <p class="text-sm text-secondary">Manage your product catalog</p>
      </NuxtLink>
      <NuxtLink
        v-if="hasPermission('orders:read')"
        to="/admin/orders"
        class="card-dark glow-hover p-6 space-y-2"
      >
        <p class="font-display font-semibold text-accent">Orders</p>
        <p class="text-sm text-secondary">View and manage orders</p>
      </NuxtLink>
      <NuxtLink
        v-if="hasPermission('reports:read')"
        to="/admin/reports/sales"
        class="card-dark glow-hover p-6 space-y-2"
      >
        <p class="font-display font-semibold text-accent">Reports</p>
        <p class="text-sm text-secondary">Sales and usage analytics</p>
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'admin',
  middleware: 'admin',
})

interface DashboardKPIs {
  total_orders: number
  total_revenue: number
  active_products: number
  total_customers: number
}

const { get } = useAdminApi()
const { hasPermission } = useAdminAuth()

const loading = ref(true)
const kpis = ref<DashboardKPIs | null>(null)

onMounted(async () => {
  try {
    kpis.value = await get<DashboardKPIs>('/reports/dashboard')
  } catch {
    // Dashboard data is non-critical; show zeros
  } finally {
    loading.value = false
  }
})
</script>
