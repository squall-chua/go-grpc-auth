# Ubiquitous Language

Canonical vocabulary for the go-grpc-auth server. New terms land here as they are
resolved during design sessions. Implementation details, code paths, and trade-off
rationale belong in `docs/adr/`, not here.

## Identity & Authentication

- **User** — A person (or service principal) that authenticates against the
  server. The only entity that holds roles and permissions. Identified by
  `id` (BSON ObjectID).

- **Namespace** — A multi-tenant isolation boundary. Every `User` belongs to
  exactly one `Namespace`. Most queries are namespace-scoped.

- **Principal** — The runtime, validated representation of a `User` for a
  single request. Contains `user_id`, `namespace`, `roles`, `permissions`, and
  an expiry. Returned by the `ValidateToken` RPC.

- **Token** — An opaque access credential (32-byte random hex), SHA-256 hashed
  at rest. Never a JWT. Looked up by hash. `id_token` is the exception — it is
  an RS256 JWT per OIDC.

## Social Login

- **SocialProvider** — An external identity service a `User` can use to sign
  in. Current canonical set: `google`, `github`, `facebook`, `twitter`, `apple`.
  Lowercase string. The same string used in URL paths
  (`/v1/auth/social/{provider}/callback`).

- **Provider key** — Synonym for `SocialProvider`. The lowercase canonical
  identifier. Do not use the marketing name (`X` instead of `twitter`); the
  API key stays `twitter` for URL stability and consistency with `github`.

- **SocialUser** — The in-memory representation of a user fetched from a
  `SocialProvider` during a single callback. Never persisted. Fields: `id`,
  `email`, `name`, `avatar_url`.

- **SocialIdentity** — A persisted linkage record on a `User` recording that
  the user has signed in via a specific `SocialProvider`. Fields: `provider`,
  `external_id`, `email`, `name`, `avatar_url`, `email_verified`. One `User`
  may have zero or many `SocialIdentity` rows.

- **Account linkage** — The behavior applied when a callback returns an
  email that matches an existing `User` in the same `Namespace`. The
  `SocialIdentity` is appended to the existing `User`. Not the same as
  authentication; the user is now able to sign in via the new provider too.
  **Exception:** when the incoming `email_verified` is `false` (Apple relay
  emails), the email is not used as a lookup key and a new `User` is
  auto-provisioned instead.

- **First-login capture (Apple-specific)** — The discipline of persisting
  `name` and `email` from a `SocialUser` into the `SocialIdentity` record on
  the *first* successful callback from that provider. Required for Apple
  because Apple's authorization response includes `name` and `email` only on
  the user's first authorization; subsequent logins return only the opaque
  `sub`. For all other providers, the same data is also re-fetched on every
  callback.

- **Apple relay email** — A private relay address
  (`*@privaterelay.appleid.com`) Apple generates when a user enables "Hide My
  Email". Different per app. The server treats relay addresses as
  `email_verified=false` and never uses them as the key for `Account linkage`.

- **PKCE** — Proof Key for Code Exchange (RFC 7636). Supported by the
  underlying `golang.org/x/oauth2` library for any OAuth 2.0 provider. Not
  currently enabled by any of our outbound social-login flows — the
  server acts as a confidential client in every case. We do not enforce
  PKCE on inbound client requests to our own OAuth endpoints.

- **Provider registration** — The startup-time process that adds a
  `SocialProvider` to the in-memory provider map. A provider is registered
  only if its required client credentials are present in config. Missing
  config → provider silently absent. Partial config + unreadable resource
  (e.g. Apple's `.p8` file) → startup fails loudly.

## Authorization

- **Role** — A named bundle of `Permission`s. Granted to a `User` within a
  `Namespace`. Roles are namespace-scoped.

- **Permission** — An atomic capability (e.g. `users:read`). Assigned to
  `Role`s. Enforced by the gRPC interceptor reading
  `options.v1.rule` annotations on RPC methods.

- **MFA method** — A second factor enrolled by a `User`: `totp`, `email`,
  or `sms`. Independently toggleable per `Namespace`.

## OAuth2 / OIDC Provider

- **OIDC client** — A third-party application registered to use this server
  as its identity provider. Identified by `client_id`; has a hashed
  `client_secret`. Namespace-scoped.

- **Auth code** — Short-lived one-time code issued by the `/oauth2/authorize`
  endpoint and exchanged at `/oauth2/token`. Stored at rest until consumed
  or expired.

- **Consent** — A persisted record that a `User` has approved a specific
  `OIDC client` to access a specific scope set. Required before the server
  issues tokens on behalf of that client to that user.
