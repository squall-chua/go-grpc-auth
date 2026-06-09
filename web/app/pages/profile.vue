<template>
  <div class="max-w-2xl mx-auto space-y-8">
    <PageHeader
      eyebrow="Account"
      title="Profile"
      subtitle="Manage your account settings"
    />

    <UCard>
      <template #header>
        <h3 class="font-semibold">Account Information</h3>
      </template>
      <dl class="space-y-3 text-sm">
        <div class="flex items-center gap-3">
          <dt class="font-heading text-xs uppercase tracking-wider text-text-muted w-28 shrink-0">Username</dt>
          <dd class="text-text">{{ auth.user?.username }}</dd>
        </div>
        <div class="flex items-center gap-3">
          <dt class="font-heading text-xs uppercase tracking-wider text-text-muted w-28 shrink-0">Email</dt>
          <dd class="text-text">{{ auth.user?.email }}</dd>
        </div>
        <div class="flex items-center gap-3">
          <dt class="font-heading text-xs uppercase tracking-wider text-text-muted w-28 shrink-0">Namespace</dt>
          <dd class="font-mono text-text">{{ auth.user?.namespace }}</dd>
        </div>
        <div class="flex items-start gap-3">
          <dt class="font-heading text-xs uppercase tracking-wider text-text-muted w-28 shrink-0">Roles</dt>
          <dd class="flex flex-wrap gap-1">
            <UBadge v-for="role in auth.user?.roles" :key="role" color="primary" variant="subtle" size="xs">{{ role }}</UBadge>
            <span v-if="!auth.user?.roles?.length" class="text-text-subtle">No roles</span>
          </dd>
        </div>
      </dl>
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
