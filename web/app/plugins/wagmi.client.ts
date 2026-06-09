import { WagmiPlugin, createConfig, http, injected } from '@wagmi/vue'
import { mainnet } from '@wagmi/vue/chains'
import { VueQueryPlugin } from '@tanstack/vue-query'

const config = createConfig({
  chains: [mainnet],
  connectors: [injected()],
  transports: {
    [mainnet.id]: http(),
  },
})

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(WagmiPlugin, { config })
  nuxtApp.vueApp.use(VueQueryPlugin)
})
