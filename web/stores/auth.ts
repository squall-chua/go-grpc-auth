import { defineStore } from 'pinia'
import { useStorage } from '@vueuse/core'

interface User {
  id: string
  email: string
  username: string
  roles: string[]
  permissions: string[]
}

export const useAuthStore = defineStore('auth', () => {
  const token = useCookie('auth_token', { maxAge: 60 * 60 * 24 * 7 })
  const refreshToken = useCookie('refresh_token', { maxAge: 60 * 60 * 24 * 30 })
  const user = useStorage<User | null>('auth_user', null)
  const mfaToken = ref<string | null>(null)
  const mfaRequired = ref(false)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function setTokens(pair: { accessToken: string; refreshToken: string; expiresIn: number }) {
    token.value = pair.accessToken
    refreshToken.value = pair.refreshToken
  }

  function setUser(userData: User) {
    user.value = userData
  }

  function setMFARequired(token: string) {
    mfaToken.value = token
    mfaRequired.value = true
  }

  function logout() {
    token.value = null
    refreshToken.value = null
    user.value = null
    mfaToken.value = null
    mfaRequired.value = false
  }

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
    logout
  }
})
