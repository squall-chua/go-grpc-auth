<template>
  <header class="mb-8">
    <p
      v-if="eyebrow"
      class="text-xs font-heading font-medium uppercase tracking-wider text-code-text mb-2"
    >
      {{ eyebrow }}
    </p>
    <h1 class="text-[32px] leading-[1.15] font-heading font-semibold text-text tracking-tight">
      {{ title }}
    </h1>
    <p v-if="subtitle" class="mt-2 text-sm text-text-muted">
      {{ subtitle }}
    </p>
    <div v-if="$slots.actions" class="mt-4 flex items-center gap-2">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  eyebrow?: string
  title: string
  subtitle?: string
}>()
</script>
