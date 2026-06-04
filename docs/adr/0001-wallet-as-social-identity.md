# 0001: Treat web3 wallet as a SocialIdentity

**Status:** Accepted (2026-06-01)

## Context

Web3 sign-in needs a way to relate an Ethereum address to a User. Options:
(a) new domain entity, (b) primary credential (no email), (c) reuse
`SocialIdentity` with provider="ethereum".

## Decision

Option (c). Wallet addresses are stored as `SocialIdentity { Provider:
"ethereum", ExternalID: address }`. First wallet login auto-provisions a User
with synthetic email `0x{address-lowercase}@wallet.local`. Multiple wallets
per user are allowed. Linking/unlinking is symmetric with OAuth social.

## Consequences

- No new domain entities; no schema migration beyond a new provider value.
- Reuse existing `userRepo.GetByEmail` for the auto-provision path.
- Synthetic emails are not deliverable; password recovery for wallet-only
  users is a future problem.
- The `User.SocialIdentities` array queries use `dot.notation` on
  `SocialIdentities.0.*`, which checks only the first element. Acceptable
  for v1 because we always link on find-time, putting the identity at
  index 0.
