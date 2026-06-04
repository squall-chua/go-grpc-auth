# 0002: SIWE EOA-only in v1; defer EIP-1271/6492

**Status:** Accepted (2026-06-01)

## Context

EOA wallets (regular Ethereum accounts) verify via secp256k1 signature
recovery. Smart-contract wallets (Safe, Argent) need a different verification
path: EIP-1271 calls `isValidSignature` on the contract, and EIP-6492 handles
counterfactual (not-yet-deployed) wallets.

## Decision

v1 supports EOA only. When a SIWE message is signed by an address with code
at that address, the server returns "smart-contract wallets are not
supported in v1" and audits `WEB3_SIGNIN_FAILED { reason:
"smart_contract_wallet" }`.

The `IsEOA` function exists in `internal/service/auth/web3_siwe.go` so v2
can swap its body (e.g., to call an RPC `eth_getCode`) without changing any
call site.

## Consequences

- EOA covers ~95% of wallets in active use today.
- v2 can add EIP-1271 by giving `IsEOA` access to a chain RPC client and
  using siwe-go's `Verify` overload, or by extending our wrapper.
- The verification code path has no chain RPC dependency in v1.
