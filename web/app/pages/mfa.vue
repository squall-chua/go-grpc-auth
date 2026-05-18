<template>
  <div class="flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <UCard class="w-full max-w-md space-y-8 p-4">
      <div class="text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 mb-4">
          <UIcon name="i-heroicons-shield-check" class="w-8 h-8" />
        </div>
        <h2 class="text-3xl font-extrabold tracking-tight">Two-Step Verification</h2>
        <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
          Enter the code from your authenticator app to continue
        </p>
      </div>

      <form class="mt-8 space-y-6" @submit.prevent="handleVerify">
        <UFormGroup label="Verification Code" name="code" help="Enter the 6-digit code">
          <UInput
            v-model="code"
            placeholder="000000"
            size="xl"
            class="text-center tracking-[1em] font-mono"
            maxlength="6"
            autocomplete="one-time-code"
            autofocus
          />
        </UFormGroup>

        <UButton type="submit" block :loading="loading" size="lg">
          Verify Code
        </UButton>

        <UButton variant="ghost" block @click="auth.logout(); navigateTo('/login')">
          Cancel and return to login
        </UButton>
      </form>
    </UCard>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const auth = useAuthStore()
const api = useApi()
const toast = useToast()

const code = ref('')
const loading = ref(false)

onMounted(() => {
  if (!auth.mfaToken) {
    navigateTo('/login')
  }
})

async function handleVerify() {
  if (code.value.length !== 6) {
    toast.add({ title: 'Invalid code', description: 'Please enter a 6-digit code', color: 'orange' })
    return
  }

  loading.value = true
  try {
    const res = await api.fetch('/v1/auth/mfa/verify', {
      method: 'POST',
      body: {
        mfa_token: auth.mfaToken,
        code: code.value
      }
    })

    auth.setTokens(res)
    const principal = await api.fetch('/v1/auth/validate', {
      method: 'POST',
      body: { token: res.access_token }
    })
    auth.setUser({
      id: principal.user_id,
      email: '',
      username: '',
      namespace: principal.namespace,
      roles: principal.roles || [],
      permissions: principal.permissions || [],
    })

    toast.add({ title: 'Verified successfully', color: 'green' })
    navigateTo('/dashboard')
  } catch (err) {
    toast.add({ title: 'Verification failed', description: err.data?.message || 'Invalid code', color: 'red' })
  } finally {
    loading.value = false
  }
}
</script>
