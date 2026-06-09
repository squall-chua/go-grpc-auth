import { defineNuxtConfig } from "nuxt/config";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ["@nuxt/ui", "@pinia/nuxt", "@wagmi/vue/nuxt"],

  app: {
    pageTransition: { name: 'page', mode: 'out-in' },
  },

  routeRules: {
    "/**": { ssr: false },
  },

  ui: {
    global: true,
  },

  colorMode: {
    preference: "dark"
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || "http://localhost:8080",
      name: process.env.NUXT_PUBLIC_APP_NAME || "Go"
    }
  },

  compatibilityDate: "2026-05-18"
})
