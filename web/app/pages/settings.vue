<template>
  <div class="max-w-2xl mx-auto space-y-8">
    <PageHeader
      eyebrow="Settings"
      title="Preferences"
      subtitle="Application preferences and security"
    />

    <UCard>
      <template #header>
        <h3 class="font-semibold">Appearance</h3>
      </template>
      <div class="flex items-center justify-between">
        <div>
          <p class="font-medium">Theme</p>
          <p class="text-sm text-slate-500">Choose your preferred color scheme</p>
        </div>
        <USelectMenu v-model="colorMode.preference" :options="['system', 'light', 'dark']" class="w-32" />
      </div>
    </UCard>

    <UCard>
      <template #header>
        <div>
          <h3 class="font-semibold">Security</h3>
          <p class="text-sm text-slate-500">Manage two-factor authentication methods</p>
        </div>
      </template>

      <div v-if="mfaMethods.length > 0 && !mfaMethods.some(m => m.enrolled)" class="mb-4 p-3 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg flex items-center gap-3">
        <UIcon name="i-heroicons-exclamation-triangle" class="w-5 h-5 text-amber-600 shrink-0" />
        <div>
          <p class="font-medium text-amber-800 dark:text-amber-200 text-sm">No MFA methods configured</p>
          <p class="text-xs text-amber-600 dark:text-amber-400">Add a method below to secure your account</p>
        </div>
      </div>

      <div class="divide-y divide-slate-200 dark:divide-slate-800">
        <div v-for="m in mfaMethods" :key="m.method" class="flex items-center justify-between py-4 first:pt-0 last:pb-0">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg flex items-center justify-center" :class="mfaMethodBg(m.method)">
              <UIcon :name="mfaMethodIcon(m.method)" class="w-5 h-5" />
            </div>
            <div>
              <p class="font-medium text-sm">{{ mfaMethodLabel(m.method) }}</p>
              <div v-if="m.enrolled" class="flex items-center gap-1.5 mt-0.5">
                <span class="w-2 h-2 rounded-full bg-green-500"></span>
                <span class="text-xs text-green-600 dark:text-green-400">{{ m.method === 'totp' ? 'Configured' : 'Enabled' }}</span>
              </div>
              <p v-else-if="!m.available" class="text-xs text-slate-400">Requires a phone number on your profile</p>
            </div>
          </div>
          <UButton v-if="m.enrolled" color="red" variant="soft" size="xs" @click="handleRemoveMethod(m.method)">
            {{ m.method === 'totp' ? 'Remove' : 'Disable' }}
          </UButton>
          <UButton v-else-if="m.method === 'totp'" size="xs" @click="openTOTPSetup">Set up</UButton>
          <UButton v-else-if="m.available" size="xs" @click="handleEnableMethod(m.method)">Enable</UButton>
          <UButton v-else size="xs" disabled>Not available</UButton>
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="font-semibold">Session</h3>
      </template>
      <div class="flex items-center justify-between">
        <div>
          <p class="font-medium">Sign out</p>
          <p class="text-sm text-slate-500">End your current session and revoke tokens</p>
        </div>
        <UButton color="red" variant="soft" @click="auth.logout()">Sign Out</UButton>
      </div>
    </UCard>
  </div>

  <UModal v-model="totpModalOpen">
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold">Set Up Authenticator App</h3>
          <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="totpModalOpen = false" />
        </div>
      </template>

      <div class="text-center space-y-4">
        <p class="text-sm text-slate-500">Scan this QR code with your authenticator app</p>

        <div v-if="totpQRSvg" class="inline-block p-4 bg-white rounded-xl border border-slate-200">
          <div class="w-44 h-44" v-html="totpQRSvg" />
        </div>

        <details class="text-left text-sm">
          <summary class="text-primary-600 cursor-pointer hover:text-primary-500 transition-colors duration-200">Can't scan? Enter key manually</summary>
          <div class="mt-2 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg font-mono text-xs break-all tracking-wider">
            {{ totpSecret }}
          </div>
        </details>

        <div class="space-y-2">
          <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 text-left">Enter verification code to confirm</label>
          <OtpInput v-model="totpVerifyCode" />
        </div>

        <div class="space-y-2">
          <UButton block :loading="totpVerifying" :disabled="totpVerifyCode.length !== 6" :class="{ 'opacity-50 cursor-not-allowed': totpVerifyCode.length !== 6 }" @click="handleVerifyTOTPSetup">Verify & Enable</UButton>
          <UButton variant="ghost" block @click="totpModalOpen = false">Cancel</UButton>
        </div>
      </div>
    </UCard>
  </UModal>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { renderSVG } from 'uqr'

