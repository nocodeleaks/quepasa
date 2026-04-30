# Task: QuePasa Architecture Refactoring - Server→Session Naming Migration

## Task Objective

Perform a phased server→session naming migration across the QuePasa codebase to better reflect that QpWhatsappServer represents a per-connection WhatsApp identity (session), not infrastructure. Implement this through backward-compatible type aliases, wrapper functions, and gradual controller/utility layer migration, culminating in a comprehensive refactoring plan for remaining architectural issues.

## Mandatory Checklist

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
- ✅ Phase 1: DispatchingHandler decomposition → models/lifecycle_handler.go + models/message_dispatcher.go
- ✅ Test fix: TestSessionServiceWrappersDelegateToServerImplementations (nil DB.Dispatching stub)
- ✅ Phase 0: P7a HasValidHandlers() in WhatsmeowConnection (6 nil-checks eliminated)
- ✅ Phase 4: DispatchPolicy interface + DefaultDispatchPolicy in dispatch/service/
- ✅ Phase 5: SessionIntent enum → replaces StopRequested+DeleteRequested bools (session_intent.go, all tests passing)
- ✅ Phase 1 Cleanup: Removed dead runtime copies (dispatching_handler.go, lifecycle_handler.go, message_dispatcher.go from runtime/) — models versions are canonical
- ✅ Phase 6: QpWhatsappServer decomposition — server_connection.go, server_messaging.go, server_persistence.go; methods removed from qp_whatsapp_server.go; build + all tests passing
- ✅ Transport-boundary cleanup slice: migrated contact search/account/info patch/poll request DTOs from `src/models` to `src/api`; swagger regenerated; api+swagger compile validation passing

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
