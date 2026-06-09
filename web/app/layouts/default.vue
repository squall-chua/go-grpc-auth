<template>
  <div class="min-h-screen bg-background text-text font-body selection:bg-cta selection:text-cta-fg">
    <header class="sticky top-0 z-40 w-full backdrop-blur-md bg-background/70 border-b border-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-cta rounded-lg flex items-center justify-center text-cta-fg">
            <UIcon name="i-heroicons-shield-check-20-solid" class="w-5 h-5" />
          </div>
          <NuxtLink to="/" class="text-lg font-heading font-semibold tracking-tight text-text">
            go<span class="text-cta">Auth</span>
          </NuxtLink>
        </div>

        <nav class="hidden md:flex items-center gap-6">
          <template v-if="!auth.isAuthenticated">
            <UButton variant="ghost" to="/login">Sign in</UButton>
            <UButton to="/register">Get started</UButton>
          </template>

          <template v-else>
            <NuxtLink
              to="/dashboard"
              class="text-sm font-medium transition-colors py-5 border-b-2"
              :class="route.path === '/dashboard' ? 'text-text border-cta' : 'text-text-muted hover:text-text border-transparent'"
            >
              Dashboard
            </NuxtLink>
            <UDropdown
              v-if="isAdmin"
              :items="adminDropdownItems"
              :popper="{ placement: 'bottom-start' }"
            >
              <button
                class="text-sm font-medium transition-colors py-5 border-b-2 border-transparent"
                :class="isAdminActive ? 'text-text border-cta' : 'text-text-muted hover:text-text'"
              >
                Admin
                <UIcon name="i-heroicons-chevron-down" class="w-3 h-3 inline-block ml-0.5" />
              </button>
            </UDropdown>

            <UDropdown :items="userDropdownItems" :popper="{ placement: 'bottom-end' }">
              <UAvatar
                :alt="auth.user?.username"
                size="sm"
                class="cursor-pointer"
              />
            </UDropdown>
          </template>

          <UButton
            :icon="colorMode.value === 'dark' ? 'i-heroicons-sun-20-solid' : 'i-heroicons-moon-20-solid'"
            color="gray"
            variant="ghost"
            aria-label="Toggle theme"
            @click="toggleColorMode"
          />
        </nav>

        <div class="md:hidden flex items-center gap-2">
          <UButton
            :icon="colorMode.value === 'dark' ? 'i-heroicons-sun-20-solid' : 'i-heroicons-moon-20-solid'"
            color="gray"
            variant="ghost"
            aria-label="Toggle theme"
            @click="toggleColorMode"
          />
          <template v-if="!auth.isAuthenticated">
            <UButton variant="ghost" size="sm" to="/login">Sign in</UButton>
            <UButton size="sm" to="/register">Get started</UButton>
          </template>
          <MobileMenu v-else />
        </div>
      </div>
    </header>

    <main id="main" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 md:py-12">
      <slot />
    </main>

    <footer class="border-t border-border bg-background-elevated">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 text-center text-xs text-text-subtle">
        © 2026 goAuth · Built with Nuxt
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()
const colorMode = useColorMode()
const route = useRoute()

const isAdmin = computed(() => auth.user?.roles?.some((r: string) => r === 'admin' || r === 'superadmin'))

const isAdminActive = computed(() => route.path.startsWith('/admin'))

const toggleColorMode = () => {
  colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
}

const adminDropdownItems = [[
  { label: 'Overview', icon: 'i-heroicons-squares-2x2', to: '/admin' },
  { label: 'Users', icon: 'i-heroicons-user-group', to: '/admin/users' },
  { label: 'Roles', icon: 'i-heroicons-shield-check', to: '/admin/roles' },
  { label: 'Permissions', icon: 'i-heroicons-key', to: '/admin/permissions' },
  { label: 'Namespaces', icon: 'i-heroicons-cube-transparent', to: '/admin/namespaces' },
  { label: 'OIDC Clients', icon: 'i-heroicons-globe-alt', to: '/admin/clients' },
  { label: 'Audit Logs', icon: 'i-heroicons-document-text', to: '/admin/audit' },
]]

const userDropdownItems = computed(() => [
  [{ label: auth.user?.username || 'User', disabled: true, slot: 'account' as const }],
  [
    { label: 'Profile', icon: 'i-heroicons-user', to: '/profile' },
    { label: 'Settings', icon: 'i-heroicons-cog-8-tooth', to: '/settings' },
  ],
  [
    { label: 'Sign out', icon: 'i-heroicons-arrow-left-on-rectangle', click: () => auth.logout() },
  ],
])
</script>
