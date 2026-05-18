<template>
  <div class="min-h-screen bg-background dark:bg-slate-950 text-text dark:text-slate-100 font-body selection:bg-primary-500 selection:text-white">
    <header class="sticky top-0 z-40 w-full backdrop-blur-md bg-white/70 dark:bg-slate-900/70 border-b border-slate-200 dark:border-slate-800">
      <div class="container mx-auto px-4 h-16 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center text-white shadow-lg shadow-primary-500/20">
            <UIcon name="i-heroicons-shield-check-20-solid" class="w-5 h-5" />
          </div>
          <span class="text-xl font-heading font-bold tracking-tight">{{ config.public.name }}<span class="text-cta">Auth</span></span>
        </div>
        
        <nav class="hidden md:flex items-center gap-6">
          <UButton v-if="!auth.isAuthenticated" variant="ghost" to="/login">Login</UButton>
          <UButton v-if="!auth.isAuthenticated" to="/register" color="orange">Get Started</UButton>
          
          <template v-if="auth.isAuthenticated">
            <UButton variant="ghost" to="/dashboard">Dashboard</UButton>
            <UDropdown v-if="auth.user?.roles?.includes('superadmin') || auth.user?.roles?.includes('admin')" :items="adminMenuItems" :popper="{ placement: 'bottom-start' }">
              <UButton variant="ghost">Admin</UButton>
            </UDropdown>
            
            <UDropdown :items="userMenuItems" :popper="{ placement: 'bottom-end' }">
              <UAvatar :alt="auth.user?.username" size="sm" class="cursor-pointer border-2 border-primary-500/50" />
            </UDropdown>
          </template>
          
          <UButton
            icon="i-heroicons-moon-20-solid"
            color="gray"
            variant="ghost"
            aria-label="Theme"
            @click="toggleColorMode"
          />
        </nav>
      </div>
    </header>

    <main class="container mx-auto px-4 py-8">
      <slot />
    </main>

    <footer class="border-t border-slate-200 dark:border-slate-800 py-12 bg-white dark:bg-slate-900">
      <div class="container mx-auto px-4 text-center text-slate-500 text-sm">
        <p>&copy; 2026 {{ config.public.name }} Auth. Built with Precision.</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
const config = useRuntimeConfig()
const auth = useAuthStore()
const colorMode = useColorMode()

const toggleColorMode = () => {
  colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
}

const adminMenuItems = [
  [
    { label: 'Overview', icon: 'i-heroicons-squares-2x2', to: '/admin' },
    { label: 'Users', icon: 'i-heroicons-user-group', to: '/admin/users' },
    { label: 'Roles', icon: 'i-heroicons-shield-check', to: '/admin/roles' },
    { label: 'Permissions', icon: 'i-heroicons-key', to: '/admin/permissions' },
  ],
  [
    { label: 'Namespaces', icon: 'i-heroicons-cube-transparent', to: '/admin/namespaces' },
    { label: 'OIDC Clients', icon: 'i-heroicons-globe-alt', to: '/admin/clients' },
    { label: 'Audit Logs', icon: 'i-heroicons-document-text', to: '/admin/audit' },
  ]
]

const userMenuItems = [
  [{
    label: auth.user?.username || 'User',
    slot: 'account',
    disabled: true
  }],
  [{
    label: 'Profile',
    icon: 'i-heroicons-user',
    to: '/profile'
  }, {
    label: 'Settings',
    icon: 'i-heroicons-cog-8-tooth',
    to: '/settings'
  }],
  [{
    label: 'Sign out',
    icon: 'i-heroicons-arrow-left-on-rectangle',
    click: () => auth.logout()
  }]
]
</script>
