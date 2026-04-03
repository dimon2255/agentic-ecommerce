<template>
  <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12 flex flex-col" style="height: calc(100vh - 4rem)">
    <div class="flex items-center justify-between mb-6 animate-fade-in-up">
      <h1 class="text-2xl font-display font-bold text-[var(--text-primary)]">Shopping Assistant</h1>
      <button
        v-if="messages.length"
        @click="clearChat"
        class="text-sm text-muted hover:text-[var(--text-primary)] transition-colors"
      >
        New chat
      </button>
    </div>

    <!-- Auth gate -->
    <ClientOnly>
      <div v-if="!user" class="text-center py-20 animate-fade-in">
        <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-surface-elevated border border-[var(--border-default)] flex items-center justify-center">
          <svg class="w-7 h-7 text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
          </svg>
        </div>
        <p class="text-muted mb-4">Sign in to use the shopping assistant</p>
        <NuxtLink to="/auth/login" class="text-accent hover:text-accent-hover font-medium transition-colors">
          Sign in
        </NuxtLink>
      </div>

      <!-- Chat area -->
      <template v-else>
        <div ref="messagesContainer" class="flex-1 overflow-y-auto space-y-4 mb-4 animate-fade-in">
          <!-- Welcome state -->
          <div v-if="!messages.length" class="text-center py-16">
            <p class="text-lg text-[var(--text-primary)] font-medium mb-2">How can I help you shop today?</p>
            <p class="text-sm text-muted mb-6">Ask me about products, comparisons, or recommendations</p>
            <div class="flex flex-wrap justify-center gap-2">
              <button
                v-for="suggestion in suggestions"
                :key="suggestion"
                @click="sendMessage(suggestion)"
                class="px-4 py-2 text-sm rounded-full border border-[var(--border-default)] text-secondary hover:text-[var(--text-primary)] hover:border-accent transition-colors"
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
              'max-w-[85%] px-4 py-3 rounded-2xl text-sm leading-relaxed',
              msg.role === 'user'
                ? 'ml-auto bg-accent/20 text-[var(--text-primary)] rounded-br-md'
                : 'mr-auto card-dark rounded-bl-md'
            ]"
          >
            <!-- Thinking indicator (tool execution in progress) -->
            <div v-if="msg.status === 'thinking' && !msg.content" class="flex items-center gap-2 text-muted">
              <span class="flex gap-1">
                <span class="w-1.5 h-1.5 bg-accent rounded-full animate-bounce" style="animation-delay: 0ms" />
                <span class="w-1.5 h-1.5 bg-accent rounded-full animate-bounce" style="animation-delay: 150ms" />
                <span class="w-1.5 h-1.5 bg-accent rounded-full animate-bounce" style="animation-delay: 300ms" />
              </span>
              Searching products...
            </div>

            <!-- Content (streaming or complete) -->
            <template v-else>
              <div v-if="msg.role === 'user'" class="whitespace-pre-wrap" v-text="msg.content" />
              <div v-else class="prose-chat" v-html="renderMarkdown(msg.content)" />
              <span v-if="msg.status === 'streaming'" class="inline-block w-1.5 h-4 bg-accent animate-pulse ml-0.5 align-text-bottom" />
            </template>
          </div>

          <!-- Error -->
          <div v-if="error" class="mr-auto max-w-[85%] px-4 py-3 rounded-2xl text-sm bg-[var(--color-error-bg)] text-[var(--color-error-border)] border border-[var(--color-error-border)]">
            {{ error }}
          </div>
        </div>

        <!-- Input bar -->
        <div class="glass-strong rounded-xl border border-[var(--border-default)] p-2 flex gap-2">
          <input
            v-model="input"
            @keydown.enter.prevent="handleSend"
            :disabled="loading"
            type="text"
            placeholder="Ask me anything about our products..."
            class="flex-1 bg-transparent px-3 py-2 text-sm text-[var(--text-primary)] placeholder:text-muted focus:outline-none"
          />
          <button
            @click="handleSend"
            :disabled="loading || !input.trim()"
            class="btn-accent px-4 py-2 rounded-lg text-sm disabled:opacity-40"
            aria-label="Send message"
          >
            Send
          </button>
        </div>
      </template>
    </ClientOnly>
  </div>
</template>

<script setup lang="ts">
import { marked } from 'marked'

marked.setOptions({ breaks: true, gfm: true })

function renderMarkdown(content: string): string {
  if (!content) return ''
  return marked.parse(content) as string
}

const user = useSupabaseUser()
const { messages, loading, error, sendMessage, clearChat } = useAssistant()
const messagesContainer = ref<HTMLElement | null>(null)
const input = ref('')

const suggestions = [
  'Best laptops under $1500',
  'Compare noise-cancelling headphones',
  'Gift ideas under $100',
]

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

watch(messages, () => scrollToBottom(), { deep: true })
</script>

<style scoped>
.prose-chat :deep(h1),
.prose-chat :deep(h2),
.prose-chat :deep(h3) {
  font-weight: 700;
  margin-top: 0.75rem;
  margin-bottom: 0.25rem;
  color: var(--text-primary);
}
.prose-chat :deep(h2) { font-size: 1.05rem; }
.prose-chat :deep(h3) { font-size: 0.95rem; }
.prose-chat :deep(p) { margin: 0.35rem 0; }
.prose-chat :deep(ul),
.prose-chat :deep(ol) {
  margin: 0.25rem 0;
  padding-left: 1.25rem;
}
.prose-chat :deep(li) { margin: 0.15rem 0; }
.prose-chat :deep(strong) { color: var(--text-primary); }
.prose-chat :deep(em) { color: var(--text-secondary); }
.prose-chat :deep(hr) {
  border-color: var(--border-default);
  margin: 0.5rem 0;
}
.prose-chat :deep(code) {
  background: var(--surface-elevated);
  padding: 0.1rem 0.3rem;
  border-radius: 0.25rem;
  font-size: 0.85em;
}
.prose-chat :deep(img) {
  max-width: 100%;
  border-radius: 0.5rem;
  margin: 0.5rem 0;
}
.prose-chat :deep(a) {
  color: var(--color-accent);
  text-decoration: underline;
}
</style>
