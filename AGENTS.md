- ✅ Database configuration documentation slice: clarified that `DB*` variables configure the Whatsmeow SQL store, not the internal `quepasa.sqlite` application DB; updated inline comments plus `src/environment/README.md`, `src/.env.example`, root `README.md`, and `docker/docker.md` to explain sqlite defaults vs postgres/mysql-only fields
- ✅ N8N integration tests slice: created comprehensive unit test suite for all n8n→QuePasa API v4 requests; `src/api/api_n8n_quepasa_integration_test.go` covers 29 test cases (send text, group invites, media download, contact info, webhooks, auth methods, request/response format, integration scenarios, error handling, edge cases); all 29 tests passing (0.365s execution); created 3 documentation files (`docs/N8N_QUEPASA_API_REQUESTS_MAPPING.md`, `docs/N8N_QUEPASA_API_TESTS_README.md`, `docs/N8N_INTEGRATION_TESTS_SUMMARY.md`); full project build passing (exit 0)
- ✅ Multi-login session-scope slice: protected SPA
- ✅ Multi-login session-scope slice: protected SPA/canonical routes now accept two auth modes — JWT (`Authorization`) for user-wide access and `X-QUEPASA-TOKEN` for single-session-scoped access; token-auth mode resolves user by session ownership and forces token-scoped operations/listing to that authenticated session; `POST /api/sessions` keeps JWT path and supports custom token override only when `RELAXED_SESSIONS=true`; canonical tests cover custom-token create success (201), strict-mode 403, master-key-only 401, scoped-list isolation, and scoped-token forced routing
- ✅ Authentication docs slice: added `docs/USAGE-authentication-modes.md` documenting the 4 auth types (`X-QUEPASA-TOKEN`, `X-QUEPASA-MASTERKEY`, JWT, anonymous), scope model, precedence, and curl examples for canonical API usage
- ✅ Vuejs LID tooling slice: added two dedicated SPA screens under `src/apps/vuejs/client/src/pages` — `LIDDirectSend.vue` for direct `@lid` send tests and `LIDMappings.vue` for bidirectional `@lid`<->phone lookup; wired new routes in `router.ts` and added quick-access cards in `Server.vue`
- ✅ Canonical identifier lookup route slice: added `GET /contacts/identifier` canonical route delegating to `GetUserIdentifierController` so SPA can query LID mappings through canonical API with token middleware
- ✅ Direct @lid testing endpoint slice: added `POST /messages/lid/direct` (`SendLIDDirectController`) for explicit `@lid` text sends without API-layer LID->phone conversion from standard send flow; route wired in canonical message routes; Swagger regenerated; build + api tests passing
- ✅ LID outbound routing investigation slice: documented why `@lid` sends can fail when LID->phone mapping is missing; mapped code flow across `FormatEndpoint`, `SendController`, `WhatsmeowConnection.Send`, and `WhatsmeowContactManager` store resolvers; created `docs/LID_MESSAGE_ROUTING.md` with diagnosis + recommended stable policy (resolve `@lid` to `@s.whatsapp.net` before send)
- ✅ Status/Stories feature slice: `PublishStatus(text, attachment)` added to `IWhatsappConnection` and `WhatsmeowConnection`; sends to `types.StatusBroadcastJID`; `POST /status/publish` endpoint via `PublishStatusController`; `UserAbout` and `UserStatusMute` events registered in event router; Swagger regenerated; build + all tests passing (exit 0)
- ✅ PLAN-ARCHITECTURE-REFACTORING.md updated: all 6 phases marked COMPLETED with accurate implementation notes
- ✅ Phase 6 dispatching extraction slice: RabbitMQ config + dispatching query methods (whatsmeow handlers): `onConnectFailureEvent`, `onStreamErrorEvent`, `onTemporaryBanEvent` now publish to internal event bus via `qpevents.Publish`; `OnEventBlocklist` also publishes `whatsapp.blocklist.updated`; events consumed by metrics Prometheus subscriber; `whatsmeow/go.mod` synced with events dependency; full build passing (exit 0)
- ✅ Whatsmeow event handler tests: `whatsmeow_handlers_events_test.go` covers 4 cases (connect failure, stream error, temp ban known code, temp ban unknown code); all 4 passing via default event bus subscription pattern
- ✅ Group Invite Link revoke slice: `RevokeInvite` added to `WhatsappGroupManagerInterface`, `WhatsmeowGroupManager`, `QpGroupManager`; `SPAGroupRevokeInviteController` + `RevokeGroupInviteLinkController` with Swagger; `DELETE /groups/invite` canonical route; Swagger regenerated; full build passing (exit 0)
- ✅ Group Invite Link revoke slice: `RevokeInvite` added to `WhatsappGroupManagerInterface`, `WhatsmeowGroupManager`, `QpGroupManager`; `SPAGroupRevokeInviteController` + `RevokeGroupInviteLinkController` with Swagger; `DELETE /groups/invite` canonical route; Swagger regenerated; full build passing (exit 0)
- ✅ SPA route path param fix: invite GET+DELETE routes changed from `/{groupid}/invite` path param pattern to `/groups/invite` query param pattern using `withCanonicalParams(canonicalGroupIDParam)` middleware; build passing (exit 0)
- ✅ Ephemeral messages slice: `ExpiresAt int64` added to `WhatsappMessage` (`json:"expiresat,omitempty"`); `extractExpirationFromMessage()` helper checks `ContextInfo.Expiration` across all common message types; `HandleEphemeralMessage` reimplemented to recursively call `HandleKnowingMessages` + set `ExpiresAt`; `evt.IsEphemeral` check in `handler.Message()` for the normal auto-unwrap flow; full build passing (exit 0)
- ✅ Phase 2 transport thread-safety slice: `transportServicesMu` (`sync.RWMutex`) now guards startup-time transport adapter globals; `ApplyTransportServices` writes under lock; realtime/RabbitMQ/lifecycle adapter reads go through lock-protected helper paths; `qp_rabbitmq_config.go` now uses thread-safe exchange/routing getters; full build + `go test` in `src/models` passing (exit 0)
- ✅ Ephemeral tests slice: new `whatsmeow_handlers_message_extensions_test.go` covers expiration extraction and ephemeral wrapper handling (`ExpiresAt` set from `ContextInfo.Expiration`, inner message processed, existing `ExpiresAt` preserved); `go test` in `src/whatsmeow` and full build passing (exit 0)
- ✅ Phase 6 decomposition slice: extracted state/options methods from `qp_whatsapp_server.go` into new `server_state.go` (`GetValidConnection`, options setters/getters, `HandlerEnsure`, `HasSignalRActiveConnections`, `GetStatus`, `GetState`, `GetWId`) to further slim the core file while preserving behavior; `go test` passing in `src/models` and `src/whatsmeow`; full build passing (exit 0)
- ✅ Phase 6 decomposition slice: extracted JSON serialization from `qp_whatsapp_server.go` into new `server_serialization.go` (`MarshalJSON` + custom payload shape), reducing core file responsibility while preserving API/webhook shape; `go test` passing in `src/models` and `src/whatsmeow`; full build passing (exit 0)
- ✅ Phase 6 decomposition slice: extracted transport adapter globals/helpers from `qp_whatsapp_server.go` into new `server_transport_adapters.go` (`RealtimePresenceChecker`, `HasActiveRealtimeConnections`, `GlobalRabbitMQClientResolver`, `ResolveRabbitMQClient`), further slimming core server definition; `go test` passing in `src/models` and `src/whatsmeow`; full build passing (exit 0)
- ✅ Public user creation endpoint slice: `POST /api/users` now implements bootstrap-friendly logic — first user can be created without master key, subsequent users require `X-QUEPASA-MASTERKEY` header; `SPAPublicUserCreateController` counts persisted users via `runtime.CountPersistedUsers()` and gates master key check on count > 0; uses `IsMatchForMaster(r)` for test compatibility; `api_user_creation_test.go` covers 5 scenarios (first user no key, second user requires key, second user with key, invalid password, duplicate email); all tests passing (exit 0); full build passing (exit 0); integrated into canonical route `/api/users` (POST public, GET+DELETE protected)
- ✅ Environment discovery endpoint slice: `GET /api/system/environment` now returns preview (anônimo) vs full settings (com master key); anonymous requests show public features only (groups, broadcasts, calls, history_sync, log_level, presence, etc) in humanized format; master key requests return complete config (API, Database, WebServer, Cache, RabbitMQ, Redis, etc); `api_environment_discovery_test.go` covers 3 scenarios (anonymous preview, master key full access, wrong key treated as anonymous); all tests passing (exit 0); `docs/USAGE-environment-discovery.md` documents full payload examples and field descriptions; README updated with environment discovery guide link; build passing (exit 0)
- ✅ Configurable default API version slice: added `API_DEFAULT_VERSION` env (default `v4`) to control which version handles the unversioned `/<API_PREFIX>/...` alias while keeping explicit `/v4`, `/v5`, and `/v3` routes alive in parallel; environment preview/full settings now expose the selected default version; canonical Vue SPA requests are normalized to explicit `/api/v5/...` in `src/apps/vuejs/client/src/services/api.ts` so the official frontend is isolated from alias switching during migration
# Task: Native WhatsApp VoIP Inbound Call Acceptance and Audio Playback

