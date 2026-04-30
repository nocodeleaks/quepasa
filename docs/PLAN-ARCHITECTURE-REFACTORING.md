# Architecture Refactoring Plan — QuePasa + Whatsmeow

> Generated: 2026-04-29  
> Branch: develop  
> Status: Draft

---

## Context

Deep architectural analysis of the QuePasa + Whatsmeow codebase revealed a set of structural issues that progressively increase coupling, reduce testability, and make the system harder to evolve. This document organizes findings into actionable work items, prioritized by impact and risk.

An additional strategic goal for this refactoring is to migrate the central runtime terminology from `server` to `session`, reflecting the actual responsibility of the object lifecycle around a WhatsApp connected identity rather than a generic infrastructure server concept.

---

## What Is Already Good (Do Not Break)

- `src/dispatch/` — Clean transport separation (HTTP, RabbitMQ, realtime) with well-defined contracts (`Target`, `OutboundRequest`, `RealtimePublisher`). Keep this structure as a reference.
- `src/cache/` — Strategy pattern with interchangeable backends (memory/disk/redis). Initialization via `InitializeCacheService()` is sound.
- `src/whatsapp/` — Domain interfaces (`IWhatsappConnection`, `IWhatsappHandlers`, `WhatsappMessage`) are clean and decoupled.

---

## Problem Inventory

### P1 — DispatchingHandler: 4 Responsibilities in 600 Lines

**File**: `src/models/dispatching_handler.go`

**Current responsibilities** (all mixed in one struct):

1. Message cache management (`QpWhatsappMessages`)
2. Event handling (`Message`, `Receipt`)
3. Lifecycle events (`OnConnected`, `OnDisconnected`, `LoggedOut`)
4. Subscriber orchestration (`Trigger`, `Register`, `aeh []QpDispatchingHandlerInterface`)

**Impact**: Cannot test any responsibility in isolation. Adding new event types or transports requires touching a 600-line file.

**Proposed split**:

```text
models/dispatching_handler.go         ← thin orchestrator (keep struct, reduce lines)
models/message_cache_handler.go       ← QpWhatsappMessages operations
runtime/message_dispatcher.go         ← Trigger + subscriber orchestration
runtime/lifecycle_handler.go          ← OnConnected, OnDisconnected, LoggedOut
```

---

### P2 — QpWhatsappServer: God Object

**File**: `src/models/qp_whatsapp_server.go`

**Symptoms**:

- 17 public fields + 6 private fields
- 50+ methods
- Mixes: persistence, runtime state, connection lifecycle, configuration, message routing

**Naming intent**:

- `QpWhatsappServer` should become a `Session`-oriented abstraction
- The rename should happen intentionally, not as a blind search/replace
- Naming migration must preserve backward compatibility during the transition when required by public/internal contracts

**Proposed decomposition**:

```text
models/qp_whatsapp_session.go         ← core identity fields (Token, Verified, etc.)
models/session_connection_manager.go  ← connection start/stop/update logic
models/session_lifecycle.go           ← Initialize, Start, Stop, Delete
```

Do not decompose all at once — extract `session_connection_manager.go` first (highest coupling surface), while keeping a temporary compatibility layer around the old `server` naming until all major call sites are migrated.

---

### P3 — 13 Global Variables for Transport Injection

**File**: `src/main.go` (lines ~92–108)

**Current pattern**:

```go
models.GlobalRealtimePresenceChecker = signalr.SignalRHub
models.GlobalDispatchingLifecyclePublisher = runtime.NewDispatchingLifecyclePublisher()
models.GlobalRabbitMQGetClient = func(...) { return rabbitmq.GetRabbitMQClient(...) }
// ... 10 more globals
```

**Problems**:

- Not thread-safe (no mutex)
- Silent NoOp fallback hides initialization failures
- Impossible to test components in isolation

**Target pattern**:

```go
type TransportServices struct {
    RabbitMQFactory    models.RabbitMQClientFactory
    RealtimeChecker    models.RealtimePresenceChecker
    LifecyclePublisher models.DispatchingLifecyclePublisher
    // ...
}

session := models.NewQpWhatsappSession(token, db, TransportServices{...})
```

Migrate globals one by one — do not replace all at once.

---

### P4 — WhatsmeowHandlers.EventsHandler: 38-Case Switch

**File**: `src/whatsmeow/whatsmeow_handlers.go`

**Current state**: One function with a 240-line type-switch dispatching 38 different event types. Mixes connection events, messages, calls, history sync, etc.

**Problem**: Single Responsibility violation. Every new event type expands the same function.

**Target pattern** — Event Router:

```go
type EventRouter struct {
    handlers map[reflect.Type]func(interface{})
}

func (r *EventRouter) Register(evt interface{}, fn func(interface{})) {
    r.handlers[reflect.TypeOf(evt)] = fn
}

func (r *EventRouter) Dispatch(evt interface{}) {
    if fn, ok := r.handlers[reflect.TypeOf(evt)]; ok {
        fn(evt)
    }
}
```

Register handlers in `NewWhatsmeowHandlers()`:

