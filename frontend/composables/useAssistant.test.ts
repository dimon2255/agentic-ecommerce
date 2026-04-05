import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createMockSupabaseClient } from './__mocks__/supabase'

const mockPost = vi.fn()
const mockCartRefresh = vi.fn()

mockNuxtImport('useApi', () => {
  return () => ({ get: vi.fn(), post: mockPost, patch: vi.fn(), del: vi.fn() })
})

mockNuxtImport('useRuntimeConfig', () => {
  return () => ({
    app: { baseURL: '/' },
    public: { apiBase: 'http://test-api', stripeKey: '' },
  })
})

let mockClient = createMockSupabaseClient({ accessToken: null })
mockNuxtImport('useSupabaseClient', () => () => mockClient)

mockNuxtImport('useSupabaseSession', () => () => ref(null))

mockNuxtImport('useCart', () => {
  return () => ({ refresh: mockCartRefresh })
})

import { useAssistant } from './useAssistant'

function createSSEResponse(events: Array<{ event: string; data: any }>) {
  const text = events
    .map(e => `event: ${e.event}\ndata: ${JSON.stringify(e.data)}\n\n`)
    .join('')
  const encoder = new TextEncoder()
  const stream = new ReadableStream({
    start(controller) {
      controller.enqueue(encoder.encode(text))
      controller.close()
    },
  })
  return new Response(stream, { status: 200 })
}

