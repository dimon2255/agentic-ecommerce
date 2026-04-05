export function useAdminApi() {
  const client = useSupabaseClient()
  const { get: rawGet, post: rawPost, patch: rawPatch, del: rawDel } = useApi()

  async function authHeaders(): Promise<Record<string, string>> {
    const { data: { session } } = await client.auth.getSession()
    if (session?.access_token) {
      return { Authorization: `Bearer ${session.access_token}` }
    }
    return {}
  }

  async function get<T>(path: string): Promise<T> {
    return rawGet<T>(`/admin${path}`, await authHeaders())
  }

  async function post<T>(path: string, body: any): Promise<T> {
    return rawPost<T>(`/admin${path}`, body, await authHeaders())
  }

  async function patch<T>(path: string, body: any): Promise<T> {
    return rawPatch<T>(`/admin${path}`, body, await authHeaders())
  }

  async function del(path: string): Promise<void> {
    return rawDel(`/admin${path}`, await authHeaders())
  }

  return { get, post, patch, del }
}
