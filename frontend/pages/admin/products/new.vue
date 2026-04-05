<template>
  <div class="max-w-2xl space-y-6">
    <div>
      <h1 class="text-2xl font-display font-bold">New Product</h1>
      <p class="text-sm text-secondary mt-1">Add a new product to your catalog</p>
    </div>

    <div class="card-dark p-6">
      <AdminProductForm
        :form="form"
        :categories="categories"
        :saving="saving"
        submit-label="Create Product"
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
  category_id: '',
  description: '',
  base_price: 0,
  status: 'draft',
  images: [] as string[],
})

const categories = ref<any[]>([])
const saving = ref(false)

onMounted(async () => {
  try {
    const resp = await get<any>('/catalog/categories?per_page=100')
    categories.value = resp.items
  } catch {}
})

async function handleCreate() {
  saving.value = true
  try {
    const payload: any = { ...form }
    if (!payload.description) payload.description = null
    await post('/catalog/products', payload)
    showToast('Product created', 'success')
    router.push('/admin/products')
  } catch (e: any) {
    showToast(e?.data?.error?.message || 'Failed to create product', 'error')
  } finally {
    saving.value = false
  }
}
</script>