```go
router.Register(events.Message{}, handler.onMessage)
router.Register(events.Receipt{}, handler.onReceipt)
router.Register(events.Connected{}, handler.onConnected)
```

---

### P5 — shouldDispatchToTarget: Business Logic in Transport Layer

**File**: `src/dispatch/service/dispatch_service.go` (lines ~167–195)

**Problem**: Filtering rules (which messages to send, based on groups/receipts/etc.) live inside the transport module, which should only know *how* to send, not *whether* to send.

**Target**: Move filtering to `runtime/` as a `DispatchPolicy`:

```go
// runtime/dispatch_policy.go
type DispatchPolicy interface {
    ShouldDispatch(target dispatch.Target, message *whatsapp.WhatsappMessage) (bool, string)
}

type DefaultDispatchPolicy struct{}

func (p *DefaultDispatchPolicy) ShouldDispatch(target, message) (bool, string) {
    if message.FromGroup() && !target.GetGroups() { return false, "groups filtered" }
    // ...
}
```

---

### P6 — Implicit State Machine in GetState()

**File**: `src/models/qp_whatsapp_server.go`

**Current**: Session state is currently calculated inside the `QpWhatsappServer` implementation by combining boolean flags (`DeleteRequested`, `StopRequested`, `Verified`, `Reconnect`). Invalid transitions are not caught.

**Target**: Explicit session state machine with validated transitions:

```go
type SessionState int

const (
    StateUnverified SessionState = iota
    StateUnprepared
    StateConnecting
    StateConnected
    StateStopping
    StateStopped
    StateDeleting
)

type SessionStateMachine struct {
    current           SessionState
    validTransitions  map[SessionState][]SessionState
    mu                sync.RWMutex
}

func (sm *SessionStateMachine) TransitionTo(next SessionState) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    for _, valid := range sm.validTransitions[sm.current] {
        if valid == next {
            sm.current = next
            return nil
        }
    }
    return fmt.Errorf("invalid transition %v → %v", sm.current, next)
}
```

---

### P7 — Minor Issues (Quick Wins)

| ID | Location | Issue | Fix |
| --- | --- | --- | --- |
| P7a | `whatsmeow_connection.go` | Null handler check repeated 6× | Extract `HasValidHandlers() bool` helper |
| P7b | `whatsapp/whatsapp_message_extensions.go` | Business logic (`IsValidForDispatch`) in domain package | Move to `models/` or `runtime/` |
| P7c | `models/cache_initialization.go` | Cache backend injected manually in 3+ places | Single injection point in service factory |
| P7d | `runtime/` directory | Nearly empty — exists in name only | Populate with handlers extracted from P1 |

---

## Execution Plan

### Phase 0 — Quick Wins (no risk, immediate value)

- [x] P7a: Extract `HasValidHandlers()` in `whatsmeow_connection.go` (COMPLETED)
- [x] P7b: Move `IsValidForDispatch()` to `models/` (ALREADY DONE — was in models all along)
- [x] P7c: Consolidate cache injection into a single call site (ALREADY STRUCTURED — `InitializeCacheService()`)

Estimated scope: 3–5 files, no interface changes.

---

### Phase 1 — Split DispatchingHandler (P1) — COMPLETED

1. ~~Create `runtime/lifecycle_handler.go` — move `OnConnected`, `OnDisconnected`, `LoggedOut`~~ → `models/lifecycle_handler.go` (in models to avoid circular import from runtime→whatsmeow→models)
2. ~~Create `runtime/message_dispatcher.go` — move `Trigger`, `Register`, subscriber loop~~ → `models/message_dispatcher.go` (same circular import constraint)
3. ~~Slim `models/dispatching_handler.go` to cache + event entry points only~~ → now delegates lifecycle to `LifecycleHandler` and dispatch to `MessageDispatcher`
4. ~~Update all callers (whatsmeow_handlers.go, tests)~~ → tests updated
5. Dead runtime copies (`runtime/dispatching_handler.go`, `runtime/lifecycle_handler.go`, `runtime/message_dispatcher.go`) removed — models versions are canonical

---

### Phase 2 — Transport Injection via ServiceContainer (P3)

1. Define `TransportServices` struct in `models/`
2. Add `NewQpWhatsappSession(token, db, TransportServices)` constructor
3. Keep a temporary compatibility constructor or alias for `QpWhatsappServer` where necessary
4. Migrate globals one by one (start with `GlobalRabbitMQGetClient`)
5. Remove global variables as each migration is verified

**Risk**: Low per variable, medium total — can be done incrementally across multiple PRs.

---

### Phase 3 — EventRouter in WhatsmeowHandlers (P4) — COMPLETED

1. ~~Implement `EventRouter` in `whatsmeow/event_router.go`~~ → `whatsmeow_event_router.go` with `EventRouter` struct + `buildRouter()` on `WhatsmeowHandlers`
2. ~~Register all handlers in `NewWhatsmeowHandlers()`~~ → handlers registered lazily via `getRouter()` on first `EventsHandler` call
3. ~~Replace switch body with `router.Dispatch(evt)`~~ → `EventsHandler` now: guard → `getRouter().Dispatch(rawEvt)` → default fallback
4. ~~Delete old switch~~ → Removed; complex case bodies extracted into dedicated methods:
   - `onConnectedEvent()`, `onConnectFailureEvent()`, `onStreamErrorEvent()`
   - `onTemporaryBanEvent()`, `onAppStateSyncCompleteEvent()`
