<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div v-if="product" class="animate-fade-in">
      <nav aria-label="Breadcrumb" class="text-sm text-muted mb-6">
        <ol class="flex items-center gap-2">
          <li><NuxtLink to="/catalog" class="hover:text-secondary transition-colors">Catalog</NuxtLink></li>
          <li class="text-muted/50">/</li>
          <li v-if="product.categories" class="flex items-center gap-2">
            <NuxtLink :to="`/catalog/${product.categories.slug}`" class="hover:text-secondary transition-colors">{{ product.categories.name }}</NuxtLink>
            <span class="text-muted/50">/</span>
          </li>
          <li aria-current="page" class="text-secondary">{{ product.name }}</li>
        </ol>
      </nav>

      <div class="grid md:grid-cols-2 gap-12">
        <!-- Product Image -->
        <div class="aspect-square bg-surface-deep rounded-2xl overflow-hidden border border-[var(--border-default)] animate-fade-in-up">
          <img
            v-if="product.images?.length"
            :src="product.images[0]"
            :alt="product.name"
            class="w-full h-full object-cover"
          />
          <div v-else class="w-full h-full flex items-center justify-center">
            <svg class="w-20 h-20 text-muted/20" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="0.5">
              <path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
        </div>

        <!-- Product Info -->
        <div class="animate-fade-in-up delay-1">
          <h1 class="text-3xl font-display font-bold text-[var(--text-primary)]">{{ product.name }}</h1>
          <p v-if="product.description" class="mt-3 text-secondary leading-relaxed">{{ product.description }}</p>

          <div class="mt-6">
            <PriceDisplay :base-price="product.base_price" :price-override="selectedSku?.price_override" />
          </div>

          <div v-if="skus?.length && attributes?.length" class="mt-8 p-5 bg-surface rounded-xl border border-[var(--border-default)]">
            <SkuSelector :skus="skus" :attributes="formattedAttributes" @select="onSkuSelect" />
          </div>

          <button
            class="mt-8 w-full btn-accent py-3.5 px-6 rounded-xl text-sm tracking-wide"
            :disabled="!selectedSku || addingToCart"
            @click="addToCart"
          >
            {{ addingToCart ? 'Adding...' : selectedSku ? 'Add to Cart' : 'Select options' }}
          </button>

          <p
            v-if="addedMsg"
            role="alert"
            class="mt-3 text-sm text-center font-medium"
            :class="addedMsg === 'Added to cart!' ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]'"
          >
            {{ addedMsg }}
          </p>
        </div>
      </div>
    </div>
    <div v-else>
      <p class="text-muted">Product not found.</p>
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

const { addItem } = useCart()
const addingToCart = ref(false)
const addedMsg = ref('')

async function addToCart() {
  if (!selectedSku.value) return
  addingToCart.value = true
  addedMsg.value = ''
  try {
    await addItem(selectedSku.value.id)
    addedMsg.value = 'Added to cart!'
    setTimeout(() => { addedMsg.value = '' }, 2000)
  } catch {
    addedMsg.value = 'Failed to add to cart'
  } finally {
    addingToCart.value = false
  }
}
</script>
