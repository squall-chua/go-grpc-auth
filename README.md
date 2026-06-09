# go-grpc-auth

An authentication and authorization server built with Go, gRPC, and MongoDB. Serves both gRPC and REST (via grpc-gateway) on a single multiplexed port. Includes a built-in admin UI (Nuxt 4 SPA) embedded in the binary.

## Features

- **Authentication** - Register, login, logout, password change, token refresh
- **MFA** - TOTP (authenticator apps), Email OTP, SMS OTP (pluggable delivery)
- **Social Login** - Google, GitHub (extensible)
- **OAuth2/OIDC Provider** - Authorization code flow with PKCE, client credentials, discovery, JWKS, userinfo
- **RBAC** - Role and permission-based access control with proto-level annotations
- **Multi-Tenancy** - Namespace isolation with per-tenant security policies (password policy, MFA, IP allowlist/denylist)
- **Admin API** - User, role, permission, namespace, OIDC client management
- **Audit Logging** - Security event tracking with filters (event type, user, time range)
- **Embedded Web UI** - Nuxt 4 SPA admin dashboard embedded in the Go binary
- **Single Port** - gRPC + HTTP/REST + SPA served on one port via cmux

## Quick Start

### Docker

```bash
docker-compose up -d
```

This starts the auth server, MongoDB, and Redis.

### Manual

```bash
# Start MongoDB and Redis
docker-compose up -d mongodb redis

# Generate RSA keys
scripts/gen-keys.sh

# Run database migrations (requires migrate-mongo)
scripts/migrate.sh up

# Build and run
go build -o auth-server ./cmd/server/main.go
./auth-server
```

The server starts on port **8080** by default.

### Default Credentials

After running migrations, a superadmin account is available:

| Field | Value |
|-------|-------|
| Username | `superadmin` |
| Email | `admin@localhost` |
| Password | `Admin@123` |
| Namespace | `default` |

## Configuration

All settings are configurable via flags or environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port (gRPC + HTTP) |
| `MONGO_URI` | `mongodb://localhost:27017/auth_db` | MongoDB connection string |
| `REDIS_URI` | `localhost:6379` | Redis URI (optional, falls back to in-memory) |
| `RSA_PRIVATE_KEY` | `keys/id_rs256` | RSA private key path |
| `RSA_PUBLIC_KEY` | `keys/id_rs256.pub` | RSA public key path |
| `ISSUER` | `https://auth.example.com` | Token issuer URL |
| `MFA_EMAIL_ENABLED` | `false` | Enable email OTP delivery |
| `MFA_SMS_ENABLED` | `false` | Enable SMS OTP delivery |
| `ACCESS_TOKEN_DURATION` | `15m` | Access token lifetime |
| `REFRESH_TOKEN_DURATION` | `168h` | Refresh token lifetime (7 days) |
| `RATE_LIMIT_REQUESTS` | `5` | Max requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Rate limit window |

Social login providers (optional):

| Variable | Description |
|----------|-------------|
| `GOOGLE_CLIENT_ID` | Google OAuth2 Client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth2 Client Secret |
| `GITHUB_CLIENT_ID` | GitHub OAuth2 Client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth2 Client Secret |

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/v1/auth/register` | Register new user | Public |
| POST | `/v1/auth/login` | Login | Public |
| POST | `/v1/auth/refresh` | Refresh token pair | Public |
| POST | `/v1/auth/logout` | Logout (revoke tokens) | Bearer |
| POST | `/v1/auth/password/change` | Change password | Bearer |
| POST | `/v1/auth/validate` | Validate token | Public |
| POST | `/v1/auth/mfa/initiate` | Start MFA setup | Bearer |
| POST | `/v1/auth/mfa/verify` | Verify MFA code | Public |
| GET | `/v1/auth/social/{provider}/url` | Get social login URL | Public |
| GET | `/v1/auth/social/{provider}/callback` | Social login callback | Public |

### OAuth2/OIDC

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/.well-known/openid-configuration` | OIDC Discovery |
| GET | `/.well-known/jwks.json` | JSON Web Key Set |
| GET | `/oauth2/authorize` | Authorization endpoint |
| POST | `/oauth2/token` | Token endpoint |
| GET | `/oauth2/userinfo` | UserInfo endpoint |
| GET | `/oauth2/logout` | Logout endpoint |

### Admin - Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/admin/users` | List users (search, filter by status) |
| GET | `/v1/admin/users/{id}` | Get user details |
| PATCH | `/v1/admin/users/{id}/status` | Update user status |
| POST | `/v1/admin/users/{id}/reset-password` | Reset user password |
| POST | `/v1/admin/users/{id}/roles` | Grant roles |
| DELETE | `/v1/admin/users/{id}/roles` | Revoke roles |
| POST | `/v1/admin/users/{id}/permissions` | Grant permissions |
| DELETE | `/v1/admin/users/{id}/permissions` | Revoke permissions |

### Admin - Roles & Permissions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/admin/roles` | Create role |
| GET | `/v1/admin/roles` | List roles |
| DELETE | `/v1/admin/roles/{id}` | Delete role |
| POST | `/v1/admin/permissions` | Create permission |
| GET | `/v1/admin/permissions` | List permissions |
| DELETE | `/v1/admin/permissions/{id}` | Delete permission |

