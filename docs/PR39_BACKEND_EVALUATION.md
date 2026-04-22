# PR 39 Backend Evaluation

## Objective

Track the backend portion of PR `#39` (`feature/spa-sync-from-chat-20260224`) so we can extract useful changes without losing the thread or merging the SPA sync as a single opaque block.

Current branch baseline:

- local `develop` is aligned with `upstream/develop`
- PR `#39` diverges before the recent server delete, webhook, WID sync, and connection-state work merged into `develop`
- GitHub shows PR `#39` as conflicting, but local merge testing showed the practical conflict is only `.gitignore`

## Current Position

Working decision as of `2026-04-22`:

- do not merge PR `#39` as-is
- review backend first
- extract backend changes in smaller slices after validating value and compatibility
- treat frontend and generated assets as a separate stage

Current extraction status:

- first backend slice integrated into local `develop`
- second backend read slice integrated into local `develop`
- third backend message slice integrated into local `develop`
- SPA send wrapper integrated into local `develop`
- SPA archive/presence wrappers integrated into local `develop`
- cable realtime command surface expanded beyond the PR websocket scope
- SPA webhook/rabbitmq wrappers integrated into local `develop`
- explicit SPA server debug/option toggles integrated into local `develop`
- SPA server create/update/delete wrappers integrated into local `develop`
- remaining PR `environment` and user-mutation handlers intentionally deferred
- dedicated websocket cable module implemented locally outside PR `#39`
- no model-layer changes imported from PR `#39`
- no frontend files imported yet

## Backend Scope In PR 39

Files changed in backend-related areas:

- `src/api/api.go`
- `src/api/api_handlers+HistoryDownloadController.go`
- `src/api/api_handlers+LoginController.go`
- `src/api/api_handlers+SPAController.go`
- `src/api/api_handlers+ServerEnableDisableController.go`
- `src/api/api_handlers.go`
- `src/api/api_spa_routes.go`
- `src/api/api_spa_utils.go`
- `src/api/api_websocket_client.go`
- `src/api/api_websocket_hub.go`
- `src/environment/api_settings.go`
- `src/environment/branding_settings.go`
- `src/environment/environment_settings.go`
- `src/environment/form_settings.go`
- `src/environment/general_settings.go`
- `src/form/form.go`
- `src/form/form_handlers.go`
- `src/form/form_json_handlers.go`
- `src/models/dispatching_handler.go`
- `src/models/qp_receive_response.go`
- `src/models/qp_webhook_test.go`
- `src/models/qp_whatsapp_extensions.go`
- `src/models/qp_whatsapp_server.go`
- `src/models/qp_whatsapp_server_delete_test.go`
- `src/models/qp_whatsapp_server_dispatching.go`
- `src/models/qp_whatsapp_server_extensions.go`
- `src/models/sqlite_migration.go`
- `src/swagger/docs.go`
- `src/swagger/swagger.json`
- `src/swagger/swagger.yaml`
- `src/webserver/webserver.go`

Diff size for these backend areas:

- `31` files
- about `+4601 / -463`

## Initial Classification

### Likely Useful

- `src/api/api_spa_routes.go`
  - defines a coherent authenticated SPA API surface under `/api`
- `src/api/api_handlers+LoginController.go`
  - exposes login/config payload for SPA bootstrap
- `src/api/api_handlers+ServerEnableDisableController.go`
  - small, isolated start/stop endpoints
- `src/api/api_handlers+HistoryDownloadController.go`
  - targeted feature for history-sync media recovery
- `src/api/api.go`
  - mounts SPA controllers under API prefix and keeps legacy routes available
- `src/webserver/webserver.go`
  - introduces SPA fallback/proxy behavior for dev and production

### Useful But Needs Adaptation

- `src/environment/api_settings.go`
  - sensible defaulting of `API_PREFIX` to `api`
  - needs review because PR also brings dev-oriented and insecure example defaults elsewhere
- `src/environment/environment_settings.go`
  - adds `Branding` to global settings and environment expansion logic
- `src/environment/branding_settings.go`
  - useful if we want login and SPA branding customization
- `src/form/*`
  - may be needed to support SPA-first auth/session flow, but needs compatibility review with current form behavior
