import { useAuthStore } from '~/stores/auth'

let refreshPromise: Promise<boolean> | null = null

export const useApi = () => {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  const apiFetch = $fetch.create({
    baseURL: config.public.apiBase,
    async onRequest({ options }) {
      if (auth.token) {
        options.headers = {
          ...options.headers,
          Authorization: `Bearer ${auth.token}`
        }
      }
    },
    async onResponseError({ request, options, response }) {
      if (response.status === 401) {
        if (!refreshPromise) {
          refreshPromise = auth.tryRefresh().finally(() => {
            refreshPromise = null
          })
        }

        const refreshed = await refreshPromise
        if (!refreshed) {
          auth.logout()
          return
        }

        // Retry the original request with the new token (raw $fetch to avoid loop)
        const retryResponse = await $fetch.raw(request, {
          ...options,
          baseURL: config.public.apiBase,
          headers: {
            ...options.headers,
            Authorization: `Bearer ${auth.token}`
          },
        })

        // Replace the error response body so the caller gets the retried result
        return retryResponse._data
      }
    }
  })

  return {
    fetch: apiFetch
  }
}
