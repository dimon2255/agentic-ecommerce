<template>
  <ClientOnly>
    <!-- Mobile backdrop -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="isOpen"
          class="fixed inset-0 bg-black/40 z-[59] lg:hidden"
          @click="close"
          aria-hidden="true"
        />
      </Transition>

      <!-- Panel -->
      <Transition name="slide-panel">
        <div
          v-if="isOpen"
          ref="panelRef"
          role="dialog"
          :aria-modal="isMobile ? 'true' : undefined"
          aria-label="Shopping assistant"
          class="fixed right-0 top-0 lg:top-16 bottom-0 z-[60] w-full lg:w-[420px] bg-surface-base lg:border-l border-[var(--border-default)] flex flex-col will-change-transform"
          @keydown.escape="close"
        >
          <!-- Header -->
          <div class="flex items-center justify-between px-4 h-14 border-b border-[var(--border-default)] glass-strong shrink-0">
            <h2 class="text-sm font-display font-bold text-[var(--text-primary)]">Shopping Assistant</h2>
            <div class="flex items-center gap-2">
              <button
                v-if="messages.length"
                @click="clearChat"
                class="text-xs text-muted hover:text-[var(--text-primary)] transition-colors px-2 py-1 rounded hover:bg-surface-hover"
              >
                New chat
              </button>
              <button
                @click="close"
                aria-label="Close assistant"
                class="w-8 h-8 flex items-center justify-center rounded-lg text-muted hover:text-[var(--text-primary)] hover:bg-surface-hover transition-colors"
              >
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Chat area -->
            <div
              ref="messagesContainer"
              class="flex-1 overflow-y-auto px-4 py-4 space-y-3"
              aria-live="polite"
              aria-relevant="additions"
            >
              <!-- Welcome state -->
              <div v-if="!messages.length" class="flex flex-col items-center justify-center h-full text-center px-2">
                <p class="text-base text-[var(--text-primary)] font-medium mb-1">How can I help you shop?</p>
                <p class="text-xs text-muted mb-5">Ask about products, comparisons, or recommendations</p>
                <div class="flex flex-wrap justify-center gap-2">
                  <button
                    v-for="suggestion in suggestions"
                    :key="suggestion"
                    @click="handleSuggestion(suggestion)"
                    class="px-3 py-1.5 text-xs rounded-full border border-[var(--border-default)] text-secondary hover:text-[var(--text-primary)] hover:border-accent transition-colors"
                  >
                    {{ suggestion }}
                  </button>
                </div>
              </div>

              <!-- Messages -->
              <div
                v-for="msg in messages"
                :key="msg.id"
                :class="[
                  'max-w-[90%] px-3 py-2.5 rounded-2xl text-sm leading-relaxed',
                  msg.role === 'user'
                    ? 'ml-auto bg-accent/20 text-[var(--text-primary)] rounded-br-md'
                    : 'mr-auto bg-surface-elevated border border-[var(--border-default)] rounded-bl-md'
                ]"
              >
                <!-- Thinking indicator -->
                <div v-if="msg.status === 'thinking' && !msg.content" class="flex items-center gap-2 text-muted">
                  <span class="flex gap-1">
                    <span class="w-1.5 h-1.5 bg-accent rounded-full animate-bounce" style="animation-delay: 0ms" />
                    <span class="w-1.5 h-1.5 bg-accent rounded-full animate-bounce" style="animation-delay: 150ms" />
                    <span class="w-1.5 h-1.5 bg-accent rounded-full animate-bounce" style="animation-delay: 300ms" />
                  </span>
                  <span class="text-xs">Searching products...</span>
                </div>

                <!-- Content -->
                <template v-else>
                  <div v-if="msg.role === 'user'" class="whitespace-pre-wrap text-sm" v-text="msg.content" />
                  <div v-else class="prose-chat" v-html="renderMarkdown(msg.content)" />
                  <span v-if="msg.status === 'streaming'" class="inline-block w-1.5 h-4 bg-accent animate-pulse ml-0.5 align-text-bottom" />
                </template>

                <!-- Product cards from tool results -->
                <div
                  v-if="msg.toolResults?.length && getProducts(msg.toolResults).length"
                  class="mt-2.5 -mx-1 flex gap-2 overflow-x-auto pb-1 scrollbar-thin"
                >
                  <ChatProductCard
                    v-for="product in getProducts(msg.toolResults)"
                    :key="product.id"
                    :product="product"
                    @click="close"
                  />
                </div>
              </div>

              <!-- Error -->
              <div v-if="error" class="mr-auto max-w-[90%] px-3 py-2.5 rounded-2xl text-xs bg-[var(--color-error-bg)] text-[var(--color-error-border)] border border-[var(--color-error-border)]">
                {{ error }}
              </div>
            </div>

            <!-- Input bar -->
            <div class="shrink-0 px-3 pb-3 pt-2 border-t border-[var(--border-default)]">
              <div class="glass-strong rounded-xl border border-[var(--border-default)] p-1.5 flex gap-1.5">
                <input
                  ref="inputRef"
                  v-model="input"
                  @keydown.enter.prevent="handleSend"
                  :disabled="loading"
                  type="text"
                  placeholder="Ask me anything..."
                  class="flex-1 bg-transparent px-2.5 py-1.5 text-sm text-[var(--text-primary)] placeholder:text-muted focus:outline-none"
                />
                <button
                  @click="handleSend"
                  :disabled="loading || !input.trim()"
                  class="btn-accent px-3 py-1.5 rounded-lg text-xs disabled:opacity-40"
                  aria-label="Send message"
                >
                  Send
                </button>
              </div>
            </div>
        </div>
      </Transition>
    </Teleport>
  </ClientOnly>