### Admin - Namespaces

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/admin/namespaces` | Create namespace |
| GET | `/v1/admin/namespaces` | List namespaces |
| GET | `/v1/admin/namespaces/{id}` | Get namespace |
| PATCH | `/v1/admin/namespaces/{id}/config` | Update namespace config |
| DELETE | `/v1/admin/namespaces/{id}` | Delete namespace |

### Admin - OIDC Clients

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/admin/oidc/clients` | Register client |
| GET | `/v1/admin/oidc/clients` | List clients |
| GET | `/v1/admin/oidc/clients/{id}` | Get client |
| PATCH | `/v1/admin/oidc/clients/{id}` | Update client |
| DELETE | `/v1/admin/oidc/clients/{id}` | Delete client |
| POST | `/v1/admin/oidc/clients/{id}/rotate-secret` | Rotate client secret |

### Admin - Audit

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/admin/audit` | List audit logs (filter by event, user, time range) |

## Web3 / SIWE Sign-In

Users can sign in with an Ethereum (or other EVM) wallet via SIWE (EIP-4361).
The server issues a nonce, the wallet signs a SIWE message, the server
verifies the signature, and on success issues the standard
access/refresh/ID token pair.

Supported chains are configured per-namespace via
`NamespaceConfig.AllowedWeb3ChainIds`. The default (when unset) is
`[1, 8453, 42161, 10, 137]` (Ethereum, Base, Arbitrum, Optimism, Polygon).

v1 supports EOA wallets only. Smart-contract wallets (Safe, Argent) return
a clear error.

## Architecture

```
Client --> cmux (PORT)
  |-- gRPC (HTTP/2 + application/grpc) --> gRPC Server --> handlers
  |-- HTTP --> grpc-gateway mux --> gRPC handlers
                |-- /v1/*          (REST API)
                |-- /oauth2/*      (OIDC endpoints)
                |-- /.well-known/* (Discovery, JWKS)
                |-- /              (Embedded Nuxt SPA)
```

**Token model:** Opaque access/refresh tokens (SHA-256 hashed, stored in MongoDB). ID tokens are RS256 JWTs. Token validation results are cached 30 seconds in-process.

**RBAC:** Proto methods are annotated with `options.v1.rule` specifying required roles/permissions. The gRPC interceptor validates tokens and enforces access (role match OR permission match).

**Automatic role/permission registration:** Services that authenticate via OIDC client credentials can automatically register their declared roles and permissions with the auth server on startup. The `authclient.RegisterServiceRolesAndPermissions()` helper uses protobuf reflection to extract all roles and permissions from `options.v1.rule` annotations and creates them via the Admin API (idempotent — duplicates are ignored). This ensures roles and permissions are visible in the admin UI for granting to users without manual setup.

```go
// In your service's main.go, after authenticating with the auth server:
authClient, err := authclient.NewClient(authHost, provider,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
serviceDesc := myservice.File_my_proto.Services().ByName("MyService")
err = authClient.RegisterServiceRolesAndPermissions(ctx, "default", "My service", serviceDesc)
```

## Project Structure

```
cmd/server/         Entry point, manual DI wiring
internal/
  config/           Flags + env var configuration
  domain/           Domain entities
  repository/       MongoDB data access (gmqb query builder)
  service/          Business logic (auth, token, oidc, admin, audit, webhook, ratelimit)
  server/           gRPC handlers, HTTP gateway, interceptors, embedded UI
  keys/             RSA key loading and JWKS
api/
  proto/            Protobuf definitions (source of truth)
  v1/               Generated Go code (do not edit)
  swagger/          Generated OpenAPI specs
pkg/
  interceptor/      Exported interceptors for consuming services
  authclient/       Exported typed gRPC client wrapper
web/                Nuxt 4 SPA (embedded in binary)
migrations/         MongoDB migrations (migrate-mongo)
scripts/            Build and setup scripts
examples/greeter/   Example gRPC service + client using exported interceptors
```

## Examples

### Greeter Service

A minimal gRPC service demonstrating the exported auth interceptors in `pkg/interceptor/`. The server authenticates itself with the auth server, automatically registers its declared roles and permissions, and validates incoming tokens. The client logs in and auto-injects tokens into every call.

**4 RPCs at different auth levels:**

| RPC | Auth Rule |
|-----|-----------|
| `SayHello` | Public |
| `SayHelloAuthenticated` | Any valid token |
| `SayHelloAdmin` | `admin` role required |
| `SayHelloEditor` | `greet:write` permission required |

**Running:**

```bash
# 1. Start auth server
go run ./cmd/server/main.go

# 2. Register an OIDC client for the greeter service and a test user
# (use the superadmin account — see Default Credentials above)
# Register OIDC client via Admin API, then register a test user:
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"testuser@example.com","username":"testuser","password":"testpass"}'

# 3. Start greeter server (authenticates via OIDC client credentials)
go run ./examples/greeter/server/main.go \
  -auth grpc://<client_id>:<client_secret>@localhost:8080

# 4. Run client
go run ./examples/greeter/client/main.go \
  -auth grpc://testuser:testpass@localhost:8080 \
  -greeter grpc://localhost:9090
```

## Development

```bash
# Generate protobuf code (requires buf)
scripts/gen-proto.sh

# Build web UI
scripts/build-web.sh

# Run database migrations
scripts/migrate.sh up      # apply
scripts/migrate.sh status  # check
scripts/migrate.sh down    # rollback

# Generate RSA keys
scripts/gen-keys.sh

# Build
go build -o auth-server ./cmd/server/main.go
```

## Tech Stack

**Backend:** Go 1.25, gRPC, grpc-gateway, MongoDB, Redis (optional), cmux, bcrypt, RS256 JWT, zap

**Frontend:** Nuxt 4, Vue 3, Pinia, @nuxt/ui, Tailwind CSS

**Tooling:** buf (proto), migrate-mongo (migrations)
