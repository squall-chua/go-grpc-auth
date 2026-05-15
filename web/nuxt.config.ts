// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ["@nuxt/ui", "@pinia/nuxt"],
  ssr: false,
  ui: {
    global: true,
  },
  colorMode: {
    preference: "dark"
  },
  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || "http://localhost:8080/",
      name: process.env.NUXT_PUBLIC_APP_NAME || "Go-gRPC-Auth"
    }
  }
})