5. Added `hasWAHandlers()` helper on `WhatsmeowHandlers` (eliminates 6 repetitions of WAHandlers nil-check)

---

### Phase 4 — DispatchPolicy in Runtime (P5) — COMPLETED

1. ~~Define `DispatchPolicy` interface in `runtime/`~~ → Defined in `dispatch/service/dispatch_policy.go` (avoids circular import)
2. ~~Implement `DefaultDispatchPolicy` with current filtering logic~~ → `DefaultDispatchPolicy` in `dispatch/service/dispatch_policy.go`
3. ~~Inject policy into `dispatch.DispatchService`~~ → `Policy DispatchPolicy` field, wired in singleton constructor
4. ~~Remove `shouldDispatchToTarget` from `dispatch_service.go`~~ → Removed, replaced with `service.Policy.ShouldDispatch(...)`

Note: `DispatchPolicy` interface was placed in `dispatch/service/` instead of `runtime/` to avoid
a circular dependency (`dispatch → runtime → dispatch`). Runtime can still implement custom
policies using the interface from `dispatch/service/` and inject them into the singleton.

---

### Phase 5 — Explicit State Machine (P6) — COMPLETED

1. ~~Define `SessionState` type and constants~~ → Defined `SessionIntent` enum in `models/session_intent.go` with `None`, `Stop`, `Delete` values
2. ~~Implement `SessionStateMachine` with transition table~~ → `SessionIntent.IsStopRequested()` and `IsDeleteRequested()` provide clean predicate API
3. ~~Replace boolean flags with state machine calls in the session abstraction~~ → `StopRequested bool` + `DeleteRequested bool` removed; single `Intent SessionIntent` field added to `QpWhatsappServer`
4. ~~Update `GetState()` to read from state machine~~ → `GetState()` now uses `server.Intent.IsDeleteRequested()` and `server.Intent.IsStopRequested()`
5. Delete/Stop/Start lifecycle methods updated; test assertions updated; all tests passing

---

### Phase 6 — Decompose QpWhatsappServer (P2)

1. Introduce `QpWhatsappSession` as the primary naming target for the central lifecycle abstraction
2. Extract `session_connection_manager.go` (connection start/stop/update)
3. Extract `session_lifecycle.go` (Initialize, Start, Stop, Delete)
4. Keep the legacy `qp_whatsapp_server.go` surface only as a thin compatibility layer during migration
5. Remove the compatibility layer once call sites, tests, and API naming are stable

**Risk**: High — most widespread type in the codebase. Do last, after other phases stabilize the boundaries.

---

## Naming Migration Strategy — Server → Session

The rename from `server` to `session` should be treated as a domain correction, not just a cosmetic cleanup.

### Goals

- Align the main lifecycle abstraction with the actual business meaning: one WhatsApp connected identity/session
- Reduce confusion between application process/server concerns and per-connection runtime concerns
- Prepare the codebase for multiple sessions per process without naming ambiguity

### Scope Candidates

- `QpWhatsappServer` → `QpWhatsappSession`
- `qp_whatsapp_server.go` → `qp_whatsapp_session.go`
- `server_connection_manager.go` → `session_connection_manager.go`
- `server_lifecycle.go` → `session_lifecycle.go`
- Local variables like `server` → `session` where they refer to the domain object, not HTTP server infrastructure

### Migration Rules

- Do not rename everything in one pass
- Rename the domain core first, adapters and transports second, API payload naming last
- Preserve compatibility where exported types or widely referenced constructors would otherwise break too much at once
- Distinguish carefully between WhatsApp session naming and real web server/runtime infrastructure naming

### Exit Criteria

- Core models and runtime use `session` terminology consistently
- Remaining `server` references are limited to actual infrastructure server concerns or explicit compatibility shims
- No exported compatibility alias remains without a reason documented in code

---

## Constraints

- Do NOT change external API contracts (REST endpoints, webhook payloads, RabbitMQ message format).
- Do NOT rename `WhatsappMessage`, `IWhatsappConnection`, `IWhatsappHandlers` — widely used.
- Do NOT rename HTTP/webserver infrastructure symbols that are genuinely about server concerns.
- Each phase must compile and pass all tests before starting the next.
- `AGENTS.md` must be updated on each branch created for a phase.
- QpVersion must be incremented on merge to main.

---

## References

- `src/models/dispatching_handler.go`
- `src/models/qp_whatsapp_server.go`
- `src/whatsmeow/whatsmeow_handlers.go`
- `src/whatsmeow/whatsmeow_connection.go`
- `src/dispatch/service/dispatch_service.go`
- `src/runtime/`
- `src/main.go`
- [CODE_ORGANIZATION.md](CODE_ORGANIZATION.md)
