import { useAccount, useConnect, useDisconnect, useSignMessage } from '@wagmi/vue'
import { SiweMessage } from 'siwe-viem'
import { useApi } from '~/composables/useApi'

export const useWeb3Auth = () => {
  const api = useApi()
  const { address, chainId, isConnected } = useAccount()
  const { connect, connectors } = useConnect()
  const { disconnect } = useDisconnect()
  const { signMessageAsync } = useSignMessage()

  async function signIn() {
    if (!address.value || !chainId.value) {
      throw new Error('Wallet not connected')
    }
    const namespace = 'default'

    // 1. Request nonce from backend.
    const { nonce, domain, uri } = await api.fetch('/v1/auth/web3/nonce', {
      method: 'POST',
      body: { namespace, wallet: address.value },
    })

    // 2. Build SIWE message.
    const message = new SiweMessage({
      domain,
      address: address.value,
      uri,
      version: '1',
      chainId: chainId.value,
      nonce,
      issuedAt: new Date().toISOString(),
      statement: 'Sign in with Ethereum to the app.',
    }).prepareMessage()

    // 3. Sign.
    const signature = await signMessageAsync({ message })

    // 4. Verify on backend.
    const tokens = await api.fetch('/v1/auth/web3/verify', {
      method: 'POST',
      body: { namespace, message, signature },
    })

    return tokens
  }

  return { address, chainId, isConnected, connect, connectors, disconnect, signIn }
}
