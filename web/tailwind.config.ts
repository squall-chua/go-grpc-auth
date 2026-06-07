import type { Config } from 'tailwindcss'

export default {
  theme: {
    extend: {
      colors: {
        // Semantic tokens — values resolved by CSS variables in app.vue
        background: 'var(--bg)',
        'background-elevated': 'var(--bg-elevated)',
        'background-input': 'var(--bg-input)',
        border: 'var(--border)',
        text: 'var(--text)',
        'text-muted': 'var(--text-muted)',
        'text-subtle': 'var(--text-subtle)',
        cta: 'var(--accent)',
        'cta-hover': 'var(--accent-hover)',
        'cta-fg': 'var(--accent-fg)',
        'cta-subtle': 'var(--accent-subtle)',
        'code-text': 'var(--code-text)',
      },
      fontFamily: {
        heading: ['"Fira Code"', 'monospace'],
        body: ['"Fira Sans"', 'sans-serif'],
        mono: ['"Fira Code"', 'monospace'],
      },
    },
  },
} satisfies Partial<Config>
