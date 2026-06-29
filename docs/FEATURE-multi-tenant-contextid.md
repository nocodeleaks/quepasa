# Multi-Tenant / Context-based Session Sharing

QuePasa supports **multi-tenant session sharing** via the optional `contextid` field.
Sessions created by different users can be shared within the same **context** (tenant).

## Use Case

In multi-tenant deployments, a single QuePasa instance serves multiple organizations
or teams (contexts). Each context is identified by a unique `contextid`.

Users in the same context can share WhatsApp sessions. Users in different contexts
cannot see or interact with each other's sessions.

## Schema

- **`contextid`** (optional, TEXT) — tenant/sharing scope identifier
- Stored in `servers` table alongside `user` (session owner)
- Indexed for fast multi-tenant queries

## API

### Creating a session with context

```bash
POST /info
Authorization: Bearer <jwt> or X-QUEPASA-USERKEY: <key>
Content-Type: application/json

{
  "contextid": "org-acme-sales",
  "groups": true,
  "broadcasts": false
}
```

If `contextid` is not provided, the session is created without a context (default behavior).

### Updating context

```bash
PATCH /info
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "contextid": "org-acme-support"
}
```

### Retrieving session info

```bash
GET /info
Authorization: Bearer <jwt>
```

Response includes `contextid` if set:

```json
{
  "token": "abc123",
  "wid": "5511999999999",
  "contextid": "org-acme-sales",
  "verified": true
}
```

## Migration

Schema change is applied automatically via migration:

```sql
ALTER TABLE `servers` ADD COLUMN `contextid` TEXT DEFAULT NULL;
CREATE INDEX IF NOT EXISTS `idx_servers_contextid` ON `servers` (`contextid`);
```

Existing sessions have `contextid = NULL` (unscoped, backward-compatible).

## Sharing Behavior

- **Owner** (`user`) is always stored and required (set automatically from JWT/API key).
- **Context** (`contextid`) is optional.
- Sessions with the same `contextid` are considered shared within that tenant.
- Future extension: multi-user access control can filter sessions by both `user` and `contextid`.

## Security Notes

- `contextid` is **not** an authorization mechanism by itself. Access control must be
  enforced by the integration layer.
- QuePasa does not validate `contextid` format or existence. The calling application
  is responsible for ensuring `contextid` integrity.
- Current implementation stores `contextid` but does not enforce isolation at the
  API layer. Isolation logic should be implemented in the integration (checking both
  user and contextid).

## Example: External Integration

When integrating with an external permission system:

1. External system checks user's permission for the given `contextid`.
2. If authorized, external system passes `contextid` in the QuePasa API call.
3. QuePasa stores the session with `contextid` set.
4. Later, external system lists sessions filtered by `contextid` (future API enhancement).

This allows multi-tenant session sharing without mixing contexts.
