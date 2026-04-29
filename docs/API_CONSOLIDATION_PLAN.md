# API Consolidation Plan

This document defines the target structure for the QuePasa HTTP API after consolidating legacy routes and recently imported SPA-oriented endpoints.

## Goal

QuePasa should expose one canonical API surface that is suitable for:

- browser frontends
- mobile frontends
- admin dashboards
- backend integrations
- CLI tools
- automation scripts

The API must not be split into "frontend API" and "normal API".

There may still be temporary compatibility aliases during migration, but there should be one canonical route model and one canonical resource vocabulary.

## Compatibility Commitment

All existing API functionality must be preserved during the consolidation.

This means:

- existing routes must keep working during migration
- no currently available feature should be dropped just because it came from the old SPA extraction
- the consolidation effort is additive first, subtractive later
- route removal only happens after the canonical replacement is complete and validated

The goal is to improve structure without losing behavior.

## Core Decision

The API should be organized by **resource families**, not by historical UI origin and not by action-style legacy paths.

Examples of valid top-level families:

- `users`
- `sessions`
- `dispatches`
- `messages`
- `groups`
- `contacts`
- `labels`
- `media`
- `system`

Examples of route styles to reduce over time:

- `/sendtext`
- `/senddocument`
- `/server/create`
- `/command`
- `/paircode`
- `/rabbitmq`
- `/webhook`

These routes reflect implementation history, not the domain model.

## Naming Decision: Replace `server` With `session`

The current name `server` is overloaded and misleading.

In QuePasa, the object currently called `server` is usually not:

- a physical server
- an HTTP server
- a network host
- an infrastructure node

What it actually represents is much closer to:

- a WhatsApp session
- a managed connection lifecycle
- a tokenized account/runtime binding
- a unit that can be created, paired, connected, disconnected, enabled, disabled, and deleted

For API vocabulary, the recommended canonical resource name is:

## `session`

Reasons:

- it matches QR/pair-code lifecycle semantics
- it matches connection state semantics
- it is easier for frontend developers to understand
- it avoids confusion with the actual web server and backend process
- it scales better when the same backend manages many WhatsApp identities

Examples:

- `GET /v4/sessions`
- `POST /v4/sessions`
- `GET /v4/sessions/{token}`
- `PATCH /v4/sessions/{token}`
- `DELETE /v4/sessions/{token}`

## Canonical API Model

The current `v4` structure must remain stable during this migration.

The new canonical structure should be introduced in the next version only.

Version policy:

- the current API version must not be restructured in-place
- each new major API structure should be introduced in the next version
- the unversioned `/api` route must always point to the newest canonical API version
- the explicit versioned route must also exist for that newest version

For this migration, that means:

- current legacy routes keep working as they already do
- the new canonical family-based model is introduced under `/api/v5`
- `/api` must point to the same route model as `/api/v5`
- legacy endpoints may also remain mounted for a transition period

Examples for this migration target:

- `/api/users`
- `/api/sessions`
- `/api/sessions/dispatches`
- `/api/messages`
- `/api/groups`
- `/api/v5/users`
- `/api/v5/sessions`
- `/api/v5/sessions/dispatches`
- `/api/v5/messages`
- `/api/v5/groups`

The `/spa` surface should be eliminated over time.

It should not remain as a permanent second-class namespace.

Everything useful currently exposed under `/spa` must be remounted into the canonical family-based API and then `/spa` should be retired.

During migration, `/spa` may remain temporarily as a compatibility alias only.

## Namespace Decision for the New Model

The recommended namespace for the canonical new model is:

## `/api/v5`

Reasons:

- it keeps the current and legacy route set intact during migration
- it makes the new family-based model explicit from day one
- it avoids mixing legacy and canonical route styles under the same root without distinction
- it allows frontend and integration clients to migrate endpoint-by-endpoint
- it reduces risk while the v5 contract is still stabilizing

Therefore:

