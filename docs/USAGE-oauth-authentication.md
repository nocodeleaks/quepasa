# OAuth/OIDC Authentication

QuePasa supports external authentication via **OAuth 2.0 / OpenID Connect** (OIDC).
Any OIDC-compliant identity provider works: Keycloak, Auth0, Google, Microsoft,
Okta, GitLab, etc.

## Configuration

Enable OAuth via environment variables:

```bash
OAUTH_ENABLED=true
OAUTH_PROVIDER_URL=https://identity.example.com
OAUTH_CLIENT_ID=quepasa-client
OAUTH_CLIENT_SECRET=<secret>
OAUTH_REDIRECT_URI=https://quepasa.example.com/oauth/callback
OAUTH_SCOPES=openid,email,profile
```

**Required fields:**
- `OAUTH_ENABLED` — set `true` to activate
- `OAUTH_PROVIDER_URL` — base URL of the OIDC provider
- `OAUTH_CLIENT_ID` — OAuth client ID registered with the provider
- `OAUTH_CLIENT_SECRET` — OAuth client secret
- `OAUTH_REDIRECT_URI` — callback URL (must match provider registration)

**Optional:**
- `OAUTH_SCOPES` — requested scopes (default: `openid,email,profile`)
- `QUEPASA_BASE_URL` — override base URL for callback construction (default: inferred from `WEBSERVER_HOST`/`WEBSERVER_PORT`)

## Flow

1. User navigates to `GET /oauth/login`
2. QuePasa redirects to provider's authorization endpoint
3. User authenticates with the provider
4. Provider redirects to `GET /oauth/callback?code=...`
5. QuePasa exchanges code for access token
6. QuePasa fetches user info (`/userinfo`) and extracts email
7. QuePasa creates or links the local user account (email becomes username)
8. QuePasa issues a **JWT** (same as form login) and sets it as a cookie
9. User is redirected to `/` (logged in)

## User Linking

When a user authenticates via OAuth:

- QuePasa looks up a local user by **email** (the email claim becomes the username).
- If the user exists, they are authenticated.
- If not, a new account is created **if `ACCOUNTSETUP` allows it** (env `ACCOUNTSETUP=true`).
- OAuth users are assigned a random password (never used; they authenticate via the provider).

## Downstream Behavior

After OAuth login, the user receives a **JWT** cookie identical to the one issued
by the form login flow. All downstream API routes, WebSocket connections, and
authorization logic treat OAuth-authenticated users exactly the same as
form-authenticated users.

## OIDC Discovery

QuePasa automatically discovers OIDC endpoints via the provider's
`.well-known/openid-configuration` document. No manual endpoint configuration needed.

## PKCE Support

QuePasa implements **PKCE** (Proof Key for Code Exchange, RFC 7636) automatically.
All authorization requests include `code_challenge` (S256 method) and the token
exchange includes `code_verifier`. This is required by most modern OIDC providers
(including Duende IdentityServer / Skoruba admin) and recommended for all OAuth flows.

## Example: Keycloak

```bash
OAUTH_ENABLED=true
OAUTH_PROVIDER_URL=https://keycloak.example.com/realms/master
OAUTH_CLIENT_ID=quepasa
OAUTH_CLIENT_SECRET=<your-secret>
OAUTH_REDIRECT_URI=https://quepasa.yourdomain.com/oauth/callback
```

Register QuePasa as an OAuth client in your OIDC provider with:
- **Redirect URI**: `https://quepasa.yourdomain.com/oauth/callback`
- **Grant type**: Authorization Code
- **Scopes**: `openid email profile`

## Security Notes

- OAuth routes (`/oauth/login`, `/oauth/callback`) are **public** (no auth middleware).
- State parameter is validated via HTTP-only cookie to prevent CSRF.
- QuePasa **never stores** the provider's access token; it is used only once to fetch user info.
- The local JWT has a 24-hour expiry (standard QuePasa session lifetime).

## Testing

To test the OAuth flow without a real provider, use a local Keycloak instance
or a public OIDC test provider like `https://oidc.example`.

## Relationship to Other Auth Modes

OAuth authentication **replaces** form login for the initial user authentication.
Once authenticated, the user receives a JWT and can use any of QuePasa's auth modes:

- **JWT** (cookie) — issued by OAuth callback
- **X-QUEPASA-USERKEY** — personal API key (rotatable via `/account/apikey`)
- **X-QUEPASA-TOKEN** — per-session token
- **X-QUEPASA-MASTERKEY** — admin gate (unchanged)

See `USAGE-authentication-modes.md` for details.
