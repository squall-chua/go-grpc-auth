<template>
  <UDropdown
    v-if="auth.isAuthenticated"
    :items="authItems"
    :popper="{ placement: 'bottom-end', offsetDistance: 8 }"
    :ui="{ width: 'w-72' }"
  >
    <UButton
      icon="i-heroicons-bars-3"
      color="gray"
      variant="ghost"
      aria-label="Open menu"
    />
  </UDropdown>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()

const adminItems = [
  { label: 'Overview', icon: 'i-heroicons-squares-2x2', to: '/admin' },
  { label: 'Users', icon: 'i-heroicons-user-group', to: '/admin/users' },
  { label: 'Roles', icon: 'i-heroicons-shield-check', to: '/admin/roles' },
  { label: 'Permissions', icon: 'i-heroicons-key', to: '/admin/permissions' },
  { label: 'Namespaces', icon: 'i-heroicons-cube-transparent', to: '/admin/namespaces' },
  { label: 'OIDC Clients', icon: 'i-heroicons-globe-alt', to: '/admin/clients' },
  { label: 'Audit Logs', icon: 'i-heroicons-document-text', to: '/admin/audit' },
]

const authItems = computed(() => {
  const isAdmin = auth.user?.roles?.some((r: string) => r === 'admin' || r === 'superadmin')
  const sections: any[][] = [
    [
      { label: 'Dashboard', to: '/dashboard', icon: 'i-heroicons-home' },
    ],
  ]
  if (isAdmin) {
    sections.push(adminItems)
  }
  sections.push([
    { label: 'Profile', to: '/profile', icon: 'i-heroicons-user' },
    { label: 'Settings', to: '/settings', icon: 'i-heroicons-cog-8-tooth' },
  ])
  sections.push([
    { label: 'Sign out', icon: 'i-heroicons-arrow-left-on-rectangle', click: () => auth.logout() },
  ])
  return sections
})
</script>