definePageMeta({ middleware: 'auth' })

const auth = useAuthStore()
const colorMode = useColorMode()
const api = useApi()
const toast = useToast()

const mfaMethods = ref([])
const totpModalOpen = ref(false)
const totpSecret = ref('')
const totpQRSvg = ref('')
const totpVerifyCode = ref('')
const totpVerifying = ref(false)

const mfaMethodMeta = {
  totp: { icon: 'i-heroicons-device-phone-mobile', label: 'Authenticator App', bg: 'bg-green-50 dark:bg-green-900/30 text-green-600 dark:text-green-400' },
  email: { icon: 'i-heroicons-envelope', label: 'Email', bg: 'bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400' },
  sms: { icon: 'i-heroicons-chat-bubble-left', label: 'SMS', bg: 'bg-amber-50 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400' },
}

function mfaMethodIcon(method) { return mfaMethodMeta[method]?.icon || 'i-heroicons-shield-check' }
function mfaMethodLabel(method) { return mfaMethodMeta[method]?.label || method }
function mfaMethodBg(method) { return mfaMethodMeta[method]?.bg || 'bg-slate-100 dark:bg-slate-800' }

async function fetchMFAMethods() {
  try {
    const res = await api.fetch('/v1/auth/mfa/methods')
    mfaMethods.value = res.methods || []
  } catch { /* ignore */ }
}

async function openTOTPSetup() {
  try {
    const res = await api.fetch('/v1/auth/mfa/initiate', {
      method: 'POST',
      body: { method: 'totp', mfa_token: '' }
    })
    totpSecret.value = res.secret
    totpQRSvg.value = renderSVG(res.qr_code_url)
    totpVerifyCode.value = ''
    totpModalOpen.value = true
  } catch (err) {
    toast.add({ title: 'Failed to start TOTP setup', description: err.data?.message || 'Try again', color: 'red' })
  }
}

async function handleVerifyTOTPSetup() {
  if (totpVerifyCode.value.length !== 6) {
    toast.add({ title: 'Invalid code', description: 'Please enter a 6-digit code', color: 'orange' })
    return
  }
  totpVerifying.value = true
  try {
    await api.fetch('/v1/auth/mfa/verify', {
      method: 'POST',
      body: { code: totpVerifyCode.value, mfa_token: '' }
    })
    toast.add({ title: 'Authenticator app configured', color: 'green' })
    totpModalOpen.value = false
    await fetchMFAMethods()
  } catch (err) {
    toast.add({ title: 'Verification failed', description: err.data?.message || 'Invalid code', color: 'red' })
  } finally {
    totpVerifying.value = false
  }
}

async function handleEnableMethod(method) {
  try {
    await api.fetch(`/v1/auth/mfa/methods/${method}/enable`, { method: 'POST' })
    toast.add({ title: `${mfaMethodLabel(method)} enabled`, color: 'green' })
    await fetchMFAMethods()
  } catch (err) {
    toast.add({ title: 'Failed to enable method', description: err.data?.message || 'Try again', color: 'red' })
  }
}

async function handleRemoveMethod(method) {
  try {
    await api.fetch(`/v1/auth/mfa/methods/${method}`, { method: 'DELETE' })
    toast.add({ title: `${mfaMethodLabel(method)} removed`, color: 'green' })
    await fetchMFAMethods()
  } catch (err) {
    toast.add({ title: 'Failed to remove method', description: err.data?.message || 'Try again', color: 'red' })
  }
}

onMounted(fetchMFAMethods)
</script>
