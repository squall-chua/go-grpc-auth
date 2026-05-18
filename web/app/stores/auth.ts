import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { $fetch } from 'ofetch'

interface User {
  id: string
  email: string
  username: string
  namespace: string
  roles: string[]
  permissions: string[]
}

export const useAuthStore = defineStore('auth', () => {
  const token = useCookie('auth_token', { maxAge: 60 * 60 * 24 * 7 })
  const refreshToken = useCookie('refresh_token', { maxAge: 60 * 60 * 24 * 30 })
  const user = ref<User | null>(null)
  const mfaToken = ref<string | null>(null)
  const mfaRequired = ref(false)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function setTokens(pair: { access_token: string; refresh_token: string; expires_in: number }) {
    token.value = pair.access_token
    refreshToken.value = pair.refresh_token
  }

  function setUser(userData: User) {
    user.value = userData
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_user', JSON.stringify(userData))
    }
  }

  function loadUser() {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('auth_user')
      if (stored) {
        try { user.value = JSON.parse(stored) } catch { user.value = null }
      }
    }
  }

  function setMFARequired(mfaTokenStr: string) {
    mfaToken.value = mfaTokenStr
    mfaRequired.value = true
  }

  async function callLogout() {
    try {
      const config = useRuntimeConfig()
      await $fetch('/v1/auth/logout', {
        baseURL: config.public.apiBase as string,
        method: 'POST',
        headers: token.value ? { Authorization: `Bearer ${token.value}` } : {},
      })
    } catch {
      // Ignore errors during logout
    }
  }

  async function logout() {
    await callLogout()
    token.value = null
    refreshToken.value = null
    user.value = null
    mfaToken.value = null
    mfaRequired.value = false
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_user')
    }
    navigateTo('/login')
  }

  async function tryRefresh(): Promise<boolean> {
    if (!refreshToken.value) return false
    try {
      const config = useRuntimeConfig()
      const res = await $fetch<{ access_token: string; refresh_token: string; expires_in: number }>('/v1/auth/refresh', {
        baseURL: config.public.apiBase as string,
        method: 'POST',
        body: { refresh_token: refreshToken.value },
      })
      setTokens(res)
      return true
    } catch {
      return false
    }
  }

  // Load user from localStorage on init
  loadUser()

  return {
    token,
    refreshToken,
    user,
    mfaToken,
    mfaRequired,
    isAuthenticated,
    setTokens,
    setUser,
    setMFARequired,
    logout,
    tryRefresh,
  }
})
