# go-grpc-auth

Multi-tenant (namespace-scoped) gRPC + grpc-gateway HTTP service providing
password auth, OAuth social (Google, GitHub), OIDC provider, MFA, and web3
wallet sign-in via SIWE.

## Language

**Wallet**:
A EIP-55 checksummed Ethereum address used as a sign-in identity.
_Avoid_: EOA, account, public key

**SIWE (Sign-In with Ethereum)**:
EIP-4361 message format used for wallet-based authentication. Server verifies
an EIP-191 `personal_sign` signature over a structured plain-text message.
_Avoid_: signed message, web3 login

**Nonce**:
A server-issued, single-use, 10-minute-TTL random token bound to
`(namespace, wallet_address)`. Consumed atomically with successful signature
verification.
_Avoid_: challenge, CSRF token

**ChainId**:
The EVM network identifier in a SIWE message (`1`=Ethereum, `8453`=Base, etc.).
Verified against the namespace's `AllowedWeb3ChainIds`. Recorded in the audit
log and ID token, but not part of the wallet identity.
_Avoid_: network, chain ID

**Web3AuthService**:
The gRPC service exposing web3-specific RPCs (RequestNonce, Verify,
ListWallets, LinkWallet, UnlinkWallet). Distinct from AuthService so web3
can be feature-flagged and extended (EIP-1271, Solana) without bloating
AuthService.
