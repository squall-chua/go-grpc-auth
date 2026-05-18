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
      rounded: 'rounded-lg',
      font: 'font-semibold',
      base: 'cursor-pointer transition-all duration-200',
    },
    input: {
      default: {
        size: 'md',
      },
      rounded: 'rounded-lg',
      base: 'transition-colors duration-200',
    },
    card: {
      rounded: 'rounded-xl',
      shadow: 'shadow-md hover:shadow-lg',
      background: 'bg-white dark:bg-slate-900',
      ring: 'ring-1 ring-slate-200 dark:ring-slate-800',
      base: 'transition-shadow duration-200',
    },
    modal: {
      rounded: 'rounded-2xl',
      shadow: 'shadow-xl',
      overlay: {
        background: 'bg-black/50 backdrop-blur-sm',
      },
    },
    table: {
      td: {
        base: 'whitespace-nowrap',
        padding: 'px-4 py-3',
      },
      th: {
        base: 'font-semibold',
        padding: 'px-4 py-3',
      },
    },
    badge: {
      rounded: 'rounded-md',
    },
    dropdown: {
      rounded: 'rounded-lg',
      shadow: 'shadow-lg',
      item: {
        base: 'cursor-pointer',
      },
    },
    notification: {
      rounded: 'rounded-lg',
      shadow: 'shadow-lg',
    },
  }
})
