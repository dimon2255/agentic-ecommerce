import { mockNuxtImport } from '@nuxt/test-utils/runtime'

const mockFetch = vi.fn()

mockNuxtImport('useRuntimeConfig', () => {
  return () => ({
    app: { baseURL: '/' },
    public: { apiBase: 'http://test-api' },
  })
})

import { useApi } from './useApi'

describe('useApi', () => {
  let originalFetch: typeof $fetch

  beforeEach(() => {
    originalFetch = globalThis.$fetch
    globalThis.$fetch = mockFetch as any
    mockFetch.mockReset()
    mockFetch.mockResolvedValue({ ok: true })
  })

  afterEach(() => {
    globalThis.$fetch = originalFetch
  })

  it('get() calls $fetch with correct URL', async () => {
    const { get } = useApi()
    await get('/products')
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test-api/api/v1/products',
      expect.objectContaining({ headers: undefined }),
    )
  })

  it('get() passes headers through', async () => {
    const { get } = useApi()
    await get('/products', { 'X-Custom': 'value' })
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test-api/api/v1/products',
      expect.objectContaining({ headers: { 'X-Custom': 'value' } }),
    )
  })

  it('get() uses retry=1 (default, no method specified)', async () => {
    const { get } = useApi()
    await get('/items')
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ retry: 1, retryDelay: 1000 }),
    )
  })

  it('post() sends method POST with body and retry=0', async () => {
    const { post } = useApi()
    const body = { name: 'Test' }
    await post('/products', body)
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test-api/api/v1/products',
      expect.objectContaining({ method: 'POST', body, retry: 0 }),
    )
  })

  it('patch() sends method PATCH with body and retry=0', async () => {
    const { patch } = useApi()
    const body = { name: 'Updated' }
    await patch('/products/1', body)
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test-api/api/v1/products/1',
      expect.objectContaining({ method: 'PATCH', body, retry: 0 }),
    )
  })

  it('del() sends method DELETE with retry=0', async () => {
    const { del } = useApi()
    await del('/products/1')
    expect(mockFetch).toHaveBeenCalledWith(
      'http://test-api/api/v1/products/1',
      expect.objectContaining({ method: 'DELETE', retry: 0 }),
    )
  })

  it('includes AbortController signal in request', async () => {
    const { get } = useApi()
    await get('/test')
    const opts = mockFetch.mock.calls[0][1]
    expect(opts.signal).toBeInstanceOf(AbortSignal)
  })

  it('propagates $fetch errors to caller', async () => {
    mockFetch.mockRejectedValue(new Error('Network error'))
    const { get } = useApi()
    await expect(get('/fail')).rejects.toThrow('Network error')
  })

  it('post() passes headers through', async () => {
    const { post } = useApi()
    await post('/items', { x: 1 }, { Authorization: 'Bearer token' })
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ headers: { Authorization: 'Bearer token' } }),
    )
  })

  it('del() passes headers through', async () => {
    const { del } = useApi()
    await del('/items/1', { Authorization: 'Bearer token' })
    expect(mockFetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ headers: { Authorization: 'Bearer token' } }),
    )
  })
})