</template>

<script setup lang="ts">
import { marked } from 'marked'

marked.setOptions({ breaks: true, gfm: true })

function renderMarkdown(content: string): string {
  if (!content) return ''
  return marked.parse(content) as string
}

interface ProductDisplay {
  id: string
  name: string
  slug: string
  base_price: number
  images?: string[]
}

function getProducts(toolResults: Array<{ tool: string; result: any }>): ProductDisplay[] {
  const products: ProductDisplay[] = []
  const seen = new Set<string>()

  for (const tr of toolResults) {
    if (tr.tool === 'search_products' && tr.result?.products) {
      for (const p of tr.result.products) {
        if (!seen.has(p.id)) {
          seen.add(p.id)
          products.push(p)
        }
      }
    } else if (tr.tool === 'get_product_details' && tr.result?.product) {
      const p = tr.result.product
      if (!seen.has(p.id)) {
        seen.add(p.id)
        products.push(p)
      }
    }
  }

  return products
}

const user = useSupabaseUser()
const { isOpen, close } = useAssistantPanel()
const { messages, loading, error, sendMessage, clearChat } = useAssistant()
const isGuest = computed(() => !user.value)

const panelRef = ref<HTMLElement | null>(null)
const messagesContainer = ref<HTMLElement | null>(null)
const inputRef = ref<HTMLInputElement | null>(null)
const input = ref('')
const isMobile = ref(false)

const suggestions = computed(() => {
  const base = ['Best laptops under $1500', 'Show me headphones']
  if (!isGuest.value) base.push("What's in my cart?")
  return base
})

function handleSuggestion(text: string) {
  input.value = ''
  sendMessage(text)
  scrollToBottom()
}

async function handleSend() {
  const msg = input.value.trim()
  if (!msg) return
  input.value = ''
  await sendMessage(msg)
  scrollToBottom()
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

function checkMobile() {
  isMobile.value = window.innerWidth < 1024
}

// Focus trap
function handleFocusTrap(e: KeyboardEvent) {
  if (e.key !== 'Tab' || !panelRef.value) return

  const focusable = panelRef.value.querySelectorAll<HTMLElement>(
    'button:not([disabled]), input:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])'
  )
  if (!focusable.length) return

  const first = focusable[0]
  const last = focusable[focusable.length - 1]

  if (e.shiftKey && document.activeElement === first) {
    e.preventDefault()
    last.focus()
  } else if (!e.shiftKey && document.activeElement === last) {
    e.preventDefault()
    first.focus()
  }
}

// Focus input when panel opens
watch(isOpen, (open) => {
  if (open) {
    nextTick(() => {
      inputRef.value?.focus()
      document.addEventListener('keydown', handleFocusTrap)
    })
  } else {
    document.removeEventListener('keydown', handleFocusTrap)
  }
})

watch(messages, () => scrollToBottom(), { deep: true })

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  document.removeEventListener('keydown', handleFocusTrap)
})
</script>

<style scoped>
.prose-chat :deep(h1),
.prose-chat :deep(h2),
.prose-chat :deep(h3) {
  font-weight: 700;
  margin-top: 0.5rem;
  margin-bottom: 0.15rem;
  color: var(--text-primary);
}
.prose-chat :deep(h2) { font-size: 0.95rem; }
.prose-chat :deep(h3) { font-size: 0.875rem; }
.prose-chat :deep(p) { margin: 0.25rem 0; }
.prose-chat :deep(ul),
.prose-chat :deep(ol) {
  margin: 0.2rem 0;
  padding-left: 1.1rem;
}
.prose-chat :deep(li) { margin: 0.1rem 0; }
.prose-chat :deep(strong) { color: var(--text-primary); }
.prose-chat :deep(em) { color: var(--text-secondary); }
.prose-chat :deep(hr) {
  border-color: var(--border-default);
  margin: 0.4rem 0;
}
.prose-chat :deep(code) {
  background: var(--surface-elevated);
  padding: 0.1rem 0.25rem;
  border-radius: 0.2rem;
  font-size: 0.8em;
}
.prose-chat :deep(a) {
  color: var(--color-accent);
  text-decoration: underline;
}

.scrollbar-thin::-webkit-scrollbar {
  height: 4px;
}
.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}
.scrollbar-thin::-webkit-scrollbar-thumb {
  background: var(--border-default);
  border-radius: 2px;
}

/* Slide panel transition */
.slide-panel-enter-active,
.slide-panel-leave-active {
  transition: transform 0.3s ease-out;
}
.slide-panel-enter-from,
.slide-panel-leave-to {
  transform: translateX(100%);
}

/* Fade transition for backdrop */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
