export default defineAppConfig({
  ui: {
    primary: 'blue',
    gray: 'slate',
    button: {
      default: {
        size: 'md',
        color: 'primary',
        variant: 'solid'
      },
      rounded: 'rounded-lg'
    },
    input: {
      default: {
        size: 'md',
      },
      rounded: 'rounded-lg'
    },
    card: {
      rounded: 'rounded-xl',
      shadow: 'shadow-xl',
      background: 'bg-white dark:bg-slate-900',
      ring: 'ring-1 ring-slate-200 dark:ring-slate-800'
    }
  }
})
