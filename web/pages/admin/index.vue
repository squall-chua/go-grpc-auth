<template>
  <div class="space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">Admin Control Center</h1>
      <p class="text-slate-500 dark:text-slate-400">System-wide management and monitoring</p>
    </header>

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <UCard v-for="stat in stats" :key="stat.label" class="relative overflow-hidden group">
        <div class="flex items-center gap-4">
          <div :class="`p-3 rounded-xl bg-${stat.color}-100 dark:bg-${stat.color}-900/30 text-${stat.color}-600 dark:text-${stat.color}-400 group-hover:scale-110 transition-transform duration-300`">
            <UIcon :name="stat.icon" class="w-6 h-6" />
          </div>
          <div>
            <p class="text-sm font-medium text-slate-500">{{ stat.label }}</p>
            <p class="text-2xl font-bold">{{ stat.value }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="font-semibold">Recent Namespaces</h3>
            <UButton size="xs" variant="ghost" to="/admin/namespaces">View all</UButton>
          </div>
        </template>
        <div class="text-sm text-slate-500 italic">Namespace list loading...</div>
      </UCard>

      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="font-semibold">Security Audit</h3>
            <UButton size="xs" variant="ghost">Logs</UButton>
          </div>
        </template>
        <div class="text-sm text-slate-500 italic">No recent critical events.</div>
      </UCard>
    </div>
  </div>
</template>

<script setup>
definePageMeta({
  middleware: 'auth'
})

const stats = [
  { label: 'Total Users', value: '1,284', icon: 'i-heroicons-user-group', color: 'blue' },
  { label: 'Active Namespaces', value: '12', icon: 'i-heroicons-cube-transparent', color: 'purple' },
  { label: 'Auth Requests (24h)', value: '48.2k', icon: 'i-heroicons-bolt', color: 'amber' },
  { label: 'Security Threats', value: '0', icon: 'i-heroicons-shield-exclamation', color: 'green' }
]
</script>