## Task Objective

Implement an evidence-driven native WhatsApp VoIP path in QuePasa focused on inbound calls: receive a real incoming call, accept it natively, negotiate relay/media successfully, and reach deterministic prerecorded audio playback directly inside the call.

## Mandatory Checklist

- [ ] Phase 0: Lab preparation and evidence capture
  - [x] Validate local Linux + Go installation without Docker
  - [x] Confirm local `systemd` service startup for QuePasa
  - [ ] Create dedicated VoIP `.env` profile
  - [ ] Prepare packet capture tooling and media fixtures
  - [ ] Capture baseline real-app-to-real-app call for comparison

- [ ] Phase 1: Signaling observability
  - [ ] Add dedicated call-session state structures
  - [ ] Track lifecycle by `CallID`
  - [ ] Expand observed inbound call event coverage
  - [ ] Add structured signaling debug logs and payload dumps
  - [ ] Add experimental feature flags for call behavior

- [ ] Phase 2: Native accept path
  - [ ] Identify exact accept node structure
  - [ ] Implement isolated call accept helper
  - [ ] Add auto-answer experimental mode
  - [ ] Validate answered call reaches relay setup

- [ ] Phase 3+: TURN, SRTP, RTP, and Opus media pipeline
  - [ ] Implement TURN auth/allocation path
  - [ ] Implement SRTP derivation and protection
  - [ ] Implement RTP packetizer and Opus framing experiments
  - [ ] Validate prerecorded audio playback heard by caller

