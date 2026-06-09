<template>
  <div class="space-y-8">
    <PageHeader
      eyebrow="Overview"
      :title="`Welcome back, ${auth.user?.username || ''}`"
      subtitle="Your account at a glance"
    />

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-user-group" class="text-cta" />
            <span>Roles</span>
          </div>
        </template>
        <div class="flex flex-wrap gap-2">
          <UBadge v-for="role in auth.user?.roles" :key="role" color="primary" variant="subtle">
            {{ role }}
          </UBadge>
          <span v-if="!auth.user?.roles?.length" class="text-sm text-text-subtle">No roles assigned</span>
        </div>
      </UCard>

      <UCard class="md:col-span-2">
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-key" class="text-cta" />
            <span>Permissions</span>
          </div>
        </template>
        <div class="flex flex-wrap gap-2">
          <UBadge v-for="perm in auth.user?.permissions" :key="perm" color="gray" variant="soft">
            {{ perm }}
          </UBadge>
          <span v-if="!auth.user?.permissions?.length" class="text-sm text-text-subtle">No permissions assigned</span>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-base font-heading font-semibold">Session Details</h2>
          <UButton
            size="xs"
            variant="ghost"
            :icon="tokenVisible ? 'i-heroicons-eye-slash' : 'i-heroicons-eye'"
            @click="tokenVisible = !tokenVisible"
          >
            {{ tokenVisible ? 'Hide token' : 'View token' }}
          </UButton>
        </div>
      </template>
      <JsonView v-if="tokenVisible" :data="sessionPayload" label="session.json" :default-open="true" />
      <p v-else class="text-sm text-text-muted">
        Token payload is hidden. Click <em>View token</em> to inspect.
      </p>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()

definePageMeta({
  middleware: 'auth'
})

const tokenVisible = ref(false)

const sessionPayload = computed(() => ({
  user: auth.user,
  tokenPresent: !!auth.token,
  mfaPending: auth.mfaRequired,
  mfaMethods: auth.mfaMethods,
}))
</script>
