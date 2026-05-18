<template>
  <div class="flex items-center justify-center min-h-[60vh]">
    <ULoadingIcon v-if="auth.isAuthenticated" />
  </div>
</template>

<script setup>
import { useAuthStore } from '~/stores/auth'
const auth = useAuthStore()

onMounted(() => {
  if (auth.isAuthenticated) {
    navigateTo('/dashboard')
  } else {
    navigateTo('/login')
  }
})
</script>