## Current Status

- ✅ Local QuePasa installation validated without Docker using Linux service + direct Go runtime
- ✅ `quepasa.service` is active locally on port `31000`
- ✅ Native VoIP implementation plan created in `docs/PLAN-whatsapp-native-voip-inbound-call-audio.md`
- ✅ Existing local baseline identified:
  - `CallOffer` and `CallOfferNotice` are observed in `src/whatsmeow/whatsmeow_event_router.go`
  - calls are converted into internal call messages in `src/whatsmeow/whatsmeow_handlers.go`
  - calls can be rejected when call handling is disabled
- ✅ Important architectural direction confirmed: `src/sipproxy/` is out of scope for the first native WhatsApp VoIP milestone
- ⏳ Next implementation focus: Phase 0/1 execution, starting with dedicated signaling observability and call-state tracking

## Next Steps

1. Create a reproducible local VoIP experiment profile and fixture directory
2. Install/validate packet capture and media conversion tooling
3. Add dedicated `whatsmeow_call_*` files for call state and signaling instrumentation
4. Capture inbound offer payload details by `CallID`
5. Prove native answer without media before attempting TURN/SRTP/audio

## Immutable Constraints

1. The first milestone is native WhatsApp inbound voice, not message fallback
2. `sipproxy` is not the implementation path for this milestone
3. Validation must happen locally on Linux without Docker
4. Each layer must be proven with captures/logs before advancing to the next
5. Audio playback inside the call is the target outcome, not chat audio attachment fallback

## Previous Context (Preserved)

- [ ] Layer 1: Session Foundation (type aliases, wrapper functions in models)
  - [x] Create qp_whatsapp_session.go with type aliases and wrappers (COMPLETED)
  - [x] Create qp_whatsapp_session_dispatching.go with dispatching wrapper (COMPLETED)
  - [x] Write tests for foundation layer (6 tests passing) (COMPLETED)

