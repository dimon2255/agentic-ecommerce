<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div v-if="product">
      <nav class="text-sm text-gray-500 mb-6">
        <NuxtLink to="/catalog" class="hover:text-gray-700">Catalog</NuxtLink>
        <span class="mx-2">/</span>
        <span class="text-gray-900">{{ product.name }}</span>
      </nav>
      <div class="grid md:grid-cols-2 gap-12">
        <div class="aspect-square bg-gray-100 rounded-xl flex items-center justify-center overflow-hidden">
          <img v-if="product.images?.length" :src="product.images[0]" :alt="product.name" class="w-full h-full object-cover" />
          <span v-else class="text-gray-300 text-8xl">&#9744;</span>
        </div>
        <div>
          <h1 class="text-3xl font-bold text-gray-900">{{ product.name }}</h1>
          <p v-if="product.description" class="mt-3 text-gray-600 leading-relaxed">{{ product.description }}</p>
          <div class="mt-6">
            <PriceDisplay :base-price="product.base_price" :price-override="selectedSku?.price_override" />
          </div>
          <div v-if="skus?.length && attributes?.length" class="mt-8">
            <SkuSelector :skus="skus" :attributes="formattedAttributes" @select="onSkuSelect" />
          </div>
          <button
            class="mt-8 w-full bg-primary-600 text-white py-3 px-6 rounded-lg font-medium hover:bg-primary-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
            :disabled="!selectedSku"
          >
            {{ selectedSku ? 'Add to Cart' : 'Select options' }}
          </button>
        </div>
      </div>
    </div>
    <div v-else>
      <p class="text-gray-500">Product not found.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const { get } = useApi()
const slug = route.params.slug as string

const { data: product } = await useAsyncData(`product-${slug}`, () =>
  get<any>(`/products/${slug}`)
)

const { data: skus } = await useAsyncData(`skus-${slug}`, async () => {
  if (!product.value) return []
  return get<any[]>(`/products/${product.value.id}/skus`)
})

const { data: attributes } = await useAsyncData(`attrs-${slug}`, async () => {
  if (!product.value) return []
  return get<any[]>(`/categories/${product.value.category_id}/attributes`)
})

const formattedAttributes = computed(() => {
  if (!attributes.value) return []
  return attributes.value.map((attr: any) => ({
    id: attr.id,
    name: attr.name,
    options: attr.options?.map((o: any) => o.value) || [],
  }))
})

const selectedSku = ref<any>(null)

function onSkuSelect(sku: any) {
  selectedSku.value = sku
}
</script>
