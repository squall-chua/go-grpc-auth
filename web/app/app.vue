<template>
  <div class="min-h-screen bg-background text-text font-body">
    <a
      href="#main"
      class="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50 focus:px-3 focus:py-2 focus:rounded-lg focus:bg-cta focus:text-cta-fg focus:font-semibold"
    >
      Skip to main content
    </a>
    <UNotifications />
    <NuxtLayout>
      <NuxtTransition>
        <NuxtPage />
      </NuxtTransition>
    </NuxtLayout>
  </div>
</template>

<script setup>
useHead({
  htmlAttrs: { lang: 'en' },
})
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap');

:root {
  /* Light mode (default) */
  --bg: #FAF7F2;
  --bg-elevated: #FFFFFF;
  --bg-input: #FDFCF9;
  --border: #E7E2D9;
  --text: #1C1917;
  --text-muted: #57534E;
  --text-subtle: #A8A29E;
  --accent: #F59E0B;
  --accent-hover: #D97706;
  --accent-fg: #1C1917;
  --accent-subtle: #FEF3C7;
  --code-text: #92400E;
  --success: #16A34A;
  --danger: #DC2626;
  --warning: #D97706;
}

:root.dark {
  /* Dark mode */
  --bg: #18181B;
  --bg-elevated: #27272A;
  --bg-input: #27272A;
  --border: #3F3F46;
  --text: #FAFAFA;
  --text-muted: #A1A1AA;
  --text-subtle: #71717A;
  --accent: #F59E0B;
  --accent-hover: #D97706;
  --accent-fg: #18181B;
  --accent-subtle: rgba(245, 158, 11, 0.12);
  --code-text: #FCD34D;
  --success: #22C55E;
  --danger: #EF4444;
  --warning: #F59E0B;
}

body {
  font-family: 'Fira Sans', sans-serif;
}

h1, h2, h3, h4, h5, h6,
.font-heading {
  font-family: 'Fira Code', monospace;
  font-feature-settings: 'calt' 1, 'liga' 1;
}

.font-mono, code, pre {
  font-family: 'Fira Code', monospace;
  font-variant-numeric: tabular-nums;
}

/* Always-visible focus indicators */
:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: 2px;
}

button, a, [role="button"] {
  cursor: pointer;
}

button, a, input, select, textarea {
  transition: background-color 150ms ease, border-color 150ms ease, color 150ms ease, box-shadow 150ms ease, opacity 150ms ease;
}

@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* Page transition */
.page-enter-active,
.page-leave-active {
  transition: opacity 150ms ease;
}
.page-enter-from,
.page-leave-to {
  opacity: 0;
}
</style>