- [ ] Layer 2: API Session Helpers & Controllers
  - [x] Create api_session_extensions.go for HTTP request helpers (COMPLETED)
  - [x] Create api_handlers+SPASessionController.go with 12 controller wrappers (COMPLETED)
  - [x] Update api_routes_sessions.go to use session controllers (COMPLETED)
  - [x] Write tests for API layer (3+ tests passing, including canonical route test) (COMPLETED)

- [ ] Layer 3: SPA Utility Function Wrappers
  - [x] Create api_spa_session_utils.go with utility wrappers (COMPLETED)
  - [x] Create api_spa_session_utils_test.go with validation tests (7 tests passing) (COMPLETED)
  - [x] Validate compilation without errors (COMPLETED)
- [ ] Layer 3b: Call Site Migration (Optional)
  - [x] Update 13 respondSPAServer* call sites in SPAMessageController (COMPLETED)
  - [x] Rename functions to respondSPASession* in SPAMessageController (COMPLETED)
  - [x] Add backward-compatible aliases for server-named functions (COMPLETED)
  - [x] Validate all tests passing (8 total session tests) (COMPLETED)
  - [x] Validate compilation without errors (COMPLETED)
- [ ] Phase 0-1 Planning & Documentation
  - [x] Create PLAN-ARCHITECTURE-REFACTORING.md with full analysis (COMPLETED)
  - [ ] Update AGENTS.md with progress tracking (IN PROGRESS)

## Current Status

