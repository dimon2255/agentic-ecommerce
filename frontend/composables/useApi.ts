export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function get<T>(path: string, headers?: Record<string, string>): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, { headers })
  }

  async function post<T>(path: string, body: any, headers?: Record<string, string>): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'POST',
      body,
      headers,
    })
  }

  async function patch<T>(path: string, body: any, headers?: Record<string, string>): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'PATCH',
      body,
      headers,
    })
  }

  async function del(path: string, headers?: Record<string, string>): Promise<void> {
    await $fetch(`${baseURL}/api/v1${path}`, {
      method: 'DELETE',
      headers,
    })
  }

  return { get, post, patch, del }
}
