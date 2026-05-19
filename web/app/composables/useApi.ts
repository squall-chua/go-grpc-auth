import { $fetch } from 'ofetch'
import { useAuthStore } from '~/stores/auth'

let refreshPromise: Promise<boolean> | null = null

export const useApi = () => {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  function authHeaders(): Record<string, string> {
    return auth.token ? { Authorization: `Bearer ${auth.token}` } : {}
  }

  async function apiFetch(request: string, options?: any) {
    try {
      return await $fetch(request, {
        ...options,
        baseURL: config.public.apiBase,
        headers: { ...options?.headers, ...authHeaders() },
      })
    } catch (error: any) {
      if (error?.response?.status !== 401) throw error

      if (!refreshPromise) {
        refreshPromise = auth.tryRefresh().finally(() => {
          refreshPromise = null
        })
      }

      const refreshed = await refreshPromise
      if (!refreshed) {
        auth.logout()
        throw error
      }

      // Retry with new token (no interceptor, so a second 401 just throws)
      return await $fetch(request, {
        ...options,
        baseURL: config.public.apiBase,
        headers: { ...options?.headers, ...authHeaders() },
      })
    }
  }

  return { fetch: apiFetch }
}