**Completed Work (All Phases):**
- ✅ Layer 1: Foundation - Type aliases, wrapper functions (6 tests passing)
- ✅ Layer 2: API Helpers & Controllers - Request-level helpers, 12 SPA controllers (3+ tests passing)
- ✅ Layer 3: SPA Utilities - 6 wrapper functions for utility layer (5 tests passing)
- ✅ Layer 3b: Call Site Migration - 13 call sites updated in SPAMessageController (8 tests total passing)
- ✅ Architecture documentation expansion - current state, target state, and roadmap docs added under `docs/`
- ✅ Architecture package map documentation added under `docs/`
- ✅ Architecture index and execution checklist added under `docs/`
- ✅ Architecture ADR set added under `docs/`
- ✅ Runtime terminology alignment in existing docs (`CONNECTION_STATES.md`, `USAGE-cable.md`)
- ✅ Architecture ADRs for application layer and composition-root wiring added under `docs/`
- ✅ `MODELS_REMODELING_AUDIT.md` refreshed to match current dependency and DTO migration state
- ✅ Removed redundant `ARCHITECTURE-DECISIONS.md` and consolidated ADR navigation in `ARCHITECTURE-INDEX.md`
- ✅ Cleaned stale branch-specific and outdated transport/request references in legacy docs
- ✅ Replaced obsolete absolute-path links in `MODELS_REMODELING_AUDIT.md` with repo-relative links
- ✅ Moved API-only request DTOs (`ContactSearchRequest`, `AccountUpdateRequest`, `InfoPatchRequest`, `PollRequest`) out of `src/models` into `src/api`
- ✅ Removed `qp_send_request+extras.go` by consolidating its helpers into `qp_to_whatsapp_attachment.go`
- ✅ Moved shared attachment hardening implementation from `src/models` to `src/media/attachment_pipeline.go`, keeping `models` aliases/wrappers for compatibility
- ✅ Phase 1: DispatchingHandler decomposition → models/lifecycle_handler.go + models/message_dispatcher.go
- ✅ Test fix: TestSessionServiceWrappersDelegateToServerImplementations (nil DB.Dispatching stub)
- ✅ Phase 0: P7a HasValidHandlers() in WhatsmeowConnection (6 nil-checks eliminated)
- ✅ Phase 4: DispatchPolicy interface + DefaultDispatchPolicy in dispatch/service/
- ✅ Phase 5: SessionIntent enum → replaces StopRequested+DeleteRequested bools (session_intent.go, all tests passing)
- ✅ Phase 1 Cleanup: Removed dead runtime copies (dispatching_handler.go, lifecycle_handler.go, message_dispatcher.go from runtime/) — models versions are canonical
- ✅ Phase 6: QpWhatsappServer decomposition — server_connection.go, server_messaging.go, server_persistence.go; methods removed from qp_whatsapp_server.go; build + all tests passing
- ✅ Transport-boundary cleanup slice: migrated contact search/account/info patch/poll request DTOs from `src/models` to `src/api`; swagger regenerated; api+swagger compile validation passing
- ✅ Attachment helper cleanup slice: consolidated `qp_send_request+extras.go` into `qp_to_whatsapp_attachment.go`; api/apps-form compile validation passing
- ✅ Attachment ownership cleanup slice: `QpToWhatsappAttachment` implementation now lives in `src/media`; `models` keeps compatibility alias/wrappers; `src/media` compile validation passing
- ✅ Multi-module manifest cleanup slice: synced `go.mod`/`go.sum` and local `replace` directives so `models`, `api`, `apps/form`, and `cable` compile checks pass again
- ✅ Application-layer runtime slice: added explicit runtime session entry points for start/stop/restart/send and rerouted selected API call sites through `src/runtime`
- ✅ Application-layer runtime slice extended: moved session option toggles, debug toggle, and configuration patch flag application behind explicit `src/runtime` helpers
- ✅ Application-layer runtime slice extended again: moved create/save/delete session flows behind explicit `src/runtime` helpers; production API no longer calls direct session lifecycle/persistence/service operations for this slice
- ✅ Test setup alignment slice: `src/api/testing_setup.go` now creates in-memory test sessions through `runtime.LoadSessionRecord`, removing the last direct `AppendNewServer` call from API helpers
- ✅ Application-layer runtime slice extended once more: existing-session owner validation/mutation now goes through `runtime.ApplySessionUser`; API no longer mutates live session ownership directly before save
- ✅ Application-layer runtime slice extended again: new-session record assembly now goes through `runtime.BuildSessionRecord`; duplicated `QpServer` construction left the API create handlers
- ✅ API request mapping cleanup slice: `buildSessionConfigurationPatch` centralizes `InfoCreateRequest`/`InfoPatchRequest` to `runtime.SessionConfigurationPatch` mapping, with focused tests passing
- ✅ Application-layer runtime slice extended again: live-session existence checks by token now go through `runtime.FindLiveSessionByToken`; direct map access left `InformationController` and SPA live lookup utils
- ✅ API persisted-record lookup cleanup slice: `findPersistedServerRecord` centralizes DB server-record lookup with case-insensitive fallback and is reused by SPA ownership lookup and conversation-label flows
- ✅ Application-layer runtime lookup slice: API server/session extension helpers now delegate live-session token lookup and first-ready-session lookup through explicit `src/runtime` wrappers instead of calling `models.Get*` helpers directly
- ✅ API persisted-record listing slice: `listPersistedServerRecords` centralizes `DB.Servers.FindAll()` and is reused by SPA read controllers and server-record fallback lookup
- ✅ API user lookup cleanup slice: `findPersistedUser` centralizes `DB.Users.Find()` so request user resolution no longer reaches the users store directly outside the shared helper
- ✅ Non-SPA user service cleanup slice: login, health credential checks, and password update now delegate through shared API user helpers instead of calling `DB.Users.Check/Exists/UpdatePassword` directly
- ✅ Non-SPA restore/runtime slice: restore diagnose/auto/manual controllers now delegate through explicit `src/runtime` wrappers instead of calling `models.WhatsappService` restore methods directly
- ✅ Non-SPA live-session listing slice: `HealthController` now consumes `runtime.ListLiveSessions*`; direct iteration over `models.WhatsappService.Servers` left the API production code and remains only in test setup
- ✅ Non-SPA history download slice: attachment URL prefix resolution now delegates through `runtime.GetSessionDownloadPrefix`; `HistoryDownloadController` no longer calls `models.GetDownloadPrefixFromToken` directly
- ✅ API conversation-label store cleanup slice: `findConversationLabelStore` centralizes conversation-label store access and is reused by controller and v5 label search entry points
- ✅ Runtime persisted-record slice: persisted server-record list/lookup now delegate through `runtime.ListPersistedSessionRecords` and `runtime.FindPersistedSessionRecord`; direct `models` calls disappeared from non-SPA API handlers
- ✅ Runtime user-service slice: persisted user find/auth/update-password helpers now live in `src/runtime`; API user helpers only delegate to runtime
- ✅ Runtime conversation-label store slice: conversation-label store resolution now lives in `src/runtime`; API helper only delegates to runtime
- ✅ Runtime user CRUD slice: `CountPersistedUsers`, `ListPersistedUsers`, `CreatePersistedUser`, `DeletePersistedUser` added to `src/runtime`; SPA admin/read controllers no longer access `DB.Users` directly
- ✅ Runtime dispatching slice: `FindPersistedDispatching` added to `src/runtime`; `CountSPADispatchingForServer` fallback no longer accesses `DB.Dispatching` directly
- ✅ **All direct `models.WhatsappService` access in API production code eliminated**; remaining occurrences are confined to `testing_setup.go` (test infrastructure only)
- ✅ Runtime mcp slice: `tool_list_servers.go`, `tool_health.go`, `mcp_server.go` no longer access `models.WhatsappService` directly; use `runtime.ListLiveSessions` and `runtime.FindLiveSessionByToken`; `mcp/go.mod` updated with runtime require
- ✅ Runtime cable slice: `cable_auth.go` and `cable_commands.go` no longer access `models.WhatsappService` directly; use `runtime.FindPersistedUser`, `runtime.GetLiveSessionByToken`, `runtime.FindPersistedSessionRecord`; `cable/go.mod` updated with runtime replace + require; `go mod tidy` synced; cable and mcp build validation passing (exit 0)
- ✅ Runtime form slice: `form_handlers.go`, `form_account.go`, and `form_extensions.go` no longer access `models.WhatsappService` directly for user auth, user lookup/create, session lookup, or session delete; they now use `runtime.AuthenticateUser`, `runtime.FindPersistedUser`, `runtime.ExistsPersistedUser`, `runtime.CreatePersistedUser`, `runtime.GetLiveSessionByToken`, and `runtime.DeleteSessionRecord`; `apps/form/go.mod` synced with runtime require/replace; compile validation passing
- ✅ Runtime helper slice: `ExistsPersistedUser` and `GetOrCreateLiveSessionByToken` added to `src/runtime`, with focused tests passing; `cable_commands.go` no longer calls `GetOrCreateServerFromToken` directly
- ✅ Runtime health slice: `api_handlers+HealthController.go` no longer checks `models.WhatsappService` directly; it now uses `runtime.IsSessionServiceAvailable`, leaving `src/runtime/session_service.go` as the sole production boundary over `models.WhatsappService`
- ✅ Runtime wrapper validation slice: confirmed `ExistsPersistedUser`, `GetOrCreateLiveSessionByToken`, `CountPersistedUsers`, `CreatePersistedUser`, `DeleteSessionRecord` all present in `src/runtime`; all runtime tests passing (33 tests, exit 0); `apps/form` and `cable` build clean (exit 0); no direct `models.WhatsappService` access remaining in `apps/form` or `cable` production code
- ✅ Whatsmeow blocklist realtime slice: `events.Blocklist` is now registered in `whatsmeow_event_router.go` and mapped to `OnEventBlocklist`, which emits a `WhatsappMessage` through `Follow()` so block/unblock confirmations can flow through the existing dispatch/cable pipeline; `whatsmeow` package build passing
- ✅ Transport wiring slice (P3 groundwork): added `models.TransportServices` + `models.ApplyTransportServices(...)` and migrated `src/main.go` bootstrap from scattered global assignments to a single centralized transport wiring call; full app build passing
- ✅ Transport adapter helper slice (P3 incremental): added `models` helper functions over realtime/RabbitMQ adapter globals (`HasActiveRealtimeConnections`, `ResolveRabbitMQClient`, `GetRabbitMQPublisherClient`, `CloseRabbitMQPublisherClient`, metric/injection helpers) and updated core call sites/tests to consume helpers instead of globals directly; focused models tests + full app build passing
- ✅ Transport helper closure slice (P3 incremental): remaining direct `models` call sites now use helper functions for lifecycle publishing and RabbitMQ error metrics as well; direct global usage is effectively confined to helper/wiring boundaries plus tests; focused models tests + full app build passing
- ✅ Per-handler lifecycle publisher slice (P3 incremental): `DispatchingHandler` now carries an injectable `lifecyclePublisher` with global fallback compatibility; `HandlerEnsure()` injects the default publisher and lifecycle tests now validate instance-level injection instead of mutating the global publisher; focused models tests + full app build passing
- ✅ Per-instance RabbitMQ client resolver slice (P3 incremental): `QpRabbitMQConfig` now carries an optional `clientResolver` field (injected via `WithClientResolver()`); `PublishMessage` uses `resolveClient()` accessor with global fallback; 2 focused publish tests passing; full models tests + build passing
- ✅ Phase 6 dispatching extraction slice: RabbitMQ config + dispatching query methods (`GetRabbitMQConfig*`, `GetDispatching*`, `GetWebhooks`, `GetWebhookDispatchings`, `HasWebhooks`, `HasRabbitMQConfigs`, `InitializeRabbitMQConnections`) extracted from `qp_whatsapp_server.go` into `server_dispatching.go`; `qp_whatsapp_server.go` reduced from 375 → 232 lines; `strings` import removed from core file; models tests + full build passing