- existing legacy routes continue to work as they do today
- new canonical family routes are introduced under `/api/v5`
- `/api` must expose the same canonical route set as `/api/v5`
- `/spa` is migrated into `/api/v5`
- once the migration is complete, `/spa` should be removed

Legacy endpoints may remain mounted alongside the new model for a while, but they are compatibility endpoints, not the canonical design target.

## Identifier Transport Policy

The canonical new API should avoid path parameters for identifiers whenever possible.

Reasons:

- QuePasa deals with identifiers that are easy to corrupt when encoded in paths
- values such as tokens, group ids, chat ids, phone-derived identifiers, filenames, and URLs are safer outside the URL path
- body and header transport reduces ambiguity around escaping and normalization

Preferred order:

1. request body for identifiers and operation payloads
2. request headers for stable operation context when the same identifier scopes many calls
3. query string only when extremely necessary
4. path parameters only for simple, safe, opaque identifiers when there is a strong reason

Recommended usage:

- use body for create, update, delete, command, and search operations
- use headers for session scope when that improves client ergonomics and avoids repeated path encoding
- use query string for pagination, filtering, and optional selectors only
- avoid putting WhatsApp-specific identifiers in path segments unless there is no better option
- when a `GET` operation needs identifiers that would otherwise become unsafe path parameters, allow the compatible transport of those identifiers in headers

Examples:

- `POST /api/v5/sessions/get` with token in body
- `POST /api/v5/messages/get` with session token and message id in body
- `POST /api/v5/groups/get` with session token and group id in body
- `GET /api/v5/messages?page=1&limit=50` with session scope in header when applicable
- `GET /api/v5/media/messages` with compatible identifiers in headers when the operation must remain a `GET`
- `POST /api/v5/media/messages/get` with the same identifiers in body for clients that prefer explicit request payloads

This is intentionally more transport-safe than a strict REST path-parameter style.

## Code Organization Decision

The code should be separated by implementation generation, even while the runtime surface remains compatible.

Recommended structure:

- `src/api/legacy/` for current route handlers and compatibility wiring
- `src/api/v5/` for the new canonical family-based model
- `src/api/models/` for shared API DTOs only when they are truly version-neutral

Important:

- this is a code organization decision, not necessarily a URL decision
- moving files to `src/api/legacy/` does not mean exposing URLs under `/api/legacy`
- the legacy routes should keep their existing URLs while the code is being reorganized
- the canonical new routes should be exposed under `/api/v5`

This keeps runtime compatibility while making the codebase easier to evolve.

## Target Resource Families

## 1. `system`

Responsibility:

- health
- version
- runtime/environment preview
- operational diagnosis that is not user-owned

Examples:

- `GET /api/v5/system/health`
- `GET /api/v5/system/version`
- `GET /api/v5/system/environment`

## 2. `auth`

Responsibility:

- login bootstrap config
- current authenticated session info when authentication is enabled
- authenticated account summary for the current user
- master key status visibility only, without ever exposing the secret value

Examples:

- `GET /api/v5/auth/config`
- `GET /api/v5/auth/session`
- `GET /api/v5/auth/account`
- `GET /api/v5/auth/masterkey/status`

The master key must never be returned in API responses.

When this capability is relevant for the current user, the API may only expose whether a master key is configured, available for use, or unavailable.

If auth routes must remain shared with form/browser login behavior, they should still expose canonical API semantics.

## 3. `users`

Responsibility:

- user creation
- user listing
- user deletion
- user administration
- public bootstrap user creation when account setup is allowed

Examples:

- `GET /api/v5/users`
- `POST /api/v5/users`
- `POST /api/v5/users/get`
- `DELETE /api/v5/users`

The same family should support both authenticated administrative user creation and public bootstrap setup creation when the environment explicitly allows that flow.

## 4. `sessions`

Responsibility:

- create and manage WhatsApp runtime units
- read lifecycle and configuration state
- update user-owned options
- delete one unit
- apply generic option changes through one normalized session option contract

