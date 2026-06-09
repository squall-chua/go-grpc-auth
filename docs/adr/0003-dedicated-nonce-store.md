# 0003: Server-issued, TTL-bounded nonces via a dedicated NonceStore

**Status:** Accepted (2026-06-01)

## Context

The existing `ratelimit.RateLimiter` is a fixed-window counter; it cannot
store values. SIWE nonces need a value (the random token) with TTL and
single-use semantics.

## Decision

Add a dedicated `NonceStore` interface in `internal/service/auth/` with two
implementations: Redis (when `REDIS_URI` is set) and in-memory (fallback).
The interface is `Save(ctx, ns, wallet, nonce, ttl)` and `Consume(ctx, ns,
wallet, nonce) (bool, error)`. `Consume` is atomic (read-then-delete) so
the nonce is single-use even under concurrent Verify calls.

## Consequences

- No new dependency; Redis is already wired in for rate limiting.
- Nonces bound to `(namespace, wallet)` as defense in depth (signature
  already proves ownership; binding prevents confused-deputy issues).
- In-memory fallback loses state on restart, so a wallet in mid-login has
  to re-request a nonce. Acceptable.
