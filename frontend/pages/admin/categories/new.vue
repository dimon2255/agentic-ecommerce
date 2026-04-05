<template>
  <div class="max-w-2xl space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">New Category</h1>
      <p class="text-sm text-secondary mt-1">Add a new product category</p>
    </div>

    <div class="card-dark p-6">
      <CategoryForm
        :form="form"
        :categories="parentOptions"
        :saving="saving"
        submit-label="Create Category"
        @submit="handleCreate"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({ layout: 'admin', middleware: 'admin' })

const { get, post } = useAdminApi()
const { showToast } = useToast()
const router = useRouter()

const form = reactive({
  name: '',
  slug: '',
  parent_id: null as string | null,
})

const parentOptions = ref<any[]>([])
const saving = ref(false)

onMounted(async () => {
  try {
    const resp = await get<any>('/catalog/categories?per_page=100')
    // Only root categories as potential parents
    parentOptions.value = resp.items.filter((c: any) => !c.parent_id)
  } catch {}
})

async function handleCreate() {
  saving.value = true
  try {
    const payload: any = { name: form.name, slug: form.slug }
    if (form.parent_id) payload.parent_id = form.parent_id
    await post('/catalog/categories', payload)
    showToast('Category created', 'success')
    router.push('/admin/categories')
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to create category', 'error')
  } finally {
    saving.value = false
  }
}
</script>