describe('useAssistant', () => {
  let originalFetch: typeof globalThis.fetch

  beforeEach(() => {
    mockPost.mockReset()
    mockCartRefresh.mockReset()
    mockClient = createMockSupabaseClient({ accessToken: null })
    localStorage.clear()
    originalFetch = globalThis.fetch

    // Reset shared state
    const { messages, loading, error } = useAssistant()
    messages.value = []
    loading.value = false
    error.value = null
  })

  afterEach(() => {
    globalThis.fetch = originalFetch
  })

  describe('sendMessageSync', () => {
    it('adds user message to messages array', async () => {
      mockPost.mockResolvedValue({
        session_id: 's1',
        message: { id: 'm1', role: 'assistant', content: 'Hi', product_ids: [], created_at: '' },
      })
      const { messages, sendMessageSync } = useAssistant()
      await sendMessageSync('Hello')
      expect(messages.value[0].role).toBe('user')
      expect(messages.value[0].content).toBe('Hello')
      expect(messages.value[0].status).toBe('complete')
    })

    it('posts to /assistant and appends response', async () => {
      mockPost.mockResolvedValue({
        session_id: 's1',
        message: { id: 'm1', role: 'assistant', content: 'Reply', product_ids: [], created_at: '' },
      })
      const { messages, sendMessageSync } = useAssistant()
      await sendMessageSync('Hi')
      expect(mockPost).toHaveBeenCalledWith('/assistant', expect.objectContaining({ message: 'Hi' }), expect.any(Object))
      expect(messages.value).toHaveLength(2)
      expect(messages.value[1].content).toBe('Reply')
    })

    it('stores session_id from response', async () => {
      mockPost.mockResolvedValue({
        session_id: 'new-session',
        message: { id: 'm1', role: 'assistant', content: 'Hi', product_ids: [], created_at: '' },
      })
      const assistant = useAssistant()
      await assistant.sendMessageSync('Test')
      // Session state is internal via useState - verify by sending again
      mockPost.mockResolvedValue({
        session_id: 'new-session',
        message: { id: 'm2', role: 'assistant', content: 'Hi2', product_ids: [], created_at: '' },
      })
      await assistant.sendMessageSync('Again')
      expect(mockPost).toHaveBeenLastCalledWith('/assistant', expect.objectContaining({ session_id: 'new-session' }), expect.any(Object))
    })

    it('ignores empty input', async () => {
      const { sendMessageSync } = useAssistant()
      await sendMessageSync('')
      await sendMessageSync('   ')
      expect(mockPost).not.toHaveBeenCalled()
    })

    it('does not send while loading', async () => {
      const { loading, sendMessageSync } = useAssistant()
      loading.value = true
      await sendMessageSync('Hello')
      expect(mockPost).not.toHaveBeenCalled()
    })

    it('sets error on failure', async () => {
      mockPost.mockRejectedValue({ message: 'Network error' })
      const { error, sendMessageSync } = useAssistant()
      await sendMessageSync('Fail')
      expect(error.value).toBe('Network error')
    })
  })

  describe('sendMessage (SSE streaming)', () => {
    it('creates user message and placeholder', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'done', data: {} },
      ]))
      const { messages, sendMessage } = useAssistant()
      await sendMessage('Hello')
      // User message + assistant message (placeholder that got completed)
      expect(messages.value.length).toBeGreaterThanOrEqual(2)
      expect(messages.value[0].role).toBe('user')
      expect(messages.value[0].content).toBe('Hello')
    })

    it('posts to correct SSE endpoint', async () => {
      const mockFetchFn = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'done', data: {} },
      ]))
      globalThis.fetch = mockFetchFn
      const { sendMessage } = useAssistant()
      await sendMessage('Test')
      expect(mockFetchFn).toHaveBeenCalledWith(
        'http://test-api/api/v1/assistant/stream/',
        expect.objectContaining({ method: 'POST' }),
      )
    })

    it('parses session event and stores session_id', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'session', data: { session_id: 'sse-session-1' } },
        { event: 'done', data: {} },
      ]))
      const assistant = useAssistant()
      await assistant.sendMessage('Hi')
      // Verify session stored by sending sync message
      mockPost.mockResolvedValue({
        session_id: 'sse-session-1',
        message: { id: 'x', role: 'assistant', content: '', product_ids: [], created_at: '' },
      })
      assistant.loading.value = false
      await assistant.sendMessageSync('Follow up')
      expect(mockPost).toHaveBeenCalledWith('/assistant', expect.objectContaining({ session_id: 'sse-session-1' }), expect.any(Object))
    })

    it('parses delta events and appends text', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'delta', data: { text: 'Hello ' } },
        { event: 'delta', data: { text: 'world' } },
        { event: 'done', data: {} },
      ]))
      const { messages, sendMessage } = useAssistant()
      await sendMessage('Hi')
      const assistantMsg = messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.content).toBe('Hello world')
    })

    it('parses tool_result event', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'tool_result', data: { tool: 'search_products', result: [{ id: 'p1' }] } },
        { event: 'done', data: {} },
      ]))
      const { messages, sendMessage } = useAssistant()
      await sendMessage('Find products')
      const assistantMsg = messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.toolResults).toHaveLength(1)
      expect(assistantMsg?.toolResults?.[0].tool).toBe('search_products')
    })

    it('sets status to complete on done event', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'delta', data: { text: 'Done!' } },
        { event: 'done', data: {} },
      ]))
      const { messages, sendMessage } = useAssistant()
      await sendMessage('Test')
      const assistantMsg = messages.value.find(m => m.role === 'assistant')
      expect(assistantMsg?.status).toBe('complete')
    })

    it('parses error event', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'error', data: { message: 'Rate limited' } },
      ]))
      const { error, sendMessage } = useAssistant()
      await sendMessage('Test')
      expect(error.value).toBe('Rate limited')
    })

    it('refreshes cart on cart_updated in done event', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'done', data: { cart_updated: true } },
      ]))
      mockCartRefresh.mockResolvedValue(undefined)
      const { sendMessage } = useAssistant()
      await sendMessage('Add item')
      expect(mockCartRefresh).toHaveBeenCalledWith(true)
    })

    it('removes empty placeholder on fetch error', async () => {
      globalThis.fetch = vi.fn().mockRejectedValue(new Error('Offline'))
      const { messages, error, sendMessage } = useAssistant()
      await sendMessage('Fail')
      // Placeholder removed since it had no content
      const assistantMessages = messages.value.filter(m => m.role === 'assistant')
      expect(assistantMessages).toHaveLength(0)
      expect(error.value).toBe('Offline')
    })
  })

  describe('clearChat', () => {
    it('empties messages, sessionId, and error', async () => {
      mockPost.mockResolvedValue({
        session_id: 's1',
        message: { id: 'm1', role: 'assistant', content: 'Hi', product_ids: [], created_at: '' },
      })
      const assistant = useAssistant()
      await assistant.sendMessageSync('Hello')
      assistant.clearChat()
      expect(assistant.messages.value).toEqual([])
      expect(assistant.error.value).toBeNull()
    })
  })

  describe('headers', () => {
    it('includes X-Session-ID from localStorage', async () => {
      localStorage.setItem('session-id', 'guest-sid')
      const mockFetchFn = vi.fn().mockResolvedValue(createSSEResponse([
        { event: 'done', data: {} },
      ]))
      globalThis.fetch = mockFetchFn
      const { sendMessage } = useAssistant()
      await sendMessage('Hi')
      const headers = mockFetchFn.mock.calls[0][1].headers
      expect(headers['X-Session-ID']).toBe('guest-sid')
    })
  })
})
