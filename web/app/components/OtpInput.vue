<template>
  <div class="flex items-center justify-center gap-2">
    <input
      v-for="(_, i) in digits"
      :key="i"
      ref="inputRefs"
      type="text"
      inputmode="numeric"
      autocomplete="one-time-code"
      maxlength="1"
      :value="digits[i]"
      class="w-12 h-14 text-center text-2xl font-mono font-semibold rounded-lg border-2 border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 focus:bg-white dark:focus:bg-slate-900 outline-none transition-all duration-200"
      @input="handleInput(i, $event)"
      @keydown="handleKeydown(i, $event)"
      @paste="handlePaste"
      @focus="$event.target.select()"
    />
  </div>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  modelValue: string
  length?: number
}>(), {
  length: 6,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const inputRefs = ref<HTMLInputElement[]>([])
const digits = reactive<string[]>(Array(props.length).fill(''))

watch(() => props.modelValue, (val) => {
  const chars = (val || '').split('')
  for (let i = 0; i < props.length; i++) {
    digits[i] = chars[i] || ''
  }
}, { immediate: true })

function emitValue() {
  emit('update:modelValue', digits.join(''))
}

function handleInput(index: number, event: Event) {
  const input = event.target as HTMLInputElement
  const value = input.value.replace(/\D/g, '')

  if (value.length === 0) {
    digits[index] = ''
    emitValue()
    return
  }

  digits[index] = value[0]
  emitValue()

  if (index < props.length - 1) {
    nextTick(() => inputRefs.value[index + 1]?.focus())
  }
}

function handleKeydown(index: number, event: KeyboardEvent) {
  if (event.key === 'Backspace') {
    if (digits[index] === '' && index > 0) {
      event.preventDefault()
      digits[index - 1] = ''
      emitValue()
      nextTick(() => inputRefs.value[index - 1]?.focus())
    }
  } else if (event.key === 'ArrowLeft' && index > 0) {
    event.preventDefault()
    inputRefs.value[index - 1]?.focus()
  } else if (event.key === 'ArrowRight' && index < props.length - 1) {
    event.preventDefault()
    inputRefs.value[index + 1]?.focus()
  }
}

function handlePaste(event: ClipboardEvent) {
  event.preventDefault()
  const pasted = (event.clipboardData?.getData('text') || '').replace(/\D/g, '').slice(0, props.length)
  for (let i = 0; i < props.length; i++) {
    digits[i] = pasted[i] || ''
  }
  emitValue()
  const focusIndex = Math.min(pasted.length, props.length - 1)
  nextTick(() => inputRefs.value[focusIndex]?.focus())
}

function focus() {
  inputRefs.value[0]?.focus()
}

defineExpose({ focus })
</script>
