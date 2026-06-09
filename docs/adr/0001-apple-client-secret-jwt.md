# 0001 — Apple Client Secret is a Signed JWT, Not a Static String

## Status

Accepted (2026-06-04).

## Context

Apple's "Sign in with Apple" is the only major social provider whose
`client_secret` is not a static value. Apple requires the `client_secret`
form parameter on the token exchange to be a JWT signed with the developer
team's `.p8` private key (ES256, P-256 curve). The JWT must have `iss` =
team ID, `sub` = service ID, `aud` = `https://appleid.apple.com`, plus
`iat` and `exp` (max 6 months).

## Decision

The Apple provider generates the JWT at startup, holds it in memory under a
mutex, and re-signs it within 1 day of expiry (TTL of 5 days, refresh
threshold 24 hours before expiry). Apple config is opt-in: when
`APPLE_CLIENT_ID` is empty the provider is silently skipped, matching the
existing Google/GitHub pattern. When Apple config is partially populated
(e.g. team ID set but `.p8` missing), startup fails loudly so misconfiguration
cannot pass silently.

## Alternatives Considered

- **Static client secret** — not supported by Apple.
- **Pre-generated external JWT injected via config** — requires operators
  to rotate manually every <6 months, error-prone, adds a 1.0-day blast
  radius for forgotten rotations.
- **Cached JWT to disk** — survives restarts faster but adds file I/O to
  a code path that runs at most once per day. Not worth the complexity.
- **HMAC (HS256) signing** — Apple requires ES256. HMAC is not a substitute.

## Consequences

- Apple provider constructor returns an `error` (the only provider that
  does), since the `.p8` file is read at construction time.
- The provider does not currently verify the `id_token` RS256 signature
  against Apple's JWKS; verification is limited to `iss` + `aud` checks.
  A follow-up will add proper RS256 verification via the `github.com/lestrrat-go/jwx/v2`
  library. Tracked as a separate task.