- `src/models/sqlite_migration.go`
  - likely productively related to onboarding/setup flow, but not obviously required for core SPA backend extraction

### High-Risk Or Oversized

- `src/api/api_handlers+SPAController.go`
  - monolithic file with more than `2200` lines
  - likely mixes multiple concerns: session, server CRUD, messaging, groups, webhooks, rabbitmq, environment
  - strong extraction candidate only after splitting by domain
- `src/api/api_handlers.go`
  - broad routing/controller changes mixed with unrelated API behavior updates
- `src/swagger/*`
  - generated output should only be refreshed after backend extraction is stable

### Probably Do Not Import From PR 39

- changes in `src/models/dispatching_handler.go`
- changes in `src/models/qp_webhook_test.go`
- changes in `src/models/qp_whatsapp_server.go`
- deletion of `src/models/qp_whatsapp_server_delete_test.go`
- changes in `src/models/qp_whatsapp_server_dispatching.go`
- changes in `src/models/qp_whatsapp_server_extensions.go`
- `src/api/api_websocket_client.go`
- `src/api/api_websocket_hub.go`

Reason:

- `develop` already moved forward in delete flow, webhook dispatch semantics, WID synchronization, and connection-state behavior
- PR `#39` is older on these concerns
- the PR websocket implementation is QR/verification-oriented and too narrow for the command/event realtime layer we now need
- importing these model-layer diffs would risk regressing current behavior

## Backend Gains

### 1. SPA Route Surface

Main gain:

- PR `#39` creates a consistent authenticated SPA route layer under `/api`

What it enables:

- `/api/session`
- `/api/servers`
- `/api/server/{token}/info`
- `/api/server/{token}/messages`
- `/api/webhooks`
- `/api/rabbitmq`
- `/api/environment`
- `/api/verify/ws`

Why this matters:

- frontend stops scraping legacy endpoints ad hoc
- route design becomes easier to reason about for a dedicated SPA client
- we can preserve current public/legacy API while offering a cleaner app-facing surface

### 2. Public Login Bootstrap Endpoint

Main gain:

- `LoginConfigController` provides app title, branding, login layout details, and setup flags to the login screen

Why this matters:

- lets the SPA render login/setup dynamically
- removes hardcoded login branding from the frontend
- creates a clean extension point for white-labeling if we want it

### 3. Server Enable / Disable Endpoints

Main gain:

- dedicated SPA endpoints to start and stop servers

Why this matters:

- cleaner than overloading generic command routes
- more explicit frontend behavior
- lower integration risk because the handlers are small and isolated

### 4. History-Sync Media Download

Main gain:

- support to fetch media referenced by WhatsApp history-sync protocol messages

Why this matters:

- reduces gaps in SPA message history UX
- adds a targeted recovery path for media that exists in protocol metadata but is not yet attached locally

Open question:

- we need to verify this logic against current message/debug structures before adoption, because it assumes specific `ProtocolMessage` internals and connection availability

### 5. SPA-Friendly Webserver Behavior

Main gain:

- dev proxy to Vite
- production static serving with SPA fallback to `index.html`
- explicit exclusion of API and legacy API paths from SPA fallback

Why this matters:

- this is the backend half that makes a Vue SPA actually operable inside the same binary
- without this, the frontend import is incomplete

## Main Risks

### 1. Monolithic SPA Controller

`src/api/api_handlers+SPAController.go` is too large to trust as a single import unit.

Risk:

- hidden regressions
- duplicated logic already solved differently in current `develop`
- hard future maintenance

Preferred direction:

- split by domain before adoption:
  - session/account
  - server lifecycle
  - messages/media
  - groups/contacts
  - webhooks/rabbitmq
  - environment/users

### 2. Environment Defaults And Branding Noise

PR `#39` brings several defaults in `.env.example` that should not be accepted blindly:

- `SIGNING_SECRET=dev-signing-secret-1234567890-abc`
- `MASTERKEY=dev-master-key`
- `MCP_ENABLED=true`
- `LOGLEVEL=DEBUG`
- `WHATSMEOW_LOGLEVEL=DEBUG`
- `WHATSMEOW_DBLOGLEVEL=DEBUG`
- `MIGRATIONS=./migrations`
- `APP_TITLE="Hermes"`
- malformed line `HTTPLOGS=HTTPLOGS=true`

