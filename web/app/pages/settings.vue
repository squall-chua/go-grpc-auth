<template>
  <div class="max-w-2xl mx-auto space-y-8">
    <header>
      <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
      <p class="text-slate-500">Application preferences</p>
    </header>

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
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'

definePageMeta({ middleware: 'auth' })

const auth = useAuthStore()
const colorMode = useColorMode()
</script>
