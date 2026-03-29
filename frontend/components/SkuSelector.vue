<template>
  <div class="space-y-4">
    <div v-for="attr in attributes" :key="attr.id">
      <label class="block text-sm font-medium text-gray-700 mb-1">{{ attr.name }}</label>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="option in attr.options" :key="option"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium border transition-colors',
            selectedValues[attr.name] === option
              ? 'border-primary-500 bg-primary-50 text-primary-700'
              : isOptionAvailable(attr.name, option)
                ? 'border-gray-200 bg-white text-gray-700 hover:border-gray-300'
                : 'border-gray-100 bg-gray-50 text-gray-300 cursor-not-allowed'
          ]"
          :disabled="!isOptionAvailable(attr.name, option)"
          @click="selectOption(attr.name, option)"
        >
          {{ option }}
        </button>
      </div>
    </div>
    <div v-if="selectedSku" class="pt-2">
      <p class="text-sm text-gray-500">SKU: {{ selectedSku.sku_code }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface SKU {
  id: string
  sku_code: string
  price_override: number | null
  attribute_values: Array<{ category_attribute_id: string; value: string }>
}

interface Attribute {
  id: string
  name: string
  options: string[]
}

const props = defineProps<{
  skus: SKU[]
  attributes: Attribute[]
}>()

const emit = defineEmits<{
  select: [sku: SKU | null]
}>()

const selectedValues = reactive<Record<string, string>>({})

function selectOption(attrName: string, value: string) {
  selectedValues[attrName] = value
  emit('select', selectedSku.value)
}

function isOptionAvailable(attrName: string, option: string): boolean {
  return props.skus.some(sku => {
    const attrMap = buildAttrMap(sku)
    if (attrMap[attrName] !== option) return false
    for (const [name, val] of Object.entries(selectedValues)) {
      if (name !== attrName && attrMap[name] !== val) return false
    }
    return true
  })
}

function buildAttrMap(sku: SKU): Record<string, string> {
  const map: Record<string, string> = {}
  for (const av of sku.attribute_values) {
    const attr = props.attributes.find(a => a.id === av.category_attribute_id)
    if (attr) map[attr.name] = av.value
  }
  return map
}

const selectedSku = computed<SKU | null>(() => {
  if (Object.keys(selectedValues).length !== props.attributes.length) return null
  return props.skus.find(sku => {
    const attrMap = buildAttrMap(sku)
    return Object.entries(selectedValues).every(([name, val]) => attrMap[name] === val)
  }) || null
})
</script>
