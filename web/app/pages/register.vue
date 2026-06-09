<template>
  <div class="flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <UCard class="w-full max-w-md space-y-8 p-4">
      <div class="text-center">
        <p class="text-xs font-heading font-medium uppercase tracking-wider text-code-text mb-2">Create account</p>
        <h1 class="text-3xl font-heading font-semibold tracking-tight">Join {{ config.public.name }}</h1>
        <p class="mt-2 text-sm text-text-muted">
          Secure your digital identity
        </p>
      </div>

      <form class="mt-8 space-y-6" @submit.prevent="handleRegister">
        <div class="space-y-4">
          <UFormGroup label="Username" name="username">
            <UInput v-model="form.username" placeholder="johndoe" icon="i-heroicons-user" />
          </UFormGroup>
          <UFormGroup label="Email address" name="email">
            <UInput v-model="form.email" type="email" placeholder="you@example.com" icon="i-heroicons-envelope" />
          </UFormGroup>
          <UFormGroup label="Password" name="password">
            <UInput v-model="form.password" type="password" placeholder="••••••••" icon="i-heroicons-lock-closed" />
          </UFormGroup>
        </div>

        <UButton type="submit" block :loading="loading" class="h-11">
          Create Account
        </UButton>
      </form>

      <p class="mt-8 text-center text-sm text-slate-500">
        Already have an account?
        <ULink to="/login" class="font-medium text-primary-600 hover:text-primary-500">
          Sign in
        </ULink>
      </p>
    </UCard>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const config = useRuntimeConfig()
const auth = useAuthStore()
const api = useApi()
const toast = useToast()

const form = reactive({
  username: '',
  email: '',
  password: '',
  namespace: 'default'
})

const loading = ref(false)

async function handleRegister() {
  loading.value = true
  try {
    const res = await api.fetch('/v1/auth/register', {
      method: 'POST',
      body: form
    })

    if (res.mfa_required) {
      auth.setMFARequired(res.mfa_token, res.mfa_methods || [])
      navigateTo('/mfa')
      return
    }

    auth.setTokens(res)
    const principal = await api.fetch('/v1/auth/validate', {
      method: 'POST',
      body: { token: res.access_token }
    })
    auth.setUser({
      id: principal.user_id,
      email: form.email,
      username: form.username,
      namespace: principal.namespace,
      roles: principal.roles || [],
      permissions: principal.permissions || [],
    })

    toast.add({ title: 'Account created!', color: 'green' })
    navigateTo('/dashboard')
  } catch (err) {
    toast.add({ title: 'Registration failed', description: err.data?.message || 'Check your details', color: 'red' })
  } finally {
    loading.value = false
  }
}
</script>
