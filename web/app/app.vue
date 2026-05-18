<template>
  <div class="min-h-screen bg-background dark:bg-gray-950 font-body text-text dark:text-gray-100">
    <UNotifications />
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap');

:root {
  --color-primary: #2563EB;
  --color-secondary: #3B82F6;
  --color-cta: #F97316;
  --color-background: #F8FAFC;
  --color-text: #1E293B;

  --space-xs: 0.25rem;
  --space-sm: 0.5rem;
  --space-md: 1rem;
  --space-lg: 1.5rem;
  --space-xl: 2rem;
  --space-2xl: 3rem;
  --space-3xl: 4rem;

  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 4px 6px rgba(0,0,0,0.1);
  --shadow-lg: 0 10px 15px rgba(0,0,0,0.1);
  --shadow-xl: 0 20px 25px rgba(0,0,0,0.15);
}

body {
  font-family: 'Fira Sans', sans-serif;
}

h1, h2, h3, h4, h5, h6 {
  font-family: 'Fira Code', monospace;
}

/* Ensure all clickable elements have pointer cursor */
button, a, [role="button"], .cursor-pointer {
  cursor: pointer;
}

/* Smooth transitions on interactive elements */
button, a, input, select, textarea {
  transition: all 200ms ease;
}

/* Focus states for accessibility */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

/* Respect reduced motion */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
</style>