Examples:

- `GET /api/v5/sessions`
- `POST /api/v5/sessions`
- `POST /api/v5/sessions/search`
- `POST /api/v5/sessions/get`
- `PATCH /api/v5/sessions`
- `DELETE /api/v5/sessions`
- `POST /api/v5/session/option`

### Session Subfamilies

#### `connection`

Responsibility:

- QR code
- pair code
- enable/disable
- connect/disconnect state transitions

Examples:

- `GET /api/v5/session/qrcode`
- `GET /api/v5/session/paircode`
- `POST /api/v5/session/enable`
- `POST /api/v5/session/disable`

The `connection` label is descriptive only.

It should not appear in the URL when it adds no domain value.

For QuePasa, QR code, pair code, enable, and disable are direct session lifecycle operations, so the simpler `session/...` path is preferred.

#### `settings`

Responsibility:

- toggles and options that are conceptually configuration, not commands

Examples:

- `PATCH /api/v5/session/settings`
- `POST /api/v5/session/debug`

Session token should travel in header or body, not in the path.

Long term, even toggle endpoints should preferably become patch operations.

## 5. `dispatches`

Responsibility:

- external delivery configuration owned by one session
- webhooks
- rabbitmq
- future external sinks

Examples:

- `GET /api/v5/dispatches`
- `POST /api/v5/dispatches`
- `PATCH /api/v5/dispatches`
- `DELETE /api/v5/dispatches`

If keeping type-specific routes is operationally useful during transition:

- `GET /api/v5/dispatches/webhooks`
- `POST /api/v5/dispatches/webhooks`
- `GET /api/v5/dispatches/rabbitmq`
- `POST /api/v5/dispatches/rabbitmq`

But the long-term model should prefer one unified dispatch resource with `type`.

## 6. `contacts`

Responsibility:

- contact listing
- contact resolution
- availability checks
- LID/phone helper queries when still needed

Examples:

- `GET /api/v5/contacts`
- `POST /api/v5/contacts/search`
- `POST /api/v5/contacts/availability`
- `POST /api/v5/contacts/get`

Phone/LID compatibility routes should eventually be represented as contact lookup operations rather than standalone utility nouns.

## 7. `messages`

Responsibility:

- send
- list
- edit
- revoke
- receive/history access
- retry when still needed operationally

Examples:

- `GET /api/v5/messages`
- `POST /api/v5/messages`
- `POST /api/v5/messages/get`
- `PATCH /api/v5/messages`
- `DELETE /api/v5/messages`
- `POST /api/v5/messages/retry`

Message send variants should be normalized into one send contract instead of multiple path names like `senddocument`, `sendbinary`, `sendencoded`, and `sendurl`.

The request body should define send mode, attachments, and content type.

## 8. `chats`

Responsibility:

- archive/unarchive
- mark read/unread
- presence/typing
- chat labels

Examples:

- `POST /api/v5/chats/archive`
- `POST /api/v5/chats/unarchive`
- `POST /api/v5/chats/read`
- `POST /api/v5/chats/unread`
- `POST /api/v5/chats/presence`
- `POST /api/v5/chats/labels/get`
- `POST /api/v5/chats/labels`
- `DELETE /api/v5/chats/labels`

## 9. `groups`

Responsibility:

- list groups
- read a group
- create a group
- leave a group
- manage group settings and participants
- update group name and description/topic through the same family

Examples:

- `GET /api/v5/groups`
- `POST /api/v5/groups`
- `POST /api/v5/groups/get`
- `POST /api/v5/groups/leave`
- `PATCH /api/v5/groups`
- `PUT /api/v5/groups/participants`
- `PUT /api/v5/groups/photo`
- `POST /api/v5/groups/invite`

The `PATCH /api/v5/groups` contract should be capable of expressing name and description/topic updates so that current SPA group mutation capabilities remain covered.

## 10. `media`

Responsibility:

