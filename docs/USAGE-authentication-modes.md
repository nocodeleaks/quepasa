# Authentication Modes

## Objective

Explain the four authentication modes available in the QuePasa API and how they interact:

- `X-QUEPASA-TOKEN`
- `X-QUEPASA-MASTERKEY`
- `JWT` (`Authorization: Bearer ...`)
- Anonymous access

This document describes current behavior for canonical routes under `/api` and `/api/v5`.

## Quick Summary

- `JWT` authenticates a user identity and allows access to all sessions owned by that user.
- `X-QUEPASA-TOKEN` authenticates a single session scope and limits access to that session.
- `X-QUEPASA-MASTERKEY` is a privileged gate for sensitive operations; it is not a user login identity by itself.
- Anonymous access is allowed only on selected public/system endpoints.

## 1) JWT Mode (User Scope)

Use when you want user-level access across all sessions owned by that user.

Header:

```http
Authorization: Bearer <jwt>
```

Behavior:

- User identity is read from `user_id` JWT claim.
- Protected routes are authorized in the user scope.
- Session routes can manage any session owned by that JWT user.

Typical endpoints:

- `GET /api/auth/session`
- `GET /api/users`
- `GET /api/sessions`
- `POST /api/sessions`

## 2) X-QUEPASA-TOKEN Mode (Single Session Scope)

Use when you want a token-bound integration that only operates on one session.

Header:

```http
X-QUEPASA-TOKEN: <session-token>
```

Behavior:

- API resolves the owner user from the session token.
- Access is forced to that authenticated session scope.
- Session list/search and session-targeted operations are constrained to that token.
- If a request attempts a different session token in path/query/body, effective scope remains the authenticated token.

Important:

- This mode is a valid authentication mode for protected SPA/canonical routes.
- In `POST /api/sessions`, this same header can also be used as the desired token for the new session only when `RELAXED_SESSIONS=true` and request identity is valid.

## 3) X-QUEPASA-MASTERKEY Mode (Privileged Gate)

Use when endpoint policy requires elevated authorization.

Header:

```http
X-QUEPASA-MASTERKEY: <master-key>
```

Behavior:

- Grants elevated permission checks on endpoints that require master access.
- Does not replace identity authentication for protected user/session routes.
- Works as an additional privilege condition depending on endpoint policy.

Example policy:

- `POST /api/sessions`:
  - `RELAXED_SESSIONS=true`: user/session identity can create without master key.
  - `RELAXED_SESSIONS=false`: master key is additionally required.

## 4) Anonymous Mode

Use for bootstrap/public visibility endpoints only.

No authentication headers required.

Typical anonymous endpoints:

- `GET /api/auth/config`
- `POST /api/auth/login`
- `POST /api/users` (public user creation, subject to server config)
- `GET /api/system/health`
- `GET /api/system/version`
- `GET /api/system/environment` (preview payload when no master key)

## Precedence and Composition

When more than one auth header is present:

- JWT provides user identity scope.
- `X-QUEPASA-TOKEN` provides session scope.
- `X-QUEPASA-MASTERKEY` provides privileged gate checks where required.

Recommended operational model:

- Use JWT for dashboards/admin flows.
- Use `X-QUEPASA-TOKEN` for automation/integration tied to one session.
- Add `X-QUEPASA-MASTERKEY` only for privileged endpoints or strict policies.

## cURL Examples

### JWT user scope

```bash
curl -X GET "http://localhost:3100/api/sessions" \
  -H "Authorization: Bearer <jwt>"
```

### Session token scope

```bash
curl -X GET "http://localhost:3100/api/sessions" \
  -H "X-QUEPASA-TOKEN: <session-token>"
```

### Strict create (master key required)

```bash
curl -X POST "http://localhost:3100/api/sessions" \
  -H "Authorization: Bearer <jwt>" \
  -H "X-QUEPASA-MASTERKEY: <master-key>" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Anonymous health check

```bash
curl -X GET "http://localhost:3100/api/system/health"
```

## Notes

- Keep `X-QUEPASA-TOKEN` secret in integrations; it is a credential in session-scope mode.
- For browser apps, prefer JWT mode for better account-level control.
- For headless bots, prefer `X-QUEPASA-TOKEN` to enforce least-privilege per session.
