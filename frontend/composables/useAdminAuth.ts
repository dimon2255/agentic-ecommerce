export function useAdminAuth() {
  const permissions = useState<string[] | null>('admin-permissions', () => null)
  const roles = useState<string[] | null>('admin-roles', () => null)
  const loading = useState('admin-auth-loading', () => false)

  async function fetchPermissions() {
    loading.value = true
    try {
      const client = useSupabaseClient()
      const { data: { session } } = await client.auth.getSession()
      if (!session?.access_token) {
        permissions.value = []
        roles.value = []
        return
      }
      const { get } = useApi()
      const resp = await get<{ user_id: string; permissions: string[]; roles: string[] }>(
        '/admin/me',
        { Authorization: `Bearer ${session.access_token}` },
      )
      permissions.value = resp.permissions
      roles.value = resp.roles
    } catch {
      permissions.value = []
      roles.value = []
    } finally {
      loading.value = false
    }
  }

  function hasPermission(...keys: string[]) {
    return keys.every(k => permissions.value?.includes(k))
  }

  function hasAnyPermission(...keys: string[]) {
    return keys.some(k => permissions.value?.includes(k))
  }

  function reset() {
    permissions.value = null
    roles.value = null
  }

  return { permissions, roles, loading, fetchPermissions, hasPermission, hasAnyPermission, reset }
}
