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
      rounded: 'rounded-xl',
      shadow: '',
      background: 'bg-background-elevated',
      ring: 'ring-1 ring-border',
      divide: '',
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
        base: 'hover:bg-cta-subtle/30 transition-colors',
      },
      loadingState: {
        base: 'py-12 text-center text-text-muted',
      },
      emptyState: {
        base: 'py-12 text-center text-text-muted',
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
      base: 'overflow-hidden',
      title: 'font-semibold text-sm',
      description: 'text-sm text-text-muted',
    },
    tooltip: {
      rounded: 'rounded-lg',
      shadow: '',
      background: 'bg-text text-background',
      base: 'text-xs font-normal',
    },
    checkbox: {
      rounded: 'rounded',
      ring: 'ring-1 ring-border',
      base: 'h-5 w-5 transition-colors',
    },
    radio: {
      ring: 'ring-1 ring-border',
      base: 'h-5 w-5 transition-colors',
    },
    link: {
      base: 'text-cta hover:text-cta-hover hover:underline transition-colors',
    },
    avatar: {
      rounded: 'rounded-lg',
      background: 'bg-cta-subtle text-cta-hover ring-1 ring-border',
      font: 'font-heading font-semibold',
    },
    pagination: {
      rounded: 'rounded-md',
      shadow: '',
      default: {
        size: 'md',
        color: 'white',
      },
      button: {
        base: 'transition-colors',
        active: 'bg-cta text-cta-fg',
        inactive: 'bg-background-elevated text-text ring-1 ring-border hover:bg-cta-subtle',
      },
    },
    divider: {
      base: 'border-border',
    },
    select: {
      rounded: 'rounded-lg',
    },
  },
})