- message download
- picture info/data
- history export/download
- direct message media retrieval compatible with current SPA download behavior

Examples:

- `GET /api/v5/media/messages`
- `POST /api/v5/media/messages/get`
- `POST /api/v5/media/pictures/get`
- `POST /api/v5/media/pictures/info`
- `POST /api/v5/media/download`

The media family must cover both current SPA-style direct message download and the newer normalized download contract.

## 11. `labels`

Responsibility:

- conversation label definitions
- label assignment and removal
- label search with assigned chat information when relevant

Examples:

- `GET /api/v5/labels`
- `POST /api/v5/labels`
- `PATCH /api/v5/labels`
- `DELETE /api/v5/labels`
- `POST /api/v5/labels/search`

For label-oriented queries, the preferred model is to search labels and include
assigned chats inside the response payload when needed, instead of exposing a
separate label-to-chat lookup route just for that relation.

## Current `/spa` Functional Coverage Requirement

The v5 plan must preserve all functional capabilities currently exposed under `/spa`, even when the final route names and grouping become clearer.

This section is intentionally capability-oriented rather than route-by-route.

Required coverage extracted from the current `/spa` surface:

### `auth`

- login config
- authenticated session read
- authenticated account read
- authenticated master key status read when enabled

### `users`

- list users
- delete user
- bootstrap/public user creation when account setup is enabled

### `sessions`

- list sessions
- search sessions
- create session
- get session details
- update session
- delete session
- get QR code
- get pair code
- enable session
- disable session
- set debug state
- change generic session option

### `dispatches`

- list webhooks for a session
- add webhook for a session
- remove webhook for a session
- list rabbitmq configs for a session
- add rabbitmq config for a session
- remove rabbitmq config for a session

### `contacts`

- list contacts for a session

### `groups`

- list groups for a session
- create group
- get one group
- leave group
- update group name
- update group description/topic
- update participants
- update photo
- get invite

### `chats`

- list chat labels
- assign label to chat
- remove label from chat
- archive chat
- send presence/typing update

### `messages`

- send message
- list messages
- edit message
- delete/revoke message
- request retry when applicable
- request history download

### `media`

- direct media download by message reference
- picture info retrieval
- normalized download request

### `labels`

- list labels
- create label
- update label
- delete label
- search labels including assigned chats when requested

If a future v5 family proposal cannot represent one of the items above, then the proposal is incomplete and must be revised before `/spa` can be retired.

## Event Emission and Metrics Capture

The canonical API and its backend flows must adopt an event-oriented operational model.

Every relevant situation should emit an internal event that can be consumed asynchronously by metrics and other observers.

This event flow must be non-blocking for the primary operation.

Examples of situations that should emit events:

- request accepted
- request rejected by validation
- authentication success or failure
- session lifecycle changes
- dispatch delivery success or failure
- webhook blocked/unblocked state changes
- message send, edit, revoke, retry, and download operations
- group mutation operations
- label assignment and removal
- media retrieval operations

### Event Rules

1. event emission must not block the HTTP response path or the core business flow
2. event emission should be fire-and-forget from the producer point of view
3. failure to enqueue or consume an event must not break the primary user operation
4. event payloads should carry operational facts, not transport-specific response formatting
5. event names and payload fields should follow domain semantics, not controller-specific naming

### Ownership Rules

Event producers should live close to the domain or application flow that owns the situation.

Examples:

- API controllers may emit request-contract events when that is the clearest ownership boundary
- runtime handlers should emit business-operation events
- dispatch-related flows should emit outbound delivery events
- session lifecycle flows should emit connection and state events

The central metrics module must remain generic.

Metrics must not become the place where business events are defined.

Instead:

1. producers emit domain or operational events
2. metrics consumers subscribe to those events asynchronously
3. metrics recording uses the existing generic recorder factories

### Metrics Integration Rules

Metrics recording should happen in asynchronous event consumers, not inline in a way that couples every business path to metrics concerns.