**Layer 3b Implementation Details:**
- Updated `src/api/api_handlers+SPAMessageController.go`:
  - Renamed all 13 call sites from respondSPAServerLookupError → respondSPASessionLookupError
  - Renamed all 6 call sites from respondSPAServerReadyError → respondSPASessionReadyError
  - Added backward-compatible aliases: respondSPAServer* now delegate to respondSPASession*
  - Maintains full backward compatibility for any external callers

**Validation Results:**
- ✅ All 8 session-related tests passing
- ✅ Full project compilation successful (no errors)
- ✅ Backward compatibility maintained throughout migration
- ✅ Semantic clarity improved - session terminology now spans layers 1-3b

## Next Steps

1. **Phase 0 Quick Wins - ✅ COMPLETE:**
   - [x] P7a: `HasValidHandlers()` added to `WhatsmeowConnection` (6 repetitions of nil-check eliminated)
   - [x] P7b: `IsValidForDispatch()` was already in `models/` (no action needed)
   - [x] P7c: `InitializeCacheService()` was already the single entry point (no action needed)

2. **Phase 4 (DispatchPolicy) - ✅ COMPLETE:**
   - [x] Define `DispatchPolicy` interface in `dispatch/service/dispatch_policy.go`
   - [x] Implement `DefaultDispatchPolicy` with current filtering logic (moved from `shouldDispatchToTarget`)
   - [x] Wire `Policy DispatchPolicy` field in `DispatchService` singleton  
   - [x] Remove `shouldDispatchToTarget` from `dispatch_service.go`
   - [x] All tests passing, build successful

