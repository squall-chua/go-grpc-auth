<template>
  <div class="flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <UCard class="w-full max-w-md space-y-8 p-4">
      <div class="text-center">
        <p class="text-xs font-heading font-medium uppercase tracking-wider text-code-text mb-2">Sign in</p>
        <h1 class="text-3xl font-heading font-semibold tracking-tight">Welcome back</h1>
        <p class="mt-2 text-sm text-text-muted">
          Enter your credentials to access your account
        </p>
      </div>

      <form class="mt-8 space-y-6" @submit.prevent="handleLogin">
        <div class="space-y-4">
          <UFormGroup label="Email or Username" name="login">
            <UInput v-model="form.login" placeholder="you@example.com" icon="i-heroicons-envelope" />
          </UFormGroup>
          <UFormGroup label="Password" name="password">
            <UInput v-model="form.password" type="password" placeholder="••••••••" icon="i-heroicons-lock-closed" />
          </UFormGroup>
        </div>

        <div class="flex items-center justify-between text-sm">
          <UCheckbox label="Remember me" v-model="rememberMe" />
          <ULink to="/forgot-password" class="font-medium text-primary-600 hover:text-primary-500">
            Forgot password?
          </ULink>
        </div>

        <UButton type="submit" block :loading="loading" class="h-11">
          Sign in
        </UButton>

        <div class="relative py-4">
          <div class="absolute inset-0 flex items-center">
            <div class="w-full border-t border-slate-200 dark:border-slate-800"></div>
          </div>
          <div class="relative flex justify-center text-xs uppercase">
            <span class="bg-white dark:bg-slate-900 px-2 text-slate-500">Or continue with</span>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <UButton color="white" block @click="handleSocialLogin('google')">
            <template #leading>
              <UIcon name="i-logos-google-icon" class="w-4 h-4" />
            </template>
            Google
          </UButton>
          <UButton color="white" block @click="handleSocialLogin('github')">
            <template #leading>
              <UIcon name="i-mdi-github" class="w-4 h-4" />
            </template>
            GitHub
          </UButton>
        </div>

        <UButton color="white" block @click="handleWalletLogin" :loading="walletLoading">
          <template #leading>
            <UIcon name="i-mdi-ethereum" class="w-4 h-4" />
          </template>
          Sign in with Wallet
        </UButton>
      </form>

      <p class="mt-8 text-center text-sm text-slate-500">
        Don't have an account?
        <ULink to="/register" class="font-medium text-primary-600 hover:text-primary-500">
          Register now
        </ULink>
      </p>
    </UCard>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'
import { useWeb3Auth } from '~/composables/useWeb3Auth'

const auth = useAuthStore()
const api = useApi()
const toast = useToast()
const web3 = useWeb3Auth()

const form = reactive({
  login: '',
  password: '',
  namespace: 'default'
})

const loading = ref(false)
const rememberMe = ref(false)
const walletLoading = ref(false)

async function handleLogin() {
  loading.value = true
  try {
    const res = await api.fetch('/v1/auth/login', {
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
      email: form.login.includes('@') ? form.login : '',
      username: form.login.includes('@') ? '' : form.login,
      namespace: principal.namespace,
      roles: principal.roles || [],
      permissions: principal.permissions || [],
    })

    toast.add({ title: 'Welcome back!', color: 'green' })
    navigateTo('/dashboard')
  } catch (err) {
    toast.add({ title: 'Error', description: err.data?.message || 'Login failed', color: 'red' })
  } finally {
    loading.value = false
  }
}

async function handleSocialLogin(provider) {
  try {
    const res = await api.fetch(`/v1/auth/social/${provider}/url`, {
      method: 'POST',
      body: { state: Math.random().toString(36).substring(7) }
    })
    window.location.href = res.url
  } catch (err) {
    toast.add({ title: 'Social login error', color: 'red' })
  }
}

async function handleWalletLogin() {
  walletLoading.value = true
  try {
    if (!web3.isConnected.value) {
      const c = web3.connectors.value[0]
      if (!c) throw new Error('No wallet connector available')
      await web3.connect({ connector: c })
    }
    const tokens = await web3.signIn()

    auth.setTokens(tokens)
    const principal = await api.fetch('/v1/auth/validate', {
      method: 'POST',
      body: { token: tokens.access_token }
    })
    auth.setUser({
      id: principal.user_id,
      email: web3.address.value ?? '',
      username: '',
      namespace: principal.namespace,
      roles: principal.roles || [],
      permissions: principal.permissions || [],
    })

    toast.add({ title: 'Welcome!', color: 'green' })
    navigateTo('/dashboard')
  } catch (err) {
    toast.add({ title: 'Wallet login error', description: err?.message ?? 'failed', color: 'red' })
  } finally {
    walletLoading.value = false
  }
}
</script>
