<template>
  <div class="ring-1 ring-border rounded-xl overflow-hidden bg-background-input">
    <button
      type="button"
      class="w-full flex items-center justify-between px-4 py-2 text-xs font-heading text-text-muted hover:bg-cta-subtle/40 transition-colors"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span class="font-mono">{{ label }}</span>
      <UIcon
        :name="open ? 'i-heroicons-chevron-up' : 'i-heroicons-chevron-down'"
        class="w-4 h-4"
      />
    </button>
    <pre
      v-if="open"
      class="font-mono text-xs text-text-muted p-4 overflow-auto max-h-96 whitespace-pre-wrap break-all"
    >{{ formatted }}</pre>
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  data: unknown
  label?: string
  defaultOpen?: boolean
}>(), {
  label: 'data',
  defaultOpen: false,
})

const open = ref(props.defaultOpen)

const formatted = computed(() => {
  if (props.data == null) return ''
  if (typeof props.data === 'string') {
    try {
      return JSON.stringify(JSON.parse(props.data), null, 2)
    } catch {
      return props.data
    }
  }
  return JSON.stringify(props.data, null, 2)
})
</script>