Impact:

- poor production defaults
- unnecessary local branding leakage
- confusion in onboarding docs

### 3. Model-Layer Drift Against Current Develop

The PR touches model files that overlap conceptually with work already merged later into `develop`.

Impact:

- possible regression in delete semantics
- possible loss of newer webhook/WID synchronization behavior
- test churn without product gain

### 4. CI Signal Is Misleading

PR `#39` shows failing checks on GitHub, but the failure was not the Go build itself.

Observed issue:

- workflow tried to run release publishing from a PR context and failed due permission constraints

Implication:

- CI red does not by itself invalidate backend code quality
- but it does mean we cannot use the current PR status as a merge-readiness signal

## SPAController Breakdown

`src/api/api_handlers+SPAController.go` is the main extraction problem. First-pass split:

### Lower-Risk Read-Oriented Handlers

- `SPASessionController`
- `SPAServersController`
- `SPAServersSearchController`
- `SPAAccountController`
- `SPAMasterKeyController`
- `SPAServerInfoController`
- `SPAServerQRCodeController`
- `SPAServerPairCodeController`
- `SPAUsersListController`
- `SPAServerContactsController`
- `SPAServerGroupsController`

Why these are attractive:

- mostly query current state instead of mutating it
- easier to adapt to current `develop`
- useful to unlock SPA navigation and dashboards early

### Medium-Risk, Likely Worth Extracting

- `SPAServerMessagesController`
- `SPAServerEditMessageController`
- `SPAServerRevokeMessageController`
- `SPAServerArchiveChatController`
- `SPAServerPresenceController`
- `SPAServerDownloadMediaController`
- `SPAServerSendController`
- `SPAVerifyWebSocketController`

Why these need a closer pass:

- they interact with live server/session state
- they may depend on current message cache semantics
- they can still be valuable because they map well to existing product behavior

### High-Risk Mutation Handlers

- `SPAServerCreateController`
- `SPAServerDeleteController`
- `SPAServerUpdateController`
- `SPAServerDebugController`
- `SPAToggleController`
- `SPAWebHooksController`
- `SPAWebHooksCreateController`
- `SPAWebHooksDeleteController`
- `SPAWebHooksUpdateController`
- `SPARabbitMQController`
- `SPARabbitMQCreateController`
- `SPARabbitMQDeleteController`
- `SPAUserController`
- `SPAUserDeleteController`
- `SPAEnvironmentController`

Why these are high risk:

- overlap with model-layer changes that have already evolved in current `develop`
- touch server lifecycle, dispatching, and configuration semantics
- `SPAEnvironmentController` also needs a security review because it centralizes environment exposure

## What To Extract First

Recommended backend extraction order:

- [ ] `src/api/api_handlers+LoginController.go`
- [ ] `src/api/api_handlers+ServerEnableDisableController.go`
- [ ] `src/api/api_handlers+HistoryDownloadController.go`
- [ ] `src/api/api_spa_routes.go`
- [ ] `src/api/api.go`
- [ ] `src/webserver/webserver.go`
- [ ] read-only handlers split out of `src/api/api_handlers+SPAController.go`
- [ ] minimal environment support strictly required by the extracted handlers
- [ ] swagger regeneration only after all chosen backend handlers are in place

## Recommended First Backend Slice

If we want a safe first extraction, start with:

- `LoginConfigController`
- `SPAServerEnableController`
- `SPAServerDisableController`
- `SPAServerHistoryDownloadController`
- route registration support in `api.go`
- SPA webserver fallback/proxy in `webserver.go`

Then evaluate read-only SPA handlers:

- session
- servers listing/search
- account
- server info
- QR/pair code
- contacts/groups

Leave for later:

- delete/update/toggle
- webhooks/rabbitmq CRUD
- environment viewer
- user mutation

## Integrated From PR 39

The following backend ideas from PR `#39` have now been integrated locally, with compatibility-oriented adaptations:

- public `login/config` endpoint
- authenticated SPA route registration scaffold
- authenticated SPA read endpoints under `/spa`
- authenticated SPA message routes under `/spa`
- authenticated SPA send route under `/spa`
- authenticated SPA chat archive/presence routes under `/spa`
- authenticated SPA webhook/rabbitmq routes under `/spa`
- authenticated SPA server debug/option toggle routes under `/spa`
- authenticated SPA server create/update/delete routes under `/spa`
- SPA server enable endpoint
- SPA server disable endpoint
- SPA history-sync media download endpoint
- SPA webserver fallback/proxy support

Files added or updated locally for this first slice:

- `src/api/api_handlers+LoginController.go`
- `src/api/api_handlers+ServerEnableDisableController.go`
- `src/api/api_handlers+HistoryDownloadController.go`
- `src/api/api_handlers+SPAMessageController.go`
- `src/api/api_handlers+SPAReadController.go`
- `src/api/api_spa_routes.go`
- `src/api/api_spa_utils.go`
- `src/api/api.go`
- `src/api/api_handlers.go`
- `src/api/models/receive_response.go`
- `src/webserver/webserver.go`
- `src/cable/*`
- `src/models/realtime_publishers.go`

### Important Adaptations

- We did **not** change the global default for `API_PREFIX`.
- We mounted the first SPA routes compatibly:
  - SPA-only endpoints live under `/spa`
  - legacy/shared API stays on the standard API surface
- We did **not** import branding settings, login customization environment keys, or `.env.example` defaults from PR `#39`.
- We did **not** import any model-layer changes from PR `#39`.
- We did **not** adopt the PR websocket files as the base realtime transport.
- We created a dedicated `cable` module with a stable command/event protocol under `GET /cable`.
- SPA fallback in `webserver` only activates when:
  - frontend dev proxy is explicitly enabled, or
  - `assets/frontend/index.html` exists

### SPA Read Endpoints Integrated

The current `/spa` read surface includes:

- `GET /spa/session`
- `GET /spa/servers`
- `POST /spa/servers/search`
- `GET /spa/account`
- `GET /spa/account/masterkey`
- `GET /spa/server/{token}/info`
- `GET /spa/server/{token}/qrcode`
- `GET /spa/server/{token}/paircode`
- `GET /spa/users`
- `GET /spa/server/{token}/contacts`
- `GET /spa/server/{token}/groups`

The current `/spa` message and media surface includes:

- `POST /spa/server/{token}/send`
- `POST /spa/server/{token}/chat/archive`
- `POST /spa/server/{token}/chat/presence`
- `GET /spa/server/{token}/messages`
- `PUT /spa/server/{token}/message/{messageid}/edit`
- `DELETE /spa/server/{token}/message/{messageid}`
- `GET /spa/server/{token}/download/{messageid}`
- `POST /spa/server/{token}/messages/{messageid}/history/download`
- `POST /spa/server/{token}/enable`
- `POST /spa/server/{token}/disable`

The current `/spa` dispatching/config surface includes:

- `POST /spa/server/create`
- `PATCH /spa/server/{token}`
- `DELETE /spa/server/{token}`
- `POST /spa/server/{token}/debug/toggle`
- `POST /spa/server/{token}/option/{option}/toggle`
- `GET /spa/server/{token}/webhooks`
- `POST /spa/server/{token}/webhooks`
- `DELETE /spa/server/{token}/webhooks`
- `GET /spa/server/{token}/rabbitmq`
- `POST /spa/server/{token}/rabbitmq`
- `DELETE /spa/server/{token}/rabbitmq`

Read-path behavior notes:

- server listing uses persisted DB records so disconnected servers still appear
- live-only endpoints such as contacts/groups still require the server to exist in memory
- ownership checks are done against persisted server records before live access
- SPA send reuses the current `SendAnyWithServer` flow instead of the PR's reduced text-only payload
- SPA archive/presence reuse the current `develop` controller behavior, including validation and typing-duration handling
- SPA webhook/rabbitmq routes reuse the current dispatching CRUD behavior from `develop` instead of the older PR-specific handlers
- SPA debug/option toggles use explicit per-route actions instead of reviving the PR's generic `/toggle` endpoint
- SPA server create/update/delete reuse the current `develop` information/service primitives instead of the PR's separate mutation handlers
- SPA message listing reuses the current receive filters and adds SPA-only pagination
- SPA media downloads keep the current attachment behavior by default and only switch to inline disposition when requested explicitly

