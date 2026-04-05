import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createMockSupabaseClient } from './__mocks__/supabase'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPatch = vi.fn()
const mockDel = vi.fn()

mockNuxtImport('useApi', () => {
  return () => ({ get: mockGet, post: mockPost, patch: mockPatch, del: mockDel })
})

let mockClient = createMockSupabaseClient({ accessToken: 'test-token' })

mockNuxtImport('useSupabaseClient', () => () => mockClient)

import { useAdminApi } from './useAdminApi'

describe('useAdminApi', () => {
  beforeEach(() => {
    mockGet.mockReset()
    mockPost.mockReset()
    mockPatch.mockReset()
    mockDel.mockReset()
    mockClient = createMockSupabaseClient({ accessToken: 'test-token' })
  })

  it('get() prefixes path with /admin', async () => {
    const { get } = useAdminApi()
    await get('/products')
    expect(mockGet).toHaveBeenCalledWith('/admin/products', { Authorization: 'Bearer test-token' })
  })

  it('post() prefixes path with /admin and passes body', async () => {
    const { post } = useAdminApi()
    const body = { name: 'New' }
    await post('/products', body)
    expect(mockPost).toHaveBeenCalledWith('/admin/products', body, { Authorization: 'Bearer test-token' })
  })

  it('patch() prefixes path with /admin and passes body', async () => {
    const { patch } = useAdminApi()
    const body = { name: 'Updated' }
    await patch('/products/1', body)
    expect(mockPatch).toHaveBeenCalledWith('/admin/products/1', body, { Authorization: 'Bearer test-token' })
  })

  it('del() prefixes path with /admin', async () => {
    const { del } = useAdminApi()
    await del('/products/1')
    expect(mockDel).toHaveBeenCalledWith('/admin/products/1', { Authorization: 'Bearer test-token' })
  })

  it('includes Bearer token in auth header when session exists', async () => {
    const { get } = useAdminApi()
    await get('/me')
    expect(mockGet).toHaveBeenCalledWith('/admin/me', { Authorization: 'Bearer test-token' })
  })

  it('sends empty headers when no session', async () => {
    mockClient = createMockSupabaseClient({ accessToken: null })
    const { get } = useAdminApi()
    await get('/me')
    expect(mockGet).toHaveBeenCalledWith('/admin/me', {})
  })

  it('awaits auth headers before each call', async () => {
    const { post } = useAdminApi()
    await post('/items', { x: 1 })
    expect(mockClient.auth.getSession).toHaveBeenCalled()
  })
})
