import { vi } from 'vitest'

export function createMockSupabaseClient(overrides?: {
  accessToken?: string | null
}) {
  return {
    auth: {
      getSession: vi.fn().mockResolvedValue({
        data: {
          session: overrides?.accessToken
            ? { access_token: overrides.accessToken }
            : null,
        },
      }),
      signOut: vi.fn().mockResolvedValue({}),
    },
  }
}

export function createMockSupabaseSession(accessToken?: string | null) {
  return ref(accessToken ? { access_token: accessToken } : null)
}
