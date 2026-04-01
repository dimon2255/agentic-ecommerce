const REQUEST_TIMEOUT = 30000

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = config.public.apiBase

  async function request<T>(path: string, opts: any = {}): Promise<T> {
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT)

    try {
      return await $fetch<T>(`${baseURL}/api/v1${path}`, {
        ...opts,
        signal: controller.signal,
        retry: opts.method && opts.method !== 'GET' ? 0 : 1,
        retryDelay: 1000,
      })
    } finally {
      clearTimeout(timeout)
    }
  }

  async function get<T>(path: string, headers?: Record<string, string>): Promise<T> {
    return request<T>(path, { headers })
  }

  async function post<T>(path: string, body: any, headers?: Record<string, string>): Promise<T> {
    return request<T>(path, { method: 'POST', body, headers })
  }

  async function patch<T>(path: string, body: any, headers?: Record<string, string>): Promise<T> {
    return request<T>(path, { method: 'PATCH', body, headers })
  }

  async function del(path: string, headers?: Record<string, string>): Promise<void> {
    await request(path, { method: 'DELETE', headers })
  }

  return { get, post, patch, del }
}
