<template>
  <div class="flex flex-col items-center justify-center min-h-[60vh] space-y-4">
    <ULoadingIcon size="xl" />
    <p class="text-slate-500 animate-pulse">Authenticating with {{ provider }}...</p>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const route = useRoute()
const auth = useAuthStore()
const api = useApi()
const toast = useToast()

const provider = route.params.provider
const code = route.query.code
const state = route.query.state

onMounted(async () => {
  if (!code) {
    toast.add({ title: 'Error', description: 'No code provided from social login', color: 'red' })
    navigateTo('/login')
    return
  }

  try {
    const res = await api.fetch('/v1/auth/social/callback', {
      method: 'POST',
      body: {
        provider,
        code,
        state,
        namespace: 'default'
      }
    })

    if (res.mfa_required) {
      auth.setMFARequired(res.mfa_token)
      navigateTo('/mfa')
      return
    }

    auth.setTokens(res)
    // Fetch profile
    const principal = await api.fetch('/v1/auth/validate', {
      method: 'POST',
      body: { token: res.access_token }
    })
    auth.setUser(principal)
    
    toast.add({ title: 'Successfully logged in!', color: 'green' })
    navigateTo('/dashboard')
  } catch (err) {
    toast.add({ title: 'Authentication failed', description: err.data?.message || 'Social login failed', color: 'red' })
    navigateTo('/login')
  }
})
</script>
