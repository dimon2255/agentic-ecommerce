<template>
  <aside class="fixed top-0 left-0 bottom-0 w-64 glass-strong z-40 flex flex-col">
    <!-- Logo -->
    <div class="h-16 flex items-center px-6 border-b border-[var(--border-default)]">
      <NuxtLink to="/admin" class="flex items-center gap-1 group">
        <span class="text-lg font-display font-bold text-accent group-hover:text-accent-hover transition-colors">Flex</span>
        <span class="text-lg font-display font-bold text-[var(--text-primary)]">Admin</span>
      </NuxtLink>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto py-4 px-3 space-y-1">
      <SidebarLink to="/admin" icon="grid" label="Dashboard" :exact="true" />

      <SidebarSection v-if="hasPermission('catalog:read')" label="Catalog">
        <SidebarLink to="/admin/products" icon="package" label="Products" />
        <SidebarLink to="/admin/categories" icon="layers" label="Categories" />
      </SidebarSection>

      <SidebarLink
        v-if="hasPermission('orders:read')"
        to="/admin/orders"
        icon="receipt"
        label="Orders"
      />

      <SidebarSection v-if="hasPermission('reports:read')" label="Reports">
        <SidebarLink to="/admin/reports/sales" icon="trending-up" label="Sales" />
        <SidebarLink to="/admin/reports/token-usage" icon="cpu" label="Token Usage" />
      </SidebarSection>

      <SidebarLink
        v-if="hasPermission('audit:read')"
        to="/admin/audit-log"
        icon="scroll"
        label="Audit Log"
      />
    </nav>

    <!-- Footer -->
    <div class="px-4 py-3 border-t border-[var(--border-default)]">
      <NuxtLink
        to="/"
        class="flex items-center gap-2 text-sm text-secondary hover:text-accent transition-colors"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" /></svg>
        Back to shop
      </NuxtLink>
    </div>
  </aside>
</template>

<script setup lang="ts">
const { hasPermission } = useAdminAuth()

// Minimal inline sub-components to keep sidebar self-contained
const SidebarLink = defineComponent({
  props: {
    to: { type: String, required: true },
    icon: { type: String, required: true },
    label: { type: String, required: true },
    exact: { type: Boolean, default: false },
  },
  setup(props) {
    const route = useRoute()
    const isActive = computed(() =>
      props.exact
        ? route.path === props.to
        : route.path.startsWith(props.to)
    )
    return () =>
      h(resolveComponent('NuxtLink'), {
        to: props.to,
        class: [
          'flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
          isActive.value
            ? 'bg-accent/10 text-accent'
            : 'text-secondary hover:text-[var(--text-primary)] hover:bg-[var(--bg-hover)]',
        ],
      }, () => [
        h('span', { class: 'w-5 h-5 flex items-center justify-center text-xs opacity-60' }, iconMap[props.icon] || props.icon.charAt(0).toUpperCase()),
        h('span', props.label),
      ])
  },
})

const SidebarSection = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots }) {
    return () =>
      h('div', { class: 'pt-4' }, [
        h('p', { class: 'px-3 pb-1 text-[10px] font-semibold uppercase tracking-wider text-muted' }, props.label),
        slots.default?.(),
      ])
  },
})

const iconMap: Record<string, string> = {
  'grid': '\u25A6',
  'package': '\u25A3',
  'layers': '\u2630',
  'receipt': '\u2709',
  'trending-up': '\u2197',
  'cpu': '\u2699',
  'scroll': '\u2637',
}
</script>
