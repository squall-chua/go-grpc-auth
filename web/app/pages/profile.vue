<template>
  <div class="max-w-2xl mx-auto space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight">Profile</h1>
      <p class="text-slate-500">Manage your account settings</p>
    </header>

    <UCard>
      <template #header>
        <h3 class="font-semibold">Account Information</h3>
      </template>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div><span class="text-slate-500">Username:</span> {{ auth.user?.username }}</div>
        <div><span class="text-slate-500">Email:</span> {{ auth.user?.email }}</div>
        <div><span class="text-slate-500">Namespace:</span> {{ auth.user?.namespace }}</div>
        <div>
          <span class="text-slate-500">Roles:</span>
          <UBadge v-for="role in auth.user?.roles" :key="role" color="primary" variant="subtle" size="xs" class="ml-1">{{ role }}</UBadge>
        </div>
      </div>
    </UCard>

    <UCard>
      <template #header>
        <h3 class="font-semibold">Change Password</h3>
      </template>
      <form @submit.prevent="handleChangePassword" class="space-y-4">
        <UFormGroup label="Current Password">
          <UInput v-model="passwordForm.current_password" type="password" placeholder="Current password" />
        </UFormGroup>
        <UFormGroup label="New Password">
          <UInput v-model="passwordForm.new_password" type="password" placeholder="New password" />
        </UFormGroup>
        <UFormGroup label="Confirm New Password">
          <UInput v-model="confirmPassword" type="password" placeholder="Confirm new password" />
        </UFormGroup>
        <UButton type="submit" :loading="changingPassword">Change Password</UButton>
      </form>
    </UCard>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
import { useApi } from '~/composables/useApi'

definePageMeta({ middleware: 'auth' })

const auth = useAuthStore()
const api = useApi()
const toast = useToast()

const changingPassword = ref(false)
const confirmPassword = ref('')
const passwordForm = reactive({ current_password: '', new_password: '' })

async function handleChangePassword() {
  if (passwordForm.new_password !== confirmPassword.value) {
    toast.add({ title: 'Passwords do not match', color: 'red' })
    return
  }
  changingPassword.value = true
  try {
    await api.fetch('/v1/auth/change-password', { method: 'POST', body: passwordForm })
    toast.add({ title: 'Password changed successfully', color: 'green' })
    passwordForm.current_password = ''
    passwordForm.new_password = ''
    confirmPassword.value = ''
  } catch (err) {
    toast.add({ title: 'Failed to change password', description: err.data?.message || 'Check your current password', color: 'red' })
  } finally {
    changingPassword.value = false
  }
}
</script>