### Validation Result

Validated after integration:

- `gofmt` on edited Go files
- `go test ./...` from `src`

Result:

- passed locally
- only existing `sqlite3` compiler warnings were observed

## Medium-Risk Handler Review

Review focus completed for the medium-risk handlers previously identified inside `src/api/api_handlers+SPAController.go`.

### Worth Extracting With Adaptation

- `SPAServerMessagesController`
  - useful because SPA pagination is real product value that the current legacy `/receive` route does not provide
  - should **not** be copied as-is:
    - it authenticates through `GetFormUserFromRequest` instead of the new SPA JWT helper path
    - it resolves the server only through `models.GetServerFromToken`, which assumes live in-memory state
    - it only supports `timestamp`, `exceptions`, `page`, and `limit`, which is narrower than the current receive filter surface
  - preferred extraction:
    - keep the SPA route shape
    - reuse the current receive filtering helpers
    - add pagination on top of the already-adapted `/spa` ownership/live-server lookup path

- `SPAServerEditMessageController`
  - route shape is useful for SPA UX because the message id sits in the path
  - implementation is only a thin wrapper around current `server.Edit(...)`
  - should be adapted to use SPA auth/ownership helpers rather than imported directly

- `SPAServerRevokeMessageController`
  - same general conclusion as edit
  - worth exposing on `/spa`, but as a compatibility wrapper over current revoke behavior instead of a raw PR import

- `SPAServerDownloadMediaController`
  - useful for SPA media preview/download flows
  - close to current legacy `DownloadController`, but changes behavior in notable ways:
    - forces `cache=false`
    - returns `inline` disposition instead of `attachment`
    - skips the current filename fallback logic based on MIME type
  - preferred extraction:
    - keep SPA route
    - reuse current download logic
    - make disposition/cache behavior an explicit SPA choice instead of silently importing the PR behavior

### Do Not Import As-Is

- `SPAServerSendController`
  - lower value than it first appears because current `develop` already has a richer send surface and a cable `message.send` command
  - PR version only supports:
    - `recipient`
    - plain text `message`
    - optional `id`
  - it drops current capabilities such as attachments, polls, locations, contacts, richer request validation, and cable-oriented send flows
  - conclusion: do not import; if SPA needs HTTP send, build it from the current send request model instead

- `SPAServerArchiveChatController`
  - redundant with the existing `ArchiveChatController`
  - PR version is weaker than current `develop`:
    - no explicit ready-state guard
    - less detailed validation/response handling
    - direct type assertion on the connection path
  - conclusion: do not import directly; expose SPA access later by delegating to the current archive implementation

- `SPAServerPresenceController`
  - redundant with the existing `ChatPresenceController`
  - PR version is materially worse:
    - string parsing instead of the typed request model
    - ignores the current async duration handling for typing indicators
    - narrower response and validation behavior
  - conclusion: do not import directly

- `SPAVerifyWebSocketController`
  - no longer fits the architecture direction
  - it is tied to the PR's QR verification websocket path, while local `develop` now has a dedicated `cable` module for authenticated realtime commands/events
  - conclusion: keep the current cable direction and do not reuse the PR websocket implementation

### Recommended Third Backend Slice

If we continue extracting backend value from PR `#39`, the next slice should be:

- `GET /spa/server/{token}/messages`
- `PUT /spa/server/{token}/message/{messageid}/edit`
- `DELETE /spa/server/{token}/message/{messageid}`
- `GET /spa/server/{token}/download/{messageid}`

Implementation rule for that slice:

- preserve the SPA route ergonomics from PR `#39`
- do **not** import the handler bodies verbatim
- adapt them onto current `develop` helpers and controller logic
- leave SPA send/archive/presence on the existing legacy or cable surfaces until we introduce shared internal service helpers

## Deferred High-Risk Scope

The remaining backend endpoints from PR `#39` that are still intentionally not imported are:

- `SPAEnvironmentController`
- `SPAUserController`
- `SPAUserDeleteController`

