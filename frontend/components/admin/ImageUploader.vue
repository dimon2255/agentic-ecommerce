<template>
  <div class="space-y-3">
    <label class="block text-sm font-medium text-secondary">Images</label>

    <!-- Thumbnail grid -->
    <div v-if="modelValue.length" class="flex flex-wrap gap-3">
      <div
        v-for="(url, idx) in modelValue"
        :key="url"
        class="relative w-20 h-20 rounded-lg overflow-hidden border border-[var(--border-default)] group"
      >
        <img :src="url" class="w-full h-full object-cover" />
        <button
          type="button"
          class="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center text-white text-xs"
          @click="remove(idx)"
        >
          Remove
        </button>
      </div>
    </div>

    <!-- Drop zone -->
    <div
      v-if="modelValue.length < max"
      class="border-2 border-dashed rounded-lg p-6 text-center transition-colors cursor-pointer"
      :class="dragOver
        ? 'border-accent bg-accent/5'
        : 'border-[var(--border-default)] hover:border-[var(--border-strong)]'"
      @dragover.prevent="dragOver = true"
      @dragleave="dragOver = false"
      @drop.prevent="handleDrop"
      @click="fileInput?.click()"
    >
      <p v-if="uploading" class="text-sm text-accent">Uploading...</p>
      <template v-else>
        <p class="text-sm text-secondary">Drop images here or click to browse</p>
        <p class="text-xs text-muted mt-1">PNG, JPG, WebP (max {{ max }} images)</p>
      </template>
    </div>

    <input
      ref="fileInput"
      type="file"
      accept="image/*"
      multiple
      class="hidden"
      @change="handleFileSelect"
    />

    <p v-if="error" class="text-xs text-[var(--color-error)]">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: string[]
  max?: number
}>(), {
  max: 10,
})

const emit = defineEmits<{
  'update:modelValue': [urls: string[]]
}>()

const { post } = useAdminApi()

const fileInput = ref<HTMLInputElement | null>(null)
const dragOver = ref(false)
const uploading = ref(false)
const error = ref('')

function remove(idx: number) {
  const updated = [...props.modelValue]
  updated.splice(idx, 1)
  emit('update:modelValue', updated)
}

async function handleDrop(e: DragEvent) {
  dragOver.value = false
  const files = Array.from(e.dataTransfer?.files || [])
  await uploadFiles(files)
}

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement
  const files = Array.from(input.files || [])
  uploadFiles(files)
  input.value = ''
}

async function uploadFiles(files: File[]) {
  const remaining = props.max - props.modelValue.length
  const toUpload = files.slice(0, remaining).filter(f => f.type.startsWith('image/'))
  if (!toUpload.length) return

  uploading.value = true
  error.value = ''

  const newUrls: string[] = []

  for (const file of toUpload) {
    try {
      // Get presigned upload URL
      const { upload_url, public_url } = await post<{ upload_url: string; public_url: string }>(
        '/images/upload-url',
        { filename: file.name, content_type: file.type },
      )

      // Upload directly to Supabase Storage
      const resp = await fetch(upload_url, {
        method: 'PUT',
        headers: { 'Content-Type': file.type },
        body: file,
      })

      if (!resp.ok) throw new Error('Upload failed')
      newUrls.push(public_url)
    } catch {
      error.value = `Failed to upload ${file.name}`
    }
  }

  if (newUrls.length) {
    emit('update:modelValue', [...props.modelValue, ...newUrls])
  }

  uploading.value = false
}
</script>
