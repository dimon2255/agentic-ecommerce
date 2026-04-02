interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  product_ids: string[]
  created_at: string
}

interface ChatResponse {
  session_id: string
  message: ChatMessage
}

export function useAssistant() {
  const { post } = useApi()
  const client = useSupabaseClient()

  const messages = useState<ChatMessage[]>('assistant-messages', () => [])
  const sessionId = useState<string | null>('assistant-session', () => null)
  const loading = useState('assistant-loading', () => false)
  const error = useState<string | null>('assistant-error', () => null)

  async function getHeaders(): Promise<Record<string, string>> {
    const headers: Record<string, string> = {}
    try {
      const { data: { session } } = await client.auth.getSession()
      if (session?.access_token) {
        headers['Authorization'] = `Bearer ${session.access_token}`
      }
    } catch {
      // No auth available
    }
    return headers
  }

  async function sendMessage(content: string) {
    if (!content.trim() || loading.value) return

    loading.value = true
    error.value = null

    // Optimistically add user message
    messages.value = [...messages.value, {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      product_ids: [],
      created_at: new Date().toISOString(),
    }]

    try {
      const headers = await getHeaders()
      const body: Record<string, string> = { message: content }
      if (sessionId.value) {
        body.session_id = sessionId.value
      }

      const res = await post<ChatResponse>('/assistant', body, headers)
      sessionId.value = res.session_id
      messages.value = [...messages.value, res.message]
    } catch (err: any) {
      error.value = err?.data?.error?.message || err?.message || 'Failed to send message'
    } finally {
      loading.value = false
    }
  }

  function clearChat() {
    messages.value = []
    sessionId.value = null
    error.value = null
  }

  return { messages, loading, error, sendMessage, clearChat }
}