3. **Remaining Phases:**
   - Phase 2: Transport Injection via ServiceContainer (P3) — deferred, globals work well
   - Phase 5: Explicit State Machine (P6) — ✅ COMPLETE (SessionIntent enum, all tests passing)
   - Phase 6: Decompose QpWhatsappServer (P2) — highest risk, do last
  - ✅ Phase B: Pairing use cases — `GetSessionPairingQRCode` and `PairSessionWithPhone` added to `src/runtime/session_service.go`; `SPAServerQRCodeController` and `SPAServerPairCodeController` in `src/api/api_handlers+SPAReadController.go` now delegate through runtime wrappers; `buildSPAPairing` helper removed and replaced by `parseSPAHistorySyncDays`; build and all runtime tests passing (exit 0)
  - ✅ Message Reactions feature — `SendReaction` added to `IWhatsappConnection` interface and `WhatsmeowConnection`; `POST /messages/react` endpoint via `SendReactionController`; Swagger regenerated; all tests passing
  - ✅ Block/Unblock Contacts — `BlockContact`/`UnblockContact` added to `IWhatsappContactManager` interface, `WhatsmeowContactManager`, `WhatsmeowStoreContactManager` (stub), and `QpContactManager`; `POST /contacts/block` + `DELETE /contacts/block` via `BlockContactController`/`UnblockContactController`; Swagger regenerated; build passing
  - ✅ Blocklist realtime propagation slice — block/unblock confirmations now produce handled system events instead of falling back to unhandled-debug logging; routed through `Follow()` for existing dispatch/realtime consumers
  - ✅ Phase 2 groundwork: transport adapter globals are now configured through `models.ApplyTransportServices(models.TransportServices{...})`; remaining P3 work is incremental migration away from the underlying globals, not bootstrap duplication cleanup
  - ✅ Phase 2 incremental adapter slice: models call sites now go through transport helper functions instead of touching adapter globals directly in several core paths (presence lookup, RabbitMQ resolver/client close/injection/metrics)
  - ✅ Phase 2 incremental adapter closure: lifecycle publish and RabbitMQ publish-error paths also moved behind helper functions, leaving transport globals mostly as implementation detail of the helper/wiring boundary
  - All major slices for this branch are complete. `src/runtime/session_service.go` is the only remaining file with intentional direct access to `models.WhatsappService`, acting as the canonical application-layer boundary.

