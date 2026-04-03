interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  product_ids: string[]
  created_at: string
  status?: 'streaming' | 'thinking' | 'complete'
}

interface ChatResponse {
  session_id: string
  message: ChatMessage
}

export function useAssistant() {
  const { post } = useApi()
  const config = useRuntimeConfig()
  const supabase = useSupabaseClient()

  const messages = useState<ChatMessage[]>('assistant-messages', () => [])
  const sessionId = useState<string | null>('assistant-session', () => null)
  const loading = useState('assistant-loading', () => false)
  const error = useState<string | null>('assistant-error', () => null)

  async function getHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {}

    // Try reactive session state first
    const authSession = useSupabaseSession()
    if (authSession.value?.access_token) {
      headers['Authorization'] = `Bearer ${authSession.value.access_token}`
      return headers
    }

    // Fallback: fetch session from Supabase client
    try {
      const { data } = await supabase.auth.getSession()
      if (data.session?.access_token) {
        headers['Authorization'] = `Bearer ${data.session.access_token}`
      }
    } catch {
      // No auth available
    }
    return headers
  }

  /** Non-streaming send (Phase 1 fallback). */
  async function sendMessageSync(content: string) {
    if (!content.trim() || loading.value) return

    loading.value = true
    error.value = null

    messages.value = [...messages.value, {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      product_ids: [],
      created_at: new Date().toISOString(),
      status: 'complete',
    }]

    try {
      const headers = await getHeaders()
      const body: Record<string, string> = { message: content }
      if (sessionId.value) {
        body.session_id = sessionId.value
      }

      const res = await post<ChatResponse>('/assistant', body, headers)
      sessionId.value = res.session_id
      messages.value = [...messages.value, { ...res.message, status: 'complete' }]
    } catch (err: any) {
      error.value = err?.data?.error?.message || err?.message || 'Failed to send message'
    } finally {
      loading.value = false
    }
  }

  /** Streaming send via SSE (Phase 2). */
  async function sendMessage(content: string) {
    if (!content.trim() || loading.value) return

    loading.value = true
    error.value = null

    // Optimistic user message
    messages.value = [...messages.value, {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      product_ids: [],
      created_at: new Date().toISOString(),
      status: 'complete',
    }]

    // Placeholder assistant message
    const placeholderId = crypto.randomUUID()
    messages.value = [...messages.value, {
      id: placeholderId,
      role: 'assistant',
      content: '',
      product_ids: [],
      created_at: new Date().toISOString(),
      status: 'streaming',
    }]

    try {
      const headers = await getHeaders()
      headers['Content-Type'] = 'application/json'

      const body: Record<string, string> = { message: content }
      if (sessionId.value) {
        body.session_id = sessionId.value
      }

      const baseURL = config.public.apiBase || 'http://localhost:9090'
      const res = await fetch(`${baseURL}/api/v1/assistant/stream/`, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
      })

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }

      const reader = res.body!.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let cartUpdated = false
      let eventType = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) {
          // Process any remaining data in the buffer
          if (buffer.trim()) {
            const remaining = buffer.split('\n')
            for (const line of remaining) {
              if (line.startsWith('event: ')) {
                eventType = line.slice(7)
              } else if (line.startsWith('data: ') && eventType) {
                handleSSEEvent(eventType, line.slice(6), placeholderId)
                if (eventType === 'done') {
                  try {
                    const payload = JSON.parse(line.slice(6))
                    if (payload.cart_updated) cartUpdated = true
                  } catch {}
                }
                eventType = ''
              }
            }
          }
          // Ensure status is complete even if done event was missed
          updatePlaceholder(placeholderId, { status: 'complete' })
          break
        }

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            eventType = line.slice(7)
          } else if (line.startsWith('data: ') && eventType) {
            const data = line.slice(6)
            handleSSEEvent(eventType, data, placeholderId)

            if (eventType === 'done') {
              try {
                const payload = JSON.parse(data)
                if (payload.cart_updated) cartUpdated = true
              } catch {}
            }
            eventType = ''
          }
        }
      }

      // Refresh cart if tools modified it
      if (cartUpdated) {
        try {
          const { refresh } = useCart()
          await refresh(true)
        } catch {}
      }
    } catch (err: any) {
      error.value = err?.message || 'Failed to stream response'
      // Remove placeholder if it's still empty
      const placeholder = messages.value.find(m => m.id === placeholderId)
      if (placeholder && !placeholder.content) {
        messages.value = messages.value.filter(m => m.id !== placeholderId)
      }
    } finally {
      loading.value = false
    }
  }

  function handleSSEEvent(event: string, data: string, placeholderId: string) {
    try {
      const payload = JSON.parse(data)

      switch (event) {
        case 'session':
          sessionId.value = payload.session_id
          break

        case 'status':
          updatePlaceholder(placeholderId, { status: 'thinking' })
          break

        case 'tool_start':
          updatePlaceholder(placeholderId, { status: 'thinking' })
          break

        case 'delta':
          if (payload.text) {
            const msg = messages.value.find(m => m.id === placeholderId)
            if (msg) {
              updatePlaceholder(placeholderId, {
                content: msg.content + payload.text,
                status: 'streaming',
              })
            }
          }
          break

        case 'done':
          updatePlaceholder(placeholderId, { status: 'complete' })
          break

        case 'error':
          error.value = payload.message || 'An error occurred'
          updatePlaceholder(placeholderId, { status: 'complete' })
          break
      }
    } catch {
      // Ignore malformed events
    }
  }

  function updatePlaceholder(id: string, updates: Partial<ChatMessage>) {
    messages.value = messages.value.map(m =>
      m.id === id ? { ...m, ...updates } : m,
    )
  }

  function clearChat() {
    messages.value = []
    sessionId.value = null
    error.value = null
  }

  return { messages, loading, error, sendMessage, sendMessageSync, clearChat }
}
