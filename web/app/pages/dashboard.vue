<template>
  <div class="space-y-8">
    <header class="flex items-end justify-between">
      <div>
        <h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p class="text-slate-500">Welcome back, {{ auth.user?.username }}</p>
      </div>
      <div class="text-sm font-heading bg-slate-100 dark:bg-slate-800 px-3 py-1 rounded-lg border border-slate-200 dark:border-slate-700">
        Namespace: <span class="text-cta">{{ auth.user?.namespace }}</span>
      </div>
    </header>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <UCard>
        <template #header>
          <div class="flex items-center gap-2 font-semibold">
            <UIcon name="i-heroicons-user-group" class="text-primary-500" />
            Roles
          </div>
        </template>
        <div class="flex flex-wrap gap-2">
          <UBadge v-for="role in auth.user?.roles" :key="role" color="primary" variant="subtle">
            {{ role }}
          </UBadge>
        </div>
      </UCard>

      <UCard class="md:col-span-2">
        <template #header>
          <div class="flex items-center gap-2 font-semibold">
            <UIcon name="i-heroicons-key" class="text-primary-500" />
            Permissions
          </div>
        </template>
        <div class="flex flex-wrap gap-2">
          <UBadge v-for="perm in auth.user?.permissions" :key="perm" color="gray" variant="soft">
            {{ perm }}
          </UBadge>
        </div>
      </UCard>
    </div>

    <UCard>
      <template #header>
        <h3 class="font-semibold">Session Details</h3>
      </template>
      <pre class="text-xs font-mono p-4 bg-slate-950 text-slate-300 rounded overflow-auto">{{ JSON.stringify(auth.user, null, 2) }}</pre>
    </UCard>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()

definePageMeta({
  middleware: 'auth'
})
</script>