## Immutable Constraints

1. **Backward Compatibility:** No breaking changes to existing server-named functions or API contracts
2. **Type Alias Pattern:** Zero-cost renaming via Go type aliases (no performance penalty)
3. **Wrapper Delegation:** Each wrapper delegates entirely to existing implementation
4. **Test-Driven Validation:** Each layer includes tests proving delegation correctness
5. **Semantic Clarity:** Only domain objects (QpWhatsappServer/Session) affected; HTTP server infrastructure (webserver, signalr, cable) retains "server" naming
6. **Incremental Progress:** One layer completed and validated before proceeding to next

## Progress Tracking

**Message Count:** Session initiated at approx. message 1 of long conversation

**Session Memory:** Initial plan and findings stored in conversation summary

**Key Files Modified/Created:**
- src/api/api_spa_session_utils.go ✅
- src/api/api_spa_session_utils_test.go ✅
- src/models/qp_whatsapp_session.go ✅
- src/models/qp_whatsapp_session.go ✅
- src/whatsmeow/whatsmeow_handlers.go ✅ (events instrumentation: connect failure, stream error, temporary ban)
- src/whatsmeow/whatsmeow_handlers_events_blocklist.go ✅ (events instrumentation: blocklist updated)
- src/models/qp_whatsapp_session_dispatching.go ✅
- src/models/qp_whatsapp_session_test.go ✅
- src/models/qp_whatsapp_session_dispatching_test.go ✅
- src/api/api_session_extensions.go ✅
- src/api/api_session_extensions_test.go ✅
- src/api/api_handlers+SPASessionController.go ✅
- src/api/api_routes_sessions.go ✅
- docs/PLAN-ARCHITECTURE-REFACTORING.md ✅
- docs/ARCHITECTURE-CURRENT-STATE.md ✅
- docs/ARCHITECTURE-INDEX.md ✅
- docs/ARCHITECTURE-TARGET-STATE.md ✅
- docs/ARCHITECTURE-ROADMAP.md ✅
- docs/ARCHITECTURE-EXECUTION-CHECKLIST.md ✅
- docs/ADR-0001-MODULAR-MONOLITH-INCREMENTAL-REFACTORING.md ✅
- docs/ADR-0002-SESSION-AS-RUNTIME-CONCEPT.md ✅
- docs/ADR-0003-MODELS-IS-NOT-THE-ESCAPE-HATCH.md ✅
- docs/ADR-0004-EXPLICIT-APPLICATION-LAYER.md ✅
- docs/ADR-0005-GROUPED-COMPOSITION-ROOT-WIRING.md ✅
- docs/ARCHITECTURE-PACKAGE-MAP.md ✅
- docs/CONNECTION_STATES.md ✅
- docs/MODELS_REMODELING_AUDIT.md ✅
- src/api/contact_search_request.go ✅
- src/api/account_update_request.go ✅
- src/api/info_patch_request.go ✅
- src/api/poll_request.go ✅
- docs/CONTACT_MESSAGES.md ✅
- docs/PLAN-ARCHITECTURE-REFACTORING.md ✅
- docs/SEND_LOCATION.md ✅
- docs/USAGE-cable.md ✅
- src/media/attachment_pipeline.go ✅
- src/models/qp_to_whatsapp_attachment_alias.go ✅
- src/runtime/session_service.go ✅
- src/runtime/session_service_test.go ✅
