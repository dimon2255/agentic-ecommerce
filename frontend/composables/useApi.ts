export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function get<T>(path: string): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`)
  }

  async function post<T>(path: string, body: any): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'POST',
      body,
    })
  }

  async function patch<T>(path: string, body: any): Promise<T> {
    return await $fetch<T>(`${baseURL}/api/v1${path}`, {
      method: 'PATCH',
      body,
    })
  }

  async function del(path: string): Promise<void> {
    await $fetch(`${baseURL}/api/v1${path}`, {
      method: 'DELETE',
    })
  }

  return { get, post, patch, del }
}
