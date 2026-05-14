import { useAuthStore } from '~/stores/auth'

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
    async onResponseError({ response }) {
      if (response.status === 401) {
        auth.logout()
        navigateTo('/login')
      }
    }
  })

  return {
    fetch: apiFetch
  }
}
