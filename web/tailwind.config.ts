import type { Config } from 'tailwindcss'

export default {
  theme: {
    extend: {
      colors: {
        background: '#F8FAFC',
        text: '#1E293B',
        cta: '#F97316',
      },
      fontFamily: {
        heading: ['"Fira Code"', 'monospace'],
        body: ['"Fira Sans"', 'sans-serif'],
      },
    },
  },
} satisfies Partial<Config>
