<template>
  <div class="max-w-3xl space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-display font-bold">Edit Category</h1>
        <p class="text-sm text-secondary mt-1">{{ category?.name }}</p>
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
      <div v-if="loadingCategory" class="space-y-4">
        <div v-for="i in 3" :key="i" class="h-10 bg-[var(--bg-hover)] rounded animate-pulse" />
      </div>
      <AdminCategoryForm
        v-else
        :form="form"
        :categories="parentOptions"
        :saving="saving"
        submit-label="Save Changes"
        @submit="handleUpdate"
      />
    </div>

    <!-- Attributes tab -->
    <div v-if="activeTab === 'attributes'" class="space-y-4">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-display font-semibold">Attributes</h2>
        <button
          v-if="hasPermission('catalog:write')"
          class="text-sm text-accent hover:text-accent-hover transition-colors"
          @click="showAttrForm = !showAttrForm"
        >
          {{ showAttrForm ? 'Cancel' : '+ Add Attribute' }}
        </button>
      </div>

      <!-- New attribute form -->
      <div v-if="showAttrForm" class="card-dark p-5 space-y-4">
        <div class="grid grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium text-secondary mb-1.5">Name</label>
            <input v-model="newAttr.name" class="input-dark text-sm" placeholder="e.g. Color" />
          </div>
          <div>
            <label class="block text-sm font-medium text-secondary mb-1.5">Type</label>
            <select v-model="newAttr.type" class="input-dark text-sm">
              <option value="text">Text</option>
              <option value="number">Number</option>
              <option value="enum">Enum</option>
            </select>
          </div>
          <div class="flex items-end">
            <label class="flex items-center gap-2 text-sm text-secondary cursor-pointer">
              <input v-model="newAttr.required" type="checkbox" class="accent-[var(--accent)]" />
              Required
            </label>
          </div>
        </div>
        <button class="btn-accent px-4 py-2 rounded-lg text-sm" @click="handleCreateAttribute">
          Create Attribute
        </button>
      </div>

      <!-- Attributes list -->
      <div class="card-dark overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-[var(--border-default)]">
              <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Name</th>
              <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Type</th>
              <th class="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-muted">Required</th>
              <th class="px-5 py-3 w-16" />
            </tr>
          </thead>
          <tbody v-if="attributes.length === 0">
            <tr><td colspan="4" class="px-5 py-8 text-center text-muted">No attributes yet</td></tr>
          </tbody>
          <tbody v-else>
            <tr v-for="attr in attributes" :key="attr.id" class="border-b border-[var(--border-subtle)]">
              <td class="px-5 py-3 font-medium">{{ attr.name }}</td>
              <td class="px-5 py-3 text-secondary">{{ attr.type }}</td>
              <td class="px-5 py-3">
                <span :class="attr.required ? 'text-accent' : 'text-muted'">{{ attr.required ? 'Yes' : 'No' }}</span>
              </td>
              <td class="px-5 py-3">
                <button
                  v-if="hasPermission('catalog:write')"
                  class="text-xs text-[var(--color-error)] hover:text-[var(--color-error)]/80"
                  @click="handleDeleteAttribute(attr.id)"
                >
                  Delete
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <AdminConfirmDialog
      :open="showDelete"
      title="Delete Category"
      message="This will delete this category. Products in this category will be affected."
      confirm-text="Delete"
      variant="danger"
      @confirm="handleDelete"
      @cancel="showDelete = false"
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
  { key: 'attributes', label: 'Attributes' },
]
const activeTab = ref('details')

const category = ref<any>(null)
const loadingCategory = ref(true)
const parentOptions = ref<any[]>([])
const saving = ref(false)
const showDelete = ref(false)

const form = reactive({
  name: '',
  slug: '',
  parent_id: null as string | null,
})

// Attributes state
const attributes = ref<any[]>([])
const showAttrForm = ref(false)
const newAttr = reactive({ name: '', type: 'text', required: false, sort_order: 0 })

onMounted(async () => {
  await Promise.all([fetchCategory(), fetchParentOptions()])
})

async function fetchCategory() {
  loadingCategory.value = true
  try {
    const cat = await get<any>(`/catalog/categories/${slug}`)
    category.value = cat
    Object.assign(form, { name: cat.name, slug: cat.slug, parent_id: cat.parent_id })
    await fetchAttributes()
  } catch {
    showToast('Category not found', 'error')
    router.push('/admin/categories')
  } finally {
    loadingCategory.value = false
  }
}

async function fetchParentOptions() {
  try {
    const resp = await get<any>('/catalog/categories?per_page=100')
    parentOptions.value = resp.items.filter((c: any) => !c.parent_id && c.slug !== slug)
  } catch {}
}

async function fetchAttributes() {
  if (!category.value) return
  try {
    attributes.value = await get<any[]>(`/catalog/categories/${category.value.id}/attributes`)
  } catch {
    attributes.value = []
  }
}

async function handleUpdate() {
  saving.value = true
  try {
    const payload: any = { name: form.name, slug: form.slug }
    if (form.parent_id) payload.parent_id = form.parent_id
    await patch(`/catalog/categories/${slug}`, payload)
    showToast('Category updated', 'success')
    if (form.slug !== slug) {
      router.push(`/admin/categories/${form.slug}`)
    }
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to update', 'error')
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  try {
    await del(`/catalog/categories/${slug}`)
    showToast('Category deleted', 'success')
    router.push('/admin/categories')
  } catch {
    showToast('Failed to delete category', 'error')
  }
  showDelete.value = false
}

async function handleCreateAttribute() {
  try {
    await post(`/catalog/categories/${category.value.id}/attributes`, { ...newAttr })
    showToast('Attribute created', 'success')
    showAttrForm.value = false
    newAttr.name = ''
    newAttr.type = 'text'
    newAttr.required = false
    await fetchAttributes()
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to create attribute', 'error')
  }
}

async function handleDeleteAttribute(attrId: string) {
  try {
    await del(`/catalog/attributes/${attrId}`)
    showToast('Attribute deleted', 'success')
    await fetchAttributes()
  } catch {
    showToast('Failed to delete attribute', 'error')
  }
}
</script>