This preserves the current design where the metrics module exposes generic factories only.

Required implications:

1. metrics consumers may translate one event into counters, histograms, or labeled metrics
2. metrics consumers must preserve existing metric names and labels unless a migration explicitly changes them
3. disabling metrics must not require producers to change behavior
4. event producers must not need conditional logic for metrics enabled or disabled state

### Delivery Model

The initial implementation may use an in-process asynchronous event bus.

That bus should be sufficient if it satisfies these conditions:

1. publish is non-blocking for producers
2. slow consumers do not stall request processing
3. consumer failures are isolated
4. the design can later evolve to external sinks without rewriting producer semantics

### Testing Requirement for Events

Whenever a migration slice introduces event emission, tests should validate at least:

1. the relevant event is emitted for the intended situation
2. the primary operation still completes when event consumers are absent or slow
3. event handling failures do not break the main operation
4. metrics consumption remains an observer concern, not a business prerequisite

## Current Structural Problems

The current codebase already contains most required functionality, but the route structure still reflects legacy growth.

Key problems:

1. flat action-based routes dominate the legacy API
2. the same domain capability exists under unrelated path styles
3. SPA-imported routes are grouped by extraction history, not by domain family
4. `server` is overloaded and semantically weak
5. send/media endpoints are fragmented by transport details instead of resource semantics
6. dispatch is partly represented as separate webhook/rabbitmq concepts instead of one family

## Recommended Route Policy

Use these rules for new canonical endpoints:

1. prefer nouns for families and resources
2. prefer HTTP method semantics over action names in the path
3. nest resources only when ownership is clear
4. keep compatibility aliases temporary and explicit
5. do not introduce frontend-only path conventions for frontend needs that belong to the canonical API
6. prefer one normalized request contract over many transport-specific route names
7. avoid path parameters for domain identifiers when body or header transport is viable
8. reserve query string primarily for filtering and pagination

## Semantic Clarity Rule

Some route differences exist only for API clarity and external contract ergonomics.

Examples:

- `retry` instead of `redispatch`
- `session` instead of `server`
- `search` instead of ad hoc relation lookup names
- `download` instead of transport/history-specific route names when the operation is the same from the client point of view

When the difference is only semantic and not behavioral, the backend must not fork the business implementation unnecessarily.

Required rule:

1. different HTTP endpoints that represent the same operation should converge to the same backend function after the API controller layer
2. controllers should only translate HTTP contract details such as headers, body, auth, and response shape
3. business behavior must live in one shared application/backend function to reduce divergence and bugs
4. semantic aliases must not create duplicated business logic paths

Practical implication:

- if `/api/v5/messages/retry` replaces an older redispatch name, both routes should call the same backend retry operation
- if multiple read endpoints differ only in transport style (`GET` with headers vs `POST .../get` with body), they should resolve to the same backend read function
- if label search and related chat expansion are just different contract shapes, they should still converge to the same core query logic where possible

This rule is important to keep the new API clearer without increasing implementation risk.

## File and Registration Structure

Routes should also be grouped by family in the codebase.

Recommended registration split:

- `api_routes_system.go`
- `api_routes_auth.go`
- `api_routes_users.go`
- `api_routes_sessions.go`
- `api_routes_dispatches.go`
- `api_routes_messages.go`
- `api_routes_chats.go`
- `api_routes_groups.go`
- `api_routes_contacts.go`
- `api_routes_media.go`
- `api_routes_labels.go`

Recommended controller split:

- one controller file per family concern
- avoid files grouped only by historical source such as SPA extraction batches
- avoid large route registrars that manually list unrelated domains in one place

Controller rule:

- keep controllers thin
- parse request
- validate HTTP contract
- call one shared backend function/use case
- serialize response

Do not place business divergence inside multiple controllers when the operation is conceptually the same.

## Inline Documentation Requirement

All newly created or migrated API code should include inline documentation where it clarifies contract and responsibility.

This is especially important for:

- canonical v5 route registrars
- family-level controller files
- shared backend functions used by multiple endpoints
- compatibility aliases that intentionally point to the same behavior
- transport rules for header/body/query usage

Required documentation goals:

1. explain why a route exists when it is not obvious
2. explain when two endpoints intentionally share the same backend behavior
3. explain compatibility decisions during migration from legacy or `/spa`
4. explain identifier transport choices when using headers/body instead of path parameters

Inline documentation should favor architectural clarity, not noise.

## Testing Requirement

Every completed API migration slice should be followed by unit tests.

At minimum, tests should validate:

1. canonical route behavior
2. compatibility alias behavior when it still exists
3. shared backend behavior reused by multiple controllers
4. request validation and error cases
5. identifier transport through body, header, or query when supported

When two endpoints intentionally map to the same backend function, tests should prove that both routes preserve the same business result.

The migration is not considered complete for a slice until the relevant unit tests are in place.

## Migration Plan

## Phase 1 - Define Canonical Vocabulary

1. adopt `session` as the canonical API name for the current `server` resource
2. define canonical family roots under `/api/v5`
3. keep `v4` route structure unchanged
4. treat all current routes as compatibility surfaces that must keep working

## Phase 2 - Add Canonical Family Routes

1. introduce canonical resource-family routes without removing compatibility routes
2. expose the canonical new version under both `/api` and `/api/v5`
3. point all new frontend and integration work to the canonical new model
4. keep legacy action routes and `/spa` only as migration aliases where needed

## Phase 3 - Normalize Request/Response Contracts

1. unify send contracts
2. unify dispatch contracts
3. normalize pagination, list envelopes, and error envelopes
4. remove frontend-specific response hacks once the canonical contract is sufficient
5. remount useful `/spa` behavior under canonical v5 families

## Phase 4 - Deprecate Alias Surfaces

1. deprecate `/spa` paths after their `/api/v5` replacements are validated
2. deprecate action-style legacy routes only after canonical v5 replacements are stable
3. keep `/api` bound to the newest canonical version
4. keep only explicit compatibility aliases that are still operationally required

## Practical Recommendation for the Current Codebase

The next implementation step should not be a full rewrite.

The correct next step is:

1. create `src/api/legacy/` and move the current mixed route set there without changing runtime behavior
2. create `src/api/v5/` and register only canonical family-based routes there
3. keep existing handlers where possible, but mount them under canonical family paths first
4. gradually rename handler/controller files from SPA-oriented extraction names to family-oriented names
5. make `/api` point to the same route set as `/api/v5`
6. migrate frontend consumers from `/spa` to `/api` or `/api/v5`
7. remove `/spa` only after full feature parity is confirmed

This allows structural improvement without breaking current behavior.

## Immediate Naming Recommendation

Starting now, new documentation and new canonical route design should use:

- `users`
- `sessions`
- `dispatches`
- `messages`
- `chats`
- `groups`
- `contacts`
- `media`
- `labels`
- `system`
- `auth`

And the term `server` should be treated as a legacy/internal compatibility name until code and routes can be migrated safely.

## Explicit Non-Goals

The consolidation does not require:

- deleting existing API endpoints immediately
- creating a permanent separate frontend-only API
- introducing `/api/legacy` as a public URL namespace
- documenting a full route mapping table before implementation begins

The first priority is to preserve behavior, clean the structure, and make `/api/v5` the canonical destination for all future frontend and integration development.

## Stable Versioning Rule

From the next major API version onward, the API should follow this rule:

1. `/api` must always point to the newest canonical API version
2. `/api/v{current}` must expose the same newest canonical API version
3. the immediately older and legacy routes may remain available temporarily for compatibility
4. the previous version structure must not be silently rewritten in place

For the current migration:

1. `v4` stays as-is
2. `v5` introduces the new family-based model
3. `/api` points to `v5`
4. legacy endpoints remain available for a transition period
5. `/spa` is temporary and must be retired