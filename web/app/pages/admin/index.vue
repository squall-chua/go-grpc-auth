<template>
  <div class="space-y-8">
    <PageHeader
      eyebrow="Administration"
      title="Control Center"
      subtitle="System-wide management and monitoring"
    />

    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <UCard v-for="stat in stats" :key="stat.label" class="relative overflow-hidden group">
        <div class="flex items-center gap-4">
          <div :class="`p-3 rounded-xl bg-${stat.color}-100 dark:bg-${stat.color}-900/30 text-${stat.color}-600 dark:text-${stat.color}-400`">
            <UIcon :name="stat.icon" class="w-6 h-6" />
          </div>
          <div>
            <p class="text-sm font-medium text-slate-500">{{ stat.label }}</p>
            <p class="text-2xl font-bold">{{ stat.value }}</p>
          </div>
        </div>
      </UCard>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <UCard v-for="link in adminLinks" :key="link.to">
        <NuxtLink :to="link.to" class="flex items-center gap-4 p-2 hover:bg-slate-50 dark:hover:bg-slate-800 rounded-lg transition">
          <div :class="`p-3 rounded-xl bg-${link.color}-100 dark:bg-${link.color}-900/30 text-${link.color}-600 dark:text-${link.color}-400`">
            <UIcon :name="link.icon" class="w-6 h-6" />
          </div>
          <div>
            <p class="font-semibold">{{ link.label }}</p>
            <p class="text-sm text-slate-500">{{ link.description }}</p>
          </div>
        </NuxtLink>
      </UCard>
    </div>
  </div>
</template>

<script setup>
import { useApi } from '~/composables/useApi'

definePageMeta({ middleware: 'auth' })

const api = useApi()

const stats = ref([
  { label: 'Total Users', value: '...', icon: 'i-heroicons-user-group', color: 'blue' },
  { label: 'Namespaces', value: '...', icon: 'i-heroicons-cube-transparent', color: 'purple' },
  { label: 'Roles', value: '...', icon: 'i-heroicons-shield-check', color: 'amber' },
  { label: 'OIDC Clients', value: '...', icon: 'i-heroicons-key', color: 'green' },
])

const adminLinks = [
  { to: '/admin/users', label: 'Users', description: 'Manage user accounts', icon: 'i-heroicons-user-group', color: 'blue' },
  { to: '/admin/roles', label: 'Roles', description: 'Manage RBAC roles', icon: 'i-heroicons-shield-check', color: 'amber' },
  { to: '/admin/permissions', label: 'Permissions', description: 'Manage permissions', icon: 'i-heroicons-key', color: 'green' },
  { to: '/admin/namespaces', label: 'Namespaces', description: 'Manage tenants', icon: 'i-heroicons-cube-transparent', color: 'purple' },
  { to: '/admin/clients', label: 'OIDC Clients', description: 'Manage OAuth2 clients', icon: 'i-heroicons-globe-alt', color: 'cyan' },
  { to: '/admin/audit', label: 'Audit Logs', description: 'View security events', icon: 'i-heroicons-document-text', color: 'red' },
]

onMounted(async () => {
  try {
    const [users, namespaces, roles, clients] = await Promise.allSettled([
      api.fetch('/v1/admin/users?page_size=1&namespace=default'),
      api.fetch('/v1/admin/namespaces?page_size=1'),
      api.fetch('/v1/admin/roles?page_size=1&namespace=default'),
      api.fetch('/v1/admin/oidc/clients?page_size=1'),
    ])
    if (users.status === 'fulfilled') stats.value[0].value = String(users.value.total_count || 0)
    if (namespaces.status === 'fulfilled') stats.value[1].value = String(namespaces.value.total_count || 0)
    if (roles.status === 'fulfilled') stats.value[2].value = String(roles.value.total_count || 0)
    if (clients.status === 'fulfilled') stats.value[3].value = String(clients.value.total_count || 0)
  } catch {}
})
</script>