Current decision:

- do not extract them yet

Reasoning:

- no local frontend or workspace reference currently requires dedicated SPA wrappers for those flows
- `SPAEnvironmentController` still needs a security review because the legacy `/environment` endpoint has explicit master-key semantics that should not be bypassed accidentally
- user mutation is narrower product value than the server/message/dispatching work already integrated
- the current `/spa` + `cable` surface is now broad enough to support the backend half of the SPA without adding those higher-risk endpoints prematurely

## Explicit Non-Goals For The First Backend Pass

- do not import the entire `src/api/api_handlers+SPAController.go` at once
- do not import frontend files yet
- do not overwrite current model-layer delete/webhook/WID behavior with older logic from the PR
- do not regenerate swagger until route and handler selection is stable

## Decision Log

### 2026-04-22

- Created this tracking document.
- Confirmed local `develop` matches `upstream/develop`.
- Confirmed backend and frontend builds from PR `#39` can pass locally when dependencies are installed.
- Confirmed GitHub check failures on the PR were caused by release workflow behavior, not by backend compilation failure.
- Chosen approach: backend-first extraction and review, not raw merge.
- Integrated the first backend slice locally:
  - `LoginConfigController`
  - `SPAServerEnableController`
  - `SPAServerDisableController`
  - `SPAServerHistoryDownloadController`
  - minimal SPA route registration
  - SPA webserver fallback/proxy support
- Integrated the second backend read slice locally:
  - session/account reads
  - server listing/search reads
  - server info reads
  - QR/pair code reads
  - contacts/groups reads
  - `/spa` auth/ownership helpers
- Reviewed the PR medium-risk SPA handlers:
  - keep messages/edit/revoke/download as extraction candidates with adaptation
  - reject direct import of send/archive/presence/verify websocket handlers
- Preserved current `develop` semantics for model-layer delete, webhook, WID, and dispatching behavior.
- Integrated the third backend SPA message slice locally:
  - paginated SPA message listing
  - SPA edit wrapper
  - SPA revoke wrapper
  - SPA media download wrapper
  - shared SPA live/ready server helpers
- Integrated SPA send as a wrapper over the current send request model:
  - preserves current send capabilities for text, attachments, URLs, base64 content, polls, locations, and contacts
  - intentionally does **not** import the PR's reduced send payload directly
- Integrated SPA archive/presence as wrappers over shared `develop` helpers:
  - archive keeps current ready-state and chat-id validation behavior
  - presence keeps current typed request handling and async duration cancellation semantics
- Expanded `cable` instead of reviving the PR websocket verification path:
  - added realtime commands for `message.edit`
  - added realtime commands for `message.revoke`
  - added realtime commands for `chat.archive`
  - added realtime commands for `chat.presence`
  - kept the command/event bus as the preferred websocket direction instead of the PR's QR-only socket
- Integrated SPA webhook/rabbitmq wrappers over shared `develop` helpers:
  - webhook CRUD now reuses the current dispatching controller behavior
  - rabbitmq CRUD now reuses the current dispatching controller behavior
  - we still do **not** import the older PR CRUD handlers directly
- Integrated explicit SPA server toggles over current `develop` primitives:
  - `debug` uses `ToggleDevel`
  - `groups`, `broadcasts`, `readreceipts`, `calls`, and `readupdate` use explicit toggle routes
  - we still do **not** import the PR's generic `/toggle` handler directly
- Integrated SPA server create/update/delete over current `develop` primitives:
  - `create` uses `AppendNewServer` plus the current persisted info model
  - `update` uses `updateServerConfiguration`
  - `delete` uses `WhatsappService.Delete`
  - we still do **not** import the PR mutation handlers directly
- Deferred the remaining high-risk backend scope:
  - no local consumer currently justifies SPA environment or user-mutation wrappers
  - keep the legacy `/environment` semantics unchanged until there is a concrete requirement and review

## Next Step

Next review slice:

- backend extraction from PR `#39` is now sufficient for the current SPA baseline
- if we continue this PR evaluation, move to the frontend/import side as a separate stage
- keep rejecting direct import of PR environment/user mutation handlers unless a concrete frontend requirement appears
