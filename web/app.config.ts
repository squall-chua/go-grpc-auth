export default defineAppConfig({
  ui: {
    primary: 'amber',
    gray: 'zinc',
    icons: {
      dynamic: true,
    },
    button: {
      default: {
        size: 'md',
        color: 'primary',
        variant: 'solid',
      },
      rounded: 'rounded-lg',
      font: 'font-semibold',
      base: 'transition-colors duration-150',
    },
    input: {
      default: {
        size: 'md',
      },
      rounded: 'rounded-lg',
      base: 'transition-colors duration-150',
    },
    textarea: {
      rounded: 'rounded-lg',
    },
    formGroup: {
      label: {
        base: 'font-semibold text-sm text-text',
      },
    },
    card: {
      rounded: 'rounded-xl',
      shadow: '',
      background: 'bg-background-elevated',
      ring: 'ring-1 ring-border',
      divide: '',
      base: 'transition-colors duration-150',
      header: {
        base: 'font-heading',
        padding: 'px-5 py-4',
      },
      body: {
        padding: 'px-5 py-4',
      },
      footer: {
        padding: 'px-5 py-4',
      },
    },
    modal: {
      rounded: 'rounded-2xl',
      shadow: '',
      background: 'bg-background-elevated',
      overlay: {
        background: 'bg-black/50 backdrop-blur-sm',
      },
      padding: 'p-6',
    },
    table: {
      divide: '',
      thead: '',
      tbody: '',
      header: {
        base: 'font-heading text-xs uppercase tracking-wider text-text-muted',
        padding: 'px-4 py-3',
      },
      th: {
        base: 'font-semibold text-xs uppercase tracking-wider text-text-muted',
        padding: 'px-4 py-3',
        color: 'text-text-muted',
      },
      td: {
        base: 'whitespace-nowrap text-sm',
        padding: 'px-4 py-3',
      },
      tr: {
        base: '',
        active: 'hover:bg-cta-subtle/30 transition-colors',
      },
      loadingState: {
        wrapper: 'py-12 text-center text-text-muted',
        label: 'text-sm',
      },
      emptyState: {
        wrapper: 'py-12 text-center text-text-muted',
        label: 'text-sm',
      },
    },
    badge: {
      rounded: 'rounded-md',
      font: 'font-semibold',
      base: 'inline-flex items-center',
    },
    dropdown: {
      rounded: 'rounded-xl',
      shadow: '',
      background: 'bg-background-elevated',
      ring: 'ring-1 ring-border',
      item: {
        base: 'cursor-pointer transition-colors',
        active: 'bg-cta-subtle text-cta-hover',
        inactive: 'text-text',
        color: 'text-text-muted',
        icon: {
          active: 'text-cta',
          inactive: 'text-text-muted',
        },
      },
      container: 'z-50',
      transition: {
        enterActiveClass: 'transition duration-150 ease-out',
        enterFromClass: 'opacity-0 -translate-y-1',
        enterToClass: 'opacity-100 translate-y-0',
        leaveActiveClass: 'transition duration-100 ease-in',
        leaveFromClass: 'opacity-100',
        leaveToClass: 'opacity-0',
      },
    },
    notification: {
      rounded: 'rounded-lg',
      shadow: '',
      background: 'bg-background-elevated',
      ring: 'ring-1 ring-border',
      title: 'font-semibold text-sm',
      description: 'text-sm text-text-muted',
    },
    tooltip: {
      rounded: 'rounded-lg',
      shadow: '',
      background: 'bg-background-elevated text-text dark:bg-text dark:text-background',
      base: 'text-xs font-normal',
    },
    checkbox: {
      rounded: 'rounded',
      ring: 'ring-1 ring-border',
      base: 'h-5 w-5 transition-colors',
    },
    radio: {
      ring: 'ring-1 ring-border',
      base: 'h-5 w-5 rounded-full transition-colors',
    },
    // ULink has no ui.link config in Nuxt UI v2; styled per-instance via `class` prop,
    // or via a global `a` rule in app.vue. Skipped here.
    avatar: {
      rounded: 'rounded-lg',
      background: 'bg-cta-subtle',
      text: 'font-heading font-semibold text-cta-hover leading-none truncate',
      ring: 'ring-[1.5px] ring-border',
    },
    pagination: {
      rounded: 'rounded-md',
      default: {
        size: 'md',
        color: 'white',
        activeButton: {
          color: 'primary',
        },
        inactiveButton: {
          color: 'white',
          class: 'bg-background-elevated text-text ring-1 ring-border hover:bg-cta-subtle',
        },
      },
    },
    divider: {
      border: {
        base: 'border-border',
      },
    },
    select: {
      rounded: 'rounded-lg',
    },
  },
})
