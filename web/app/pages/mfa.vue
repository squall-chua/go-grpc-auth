<template>
  <div class="flex flex-col items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
    <UCard class="w-full max-w-md space-y-6 p-4">
      <div class="text-center">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400 mb-4">
          <UIcon name="i-heroicons-shield-check" class="w-8 h-8" />
        </div>
        <h2 class="text-3xl font-extrabold tracking-tight">Two-Step Verification</h2>
        <p class="mt-2 text-sm text-slate-500 dark:text-slate-400">
          {{ subtitle }}
        </p>
      </div>

      <!-- Method Selection -->
      <div v-if="!selectedMethod" class="space-y-3">
        <div
          v-for="m in auth.mfaMethods"
          :key="m"
          class="flex items-center gap-3 p-4 border-2 border-slate-200 dark:border-slate-700 rounded-lg cursor-pointer hover:border-primary-400 dark:hover:border-primary-500 transition-colors duration-200"
          @click="selectMethod(m)"
        >
          <div class="w-10 h-10 rounded-lg flex items-center justify-center" :class="methodBg(m)">
            <UIcon :name="methodIcon(m)" class="w-5 h-5" />
          </div>
          <div>
            <p class="font-semibold text-sm">{{ methodLabel(m) }}</p>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ methodDesc(m) }}</p>
          </div>
        </div>
      </div>

      <!-- Code Entry -->
      <form v-else class="space-y-6" @submit.prevent="handleVerify">
        <div v-if="maskedRecipient" class="text-center text-sm text-slate-500 dark:text-slate-400">
          <p>Code expires in 5 minutes</p>
        </div>

        <div class="space-y-2">
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300">Verification Code</label>
          <OtpInput v-model="code" />
          <p class="text-xs text-slate-500">Enter the 6-digit code</p>
        </div>

        <UButton type="submit" block :loading="loading" size="lg">
          Verify Code
        </UButton>

        <div class="flex items-center justify-center gap-4 text-sm">
          <button
            v-if="selectedMethod !== 'totp'"
            type="button"
            class="text-primary-600 hover:text-primary-500 cursor-pointer transition-colors duration-200"
            @click="resendCode"
          >
            Resend code
          </button>
          <span v-if="selectedMethod !== 'totp' && auth.mfaMethods.length > 1" class="text-slate-300 dark:text-slate-600">|</span>
          <button
            v-if="auth.mfaMethods.length > 1"
            type="button"
            class="text-slate-500 hover:text-slate-400 cursor-pointer transition-colors duration-200"
            @click="switchMethod"
          >
            Try another method
          </button>
        </div>
      </form>

      <UButton variant="ghost" block @click="auth.logout(); navigateTo('/login')">
        Cancel and return to login
      </UButton>
    </UCard>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

const auth = useAuthStore()
const api = useApi()
const toast = useToast()

const selectedMethod = ref(null)
const maskedRecipient = ref('')
const code = ref('')
const loading = ref(false)
const initiating = ref(false)

const methodMeta = {
  totp: { icon: 'i-heroicons-device-phone-mobile', label: 'Authenticator App', desc: 'Enter code from your TOTP app', bg: 'bg-green-50 dark:bg-green-900/30 text-green-600 dark:text-green-400' },
  email: { icon: 'i-heroicons-envelope', label: 'Email', desc: 'Send a code to your email', bg: 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400' },
  sms: { icon: 'i-heroicons-chat-bubble-left', label: 'SMS', desc: 'Send a code to your phone', bg: 'bg-amber-50 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400' },
}

function methodIcon(m) { return methodMeta[m]?.icon || 'i-heroicons-shield-check' }
function methodLabel(m) { return methodMeta[m]?.label || m }
function methodDesc(m) { return methodMeta[m]?.desc || '' }
function methodBg(m) { return methodMeta[m]?.bg || 'bg-slate-100 dark:bg-slate-800' }

const subtitle = computed(() => {
  if (!selectedMethod.value) return 'Choose how you\'d like to verify your identity'
  if (selectedMethod.value === 'totp') return 'Enter the code from your authenticator app'
  if (maskedRecipient.value) return `A code was sent to ${maskedRecipient.value}`
  return 'Enter the verification code'
})

onMounted(async () => {
  if (!auth.mfaToken) {
    navigateTo('/login')
    return
  }
  if (auth.mfaMethods.length === 1) {
    await selectMethod(auth.mfaMethods[0])
  }
})

async function selectMethod(method) {
  selectedMethod.value = method
  code.value = ''
  maskedRecipient.value = ''
  if (method === 'totp') return

  await initiateCode(method)
}

async function initiateCode(method) {
  initiating.value = true
  try {
    const res = await api.fetch('/v1/auth/mfa/initiate', {
      method: 'POST',
      body: { mfa_token: auth.mfaToken, method }
    })
    maskedRecipient.value = res.masked_recipient || ''
  } catch (err) {
    toast.add({ title: 'Failed to send code', description: err.data?.message || 'Try again', color: 'red' })
  } finally {
    initiating.value = false
  }
}

async function resendCode() {
  if (!selectedMethod.value) return
  await initiateCode(selectedMethod.value)
  toast.add({ title: 'Code resent', color: 'green' })
}

function switchMethod() {
  selectedMethod.value = null
  maskedRecipient.value = ''
  code.value = ''
}

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
