import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createMockSupabaseClient } from './__mocks__/supabase'

const mockGet = vi.fn()

mockNuxtImport('useApi', () => {
  return () => ({ get: mockGet, post: vi.fn(), patch: vi.fn(), del: vi.fn() })
})

let mockClient = createMockSupabaseClient({ accessToken: 'admin-token' })

mockNuxtImport('useSupabaseClient', () => () => mockClient)

import { useAdminAuth } from './useAdminAuth'

describe('useAdminAuth', () => {
  beforeEach(() => {
    mockGet.mockReset()
    mockClient = createMockSupabaseClient({ accessToken: 'admin-token' })
  })

  it('fetchPermissions() calls /admin/me with auth header', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: ['catalog:read'], roles: ['admin'] })
    const { fetchPermissions } = useAdminAuth()
    await fetchPermissions()
    expect(mockGet).toHaveBeenCalledWith('/admin/me', { Authorization: 'Bearer admin-token' })
  })

  it('fetchPermissions() stores permissions and roles', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: ['catalog:read', 'orders:read'], roles: ['admin'] })
    const { permissions, roles, fetchPermissions } = useAdminAuth()
    await fetchPermissions()
    expect(permissions.value).toEqual(['catalog:read', 'orders:read'])
    expect(roles.value).toEqual(['admin'])
  })

  it('loading is true during fetch, false after', async () => {
    let resolveGet: (v: any) => void
    mockGet.mockReturnValue(new Promise((r) => { resolveGet = r }))

    const { loading, fetchPermissions } = useAdminAuth()
    const promise = fetchPermissions()
    expect(loading.value).toBe(true)

    resolveGet!({ user_id: 'u1', permissions: [], roles: [] })
    await promise
    expect(loading.value).toBe(false)
  })

  it('hasPermission() returns true when ALL keys present (AND)', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: ['catalog:read', 'orders:read', 'reports:read'], roles: [] })
    const { fetchPermissions, hasPermission } = useAdminAuth()
    await fetchPermissions()
    expect(hasPermission('catalog:read', 'orders:read')).toBe(true)
  })

  it('hasPermission() returns false when any key missing', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: ['catalog:read'], roles: [] })
    const { fetchPermissions, hasPermission } = useAdminAuth()
    await fetchPermissions()
    expect(hasPermission('catalog:read', 'orders:write')).toBe(false)
  })

  it('hasAnyPermission() returns true when ANY key present (OR)', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: ['catalog:read'], roles: [] })
    const { fetchPermissions, hasAnyPermission } = useAdminAuth()
    await fetchPermissions()
    expect(hasAnyPermission('catalog:read', 'orders:write')).toBe(true)
  })

  it('hasAnyPermission() returns false when NO keys present', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: [], roles: [] })
    const { fetchPermissions, hasAnyPermission } = useAdminAuth()
    await fetchPermissions()
    expect(hasAnyPermission('catalog:read', 'orders:write')).toBe(false)
  })

  it('sets empty permissions when no session', async () => {
    mockClient = createMockSupabaseClient({ accessToken: null })
    const { permissions, roles, fetchPermissions } = useAdminAuth()
    await fetchPermissions()
    expect(permissions.value).toEqual([])
    expect(roles.value).toEqual([])
    expect(mockGet).not.toHaveBeenCalled()
  })

  it('sets empty permissions on API error', async () => {
    mockGet.mockRejectedValue(new Error('Server error'))
    const { permissions, roles, fetchPermissions } = useAdminAuth()
    await fetchPermissions()
    expect(permissions.value).toEqual([])
    expect(roles.value).toEqual([])
  })

  it('reset() clears permissions and roles to null', async () => {
    mockGet.mockResolvedValue({ user_id: 'u1', permissions: ['x'], roles: ['y'] })
    const { permissions, roles, fetchPermissions, reset } = useAdminAuth()
    await fetchPermissions()
    reset()
    expect(permissions.value).toBeNull()
    expect(roles.value).toBeNull()
  })
})
