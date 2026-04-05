<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-display font-bold">Categories</h1>
        <p class="text-sm text-secondary mt-1">Organize your product catalog</p>
      </div>
      <NuxtLink
        v-if="hasPermission('catalog:write')"
        to="/admin/categories/new"
        class="btn-accent px-4 py-2 rounded-lg text-sm"
      >
        New Category
      </NuxtLink>
    </div>

    <div class="card-dark overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-[var(--border-default)]">
            <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Name</th>
            <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Slug</th>
            <th class="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Type</th>
          </tr>
        </thead>
        <tbody v-if="loading">
          <tr v-for="i in 6" :key="i">
            <td class="px-6 py-4" colspan="3">
              <div class="h-4 bg-[var(--bg-hover)] rounded animate-pulse" />
            </td>
          </tr>
        </tbody>
        <tbody v-else-if="tree.length === 0">
          <tr><td colspan="3" class="px-6 py-12 text-center text-muted">No categories yet</td></tr>
        </tbody>
        <tbody v-else>
          <tr
            v-for="item in tree"
            :key="item.id"
            class="border-b border-[var(--border-subtle)] hover:bg-[var(--bg-hover)] transition-colors cursor-pointer"
            @click="navigateTo(`/admin/categories/${item.slug}`)"
          >
            <td class="px-6 py-4">
              <span :style="{ paddingLeft: `${item.depth * 24}px` }" class="flex items-center gap-2">
                <span v-if="item.depth > 0" class="text-muted">&lfloor;</span>
                <span class="font-medium">{{ item.name }}</span>
              </span>
            </td>
            <td class="px-6 py-4 font-mono text-xs text-secondary">{{ item.slug }}</td>
            <td class="px-6 py-4">
              <span class="text-xs" :class="item.depth === 0 ? 'text-accent' : 'text-secondary'">
                {{ item.depth === 0 ? 'Root' : 'Sub' }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get } = useAdminApi()
const { hasPermission } = useAdminAuth()

interface Category {
  id: string
  name: string
  slug: string
  parent_id: string | null
}

interface TreeItem extends Category {
  depth: number
}

const categories = ref<Category[]>([])
const loading = ref(true)

const tree = computed<TreeItem[]>(() => {
  const roots = categories.value.filter(c => !c.parent_id)
  const children = categories.value.filter(c => c.parent_id)
  const result: TreeItem[] = []

  for (const root of roots.sort((a, b) => a.name.localeCompare(b.name))) {
    result.push({ ...root, depth: 0 })
    const kids = children
      .filter(c => c.parent_id === root.id)
      .sort((a, b) => a.name.localeCompare(b.name))
    for (const kid of kids) {
      result.push({ ...kid, depth: 1 })
    }
  }
  return result
})

onMounted(async () => {
  try {
    const resp = await get<any>('/catalog/categories?per_page=100')
    categories.value = resp.items
  } catch {} finally {
    loading.value = false
  }
})
</script>
