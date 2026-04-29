# QuePasa + Whatsmeow: Architectural Analysis

## Overview

This document captures a deep structural analysis of the QuePasa + Whatsmeow integration,
identifying risks, technical debt, and architectural violations with actionable recommendations.

---

## Issue 1 — Duplicated `DispatchingHandler` (High Risk)

Three parallel implementations exist with similar names and overlapping responsibilities:

| Type | Package | Status |
|---|---|---|
| `DispatchingHandler` | `models` | Active — used via `server.Handler` |
| `QPDispatchingHandler` | `models` | Marked DEPRECATED in comments |
| `DispatchingHandler` | `runtime` | Intended replacement (new version) |

The functions `dispatchOutboundFromServer` and `dispatchOutboundToTargets` are **duplicated** in both
`models/qp_dispatching_handler.go` and `runtime/dispatching_handler.go`. This creates a real risk
of behavioral divergence between code paths.

Additionally, the field `server.DispatchingHandler` (`*QPDispatchingHandler`) is declared in
`QpWhatsappServer` and assigned (line 1161), but **never called anywhere** — dead code.

**Recommendation:**
- Remove `QPDispatchingHandler` from `models`.
- Remove duplicated `dispatchOutboundFromServer` / `dispatchOutboundToTargets` from `models`.
- Consolidate all dispatching business logic exclusively in `runtime.DispatchingHandler`.

**Affected files:**
- `src/models/qp_dispatching_handler.go`
- `src/models/qp_whatsapp_server.go` (field `DispatchingHandler`)
- `src/runtime/dispatching_handler.go`

---

## Issue 2 — `models.DispatchingHandler` is a God Object (High Risk)

`src/models/dispatching_handler.go` accumulates responsibilities that violate the stated architecture:

- Message cache (`appendMsgToCache`, `GetById`, `Count`, `GetByTime`)
- Dispatching lifecycle (`Trigger`, `Receipt`) — imports `dispatch/service`
- Connection lifecycle (`LoggedOut`, `OnConnected`, `OnDisconnected`)
- Message filtering (`HandleGroups`, `HandleBroadcasts`)

The `models` package importing `dispatch/service` directly violates the architectural rule:
> *"Modules that should NEVER import dispatch directly: models"* — copilot-instructions.md

This creates tight coupling between the domain layer and transport mechanisms.

**Recommendation:**
- Keep `DispatchingHandler` in `models` responsible only for **cache orchestration**.
- Move lifecycle event publishing (`LoggedOut`, `OnConnected`) to the `runtime` module.
- Remove the `dispatch/service` import from `models`.

**Affected files:**
- `src/models/dispatching_handler.go`

---

## Issue 3 — Module Boundary Violations (Medium Risk)

`src/models/qp_whatsapp_server.go` directly imports transport packages:

```go
import (
    rabbitmq "github.com/nocodeleaks/quepasa/rabbitmq"
    signalr  "github.com/nocodeleaks/quepasa/signalr"
)
```

This couples the domain layer to concrete transport implementations, preventing independent
evolution of transport modules.

**Recommendation:**
- Pass `RealtimePublisher` and RabbitMQ client as interfaces via dependency injection.
- The `models` package should only reference transport through interfaces defined in `whatsapp` or `dispatch`.

**Affected files:**
- `src/models/qp_whatsapp_server.go`

---

## Issue 4 — Race Condition on `IsConnecting` (Medium Risk)

In `src/whatsmeow/whatsmeow_connection.go`:

```go
IsConnecting bool `json:"isconnecting"` // used to avoid multiple connection attempts
```

This field is read and written from multiple goroutines without atomic protection or mutex coverage.
The `syncConnection` mutex exists in `QpWhatsappServer` but is **not present** in `WhatsmeowConnection`.

**Recommendation:**
- Replace with `atomic.Bool` (available since Go 1.19) for lock-free safe access:

```go
isConnecting atomic.Bool
```

**Affected files:**
- `src/whatsmeow/whatsmeow_connection.go`

---

## Issue 5 — Fragile Timestamp Sequencing (Medium Risk)

In `src/whatsmeow/whatsmeow_handlers.go`, `getTimestamp()`:

```go
sequence := atomic.AddUint64(&handler.eventSequence, 1)
return now.Add(time.Duration(sequence%1000000) * time.Nanosecond)
```

The `% 1000000` modulo causes the nanosecond offset to **wrap back to zero** every 1M events,
producing real timestamp inversions. Additionally, injecting fake nanoseconds into real timestamps
can confuse downstream systems that use time for deduplication or ordering.

**Recommendation:**
- Keep `EventSequence uint64` as a separate field on `WhatsappMessage`.
- Do not alter the real timestamp.
- Let consumers sort by `(Timestamp, Sequence)` when ordering is required.

**Affected files:**
- `src/whatsmeow/whatsmeow_handlers.go`
- `src/whatsapp/` (add `Sequence` field to `WhatsappMessage`)

---

## Issue 6 — `Receipt` Uses the Same Trigger as New Messages (Technical Debt)

In `src/models/dispatching_handler.go`:

```go
func (source *DispatchingHandler) Receipt(msg *whatsapp.WhatsappMessage) {
    // should implement a better method for that !!!!
    // should implement a better method for that !!!!
    // (×5 identical comments)
    source.Trigger(msg)
}
```

Read/delivery receipts are dispatched via the same `Trigger` path as new inbound messages.
Webhook consumers that do not filter by message type may reprocess receipts as new messages.

**Recommendation:**
- Implement a dedicated `TriggerReceipt` method.
- Ensure the `WhatsappMessage` payload for receipts carries a distinct, recognizable type before dispatch.

**Affected files:**
- `src/models/dispatching_handler.go`

---

## Issue 7 — Triple Coupling in `WhatsmeowHandlers` (Design Concern)

`WhatsmeowHandlers` embeds three overlapping option sources:

```go
type WhatsmeowHandlers struct {
    WhatsmeowOptions               // value embed
    *WhatsmeowConnection           // pointer embed
    *whatsapp.WhatsappOptions      // pointer embed
    ...
}
```

Two option sets (`WhatsmeowOptions` and `WhatsappOptions`) have overlapping fields with a merge
function `GetServiceOptions()` that computes effective values at call time. The precedence rules
are non-obvious and spread across multiple files.

**Recommendation:**
- Compute a single `EffectiveOptions` struct at handler construction time.
- Remove runtime merging from hot paths.

**Affected files:**
- `src/whatsmeow/whatsmeow_handlers.go`
- `src/whatsmeow/whatsmeow_options.go`

---

## Issue 8 — Global Singletons Limit Testability (Low Risk / Long Term)

Multiple global singletons make unit testing without real database connections difficult:

```go
var WhatsappService  *QPWhatsappService
var WhatsmeowService *WhatsmeowServiceModel
```

Tests in `qp_whatsapp_server_delete_test.go` require real DB setup as a consequence.

**Recommendation:**
- Interfaces already exist (`QpDataServersInterface`, etc.) — use them through constructors.
- Expose factory functions that accept interface parameters for test injection.

---

## Priority Summary

| Priority | Issue | Effort |
|---|---|---|
| High | Remove `QPDispatchingHandler` and duplications | Medium |
| High | Extract `dispatch/service` from `models` package | Medium |
| Medium | Fix `IsConnecting` with `atomic.Bool` | Low |
| Medium | Fix timestamp with separate sequence field | Low |
| Medium | Implement dedicated `TriggerReceipt` | Medium |
| Low | Consolidate `WhatsappOptions` into `EffectiveOptions` | High |
| Low | Dependency injection for singletons | High |

---

## Message Flow Reference (as-is)

```
WhatsApp Event
    └─► WhatsmeowHandlers.EventsHandler()
            └─► (per event type) HandleTextMessage / HandleImageMessage / etc.
                    └─► models.DispatchingHandler.Message()
                            ├─► appendMsgToCache()
                            └─► Trigger()
                                    ├─► runtime.DispatchingHandler.HandleDispatching()
                                    │       └─► dispatch/service.DispatchOutbound()
                                    │               ├─► SendWebhook()
                                    │               └─► PublishRabbitMQ()
                                    └─► dispatch/service.PublishRealtime()
                                            ├─► cable (WebSocket)
                                            └─► signalr
```

---

## Execution Plan

Each task is self-contained and safe to execute independently.
Tasks within the same phase can be done in any order. Complete one phase before starting the next.

---

### Phase 1 — Quick Wins (Zero Functional Risk)

#### Task 1.1 — Fix `IsConnecting` race condition
**File:** `src/whatsmeow/whatsmeow_connection.go`

Replace the exported `bool` field with an unexported `atomic.Bool`:

```go
// Before
IsConnecting bool `json:"isconnecting"`

// After
isConnecting atomic.Bool
```

Update all read/write sites:
- `source.IsConnecting = true` → `source.isConnecting.Store(true)`
- `source.IsConnecting = false` → `source.isConnecting.Store(false)`
- `if source.IsConnecting {` → `if source.isConnecting.Load() {`

Check if `IsConnecting` is exposed via JSON in any API response. If so, add a getter method:
```go
func (source *WhatsmeowConnection) GetIsConnecting() bool {
    return source.isConnecting.Load()
}
```

**Test:** Build and run — no behavior change expected.

---

#### Task 1.2 — Fix timestamp wrap in `getTimestamp()`
**File:** `src/whatsmeow/whatsmeow_handlers.go`

The `% 1000000` causes ordering inversions after 1M events. Fix by removing the modulo:

```go
// Before
return now.Add(time.Duration(sequence%1000000) * time.Nanosecond)

// After
return now.Add(time.Duration(sequence%1000) * time.Nanosecond)
```

Use `% 1000` (max 1μs offset per event, never overflows to next millisecond) as the minimal
safe fix. Long-term goal is to add an explicit `Sequence` field to `WhatsappMessage` instead.

**Test:** Build only — observable only under sustained high-frequency event load.

---

#### Task 1.3 — Remove dead field `DispatchingHandler *QPDispatchingHandler`
**File:** `src/models/qp_whatsapp_server.go`

The field `DispatchingHandler *QPDispatchingHandler` (line ~32) is assigned in two places but
the assigned value (`QPDispatchingHandler`) has its `HandleDispatching` never called from
outside — the live path uses `runtime.DispatchingHandler` directly.

Steps:
1. Confirm by searching all callers of `server.DispatchingHandler.` — currently zero.
2. Search all references to `source.DispatchingHandler` (lines ~1151, ~1161) — these are the
   only two: assignment in `EnsureUnderlying` pattern and the struct field declaration.
3. Remove the field declaration from `QpWhatsappServer`.
4. Remove the assignment block (lines ~1151–1163) in `qp_whatsapp_server.go`.
5. Also remove the `QPDispatchingHandler` registration in `Start()` and `EnsureReady()`:
   ```go
   // Remove this line from Start() and EnsureReady():
   source.Handler.Register(source.DispatchingHandler)
   ```
   Replace with the equivalent `runtime.DispatchingHandler` registration if not already present.

**Test:** Full build + manual QA of message dispatch (send a message, verify webhook fires).

---

### Phase 2 — Remove `QPDispatchingHandler` Entirely

#### Task 2.1 — Audit `QPDispatchingHandler` remaining usages
**File:** `src/models/qp_dispatching_handler.go`

Before deleting, run a full search for `QPDispatchingHandler` across the codebase.

Currently known usages:
- `UpdateConnection()` in `qp_whatsapp_server.go`: creates a new `QPDispatchingHandler` and
  calls `source.Handler.Register(dispatchingHandler)`.

This `Register` call is what actually hooks `QPDispatchingHandler.HandleDispatching` into the
`models.DispatchingHandler.aeh` slice (appended events handler). This is the live dispatch path.

Steps:
1. In `UpdateConnection()`, replace:
   ```go
   dispatchingHandler := &QPDispatchingHandler{server: source}
   if !source.Handler.IsAttached() {
       source.Handler.Register(dispatchingHandler)
   }
   ```
   With a `runtime.DispatchingHandler` registration (which already has `HandleDispatching`):
   ```go
   if !source.Handler.IsAttached() {
       runtimeHandler := runtime.NewDispatchingHandler(source)
       source.Handler.Register(runtimeHandler)
   }
   ```
2. Ensure `runtime.DispatchingHandler` implements `QpDispatchingHandlerInterface`.
3. Delete `src/models/qp_dispatching_handler.go`.
4. Remove the `DispatchingHandler *QPDispatchingHandler` field from `QpWhatsappServer` (if not
   already done in Task 1.3).

**Test:** Full build + send/receive messages + verify webhook delivery.

---

#### Task 2.2 — Remove duplicated `dispatchOutboundFromServer` / `dispatchOutboundToTargets` from `models`
**Files:**
- `src/models/qp_dispatching_handler.go` (will be deleted in 2.1)
- `src/models/qp_whatsapp_server_extensions.go`

After Task 2.1, the functions `dispatchOutboundFromServer` and `dispatchOutboundToTargets`
in `models` become dead code (they were only called by `QPDispatchingHandler`).

Also update these wrapper functions in `qp_whatsapp_server_extensions.go`:
```go
// PostToDispatchingFromServer — currently calls models.dispatchOutboundFromServer
// After: delegate to runtime package instead
func PostToDispatchingFromServer(server *QpWhatsappServer, message *whatsapp.WhatsappMessage) error {
    return runtime.DispatchOutboundFromServer(server, message)  // expose from runtime
}
```

Or keep the wrappers but make them call `runtime` internals. Either way, remove the duplicated
implementations from `models`.

**Test:** Full build. Regression: redispatch API endpoint, webhook firing.

---

### Phase 3 — Fix `Receipt` Trigger Path

#### Task 3.1 — Implement `TriggerReceipt` separate from `Trigger`
**File:** `src/models/dispatching_handler.go`

Currently `Receipt` calls `source.Trigger(msg)` — same path as new inbound messages.

Steps:
1. Verify what `Trigger` does: it calls all registered `QpDispatchingHandlerInterface.HandleDispatching`.
2. Verify that `WhatsappMessage` carries a `Type` field for receipts (e.g., `whatsapp.ReceiptMessageType`).
3. If the type is already set correctly before `Trigger` is called, the existing behavior may be
   acceptable — webhook consumers that filter by type are safe. **Confirm this before any change.**
4. If type is NOT reliably set: add explicit type enforcement in `Receipt`:
   ```go
   func (source *DispatchingHandler) Receipt(msg *whatsapp.WhatsappMessage) {
       if msg.Type == whatsapp.UnhandledMessageType || msg.Type == "" {
           msg.Type = whatsapp.ReceiptMessageType
       }
       source.Trigger(msg)
   }
   ```
5. Remove the 5 duplicate TODO comments.

**Test:** Send a message from an external device to the bot, verify receipt webhook payload has
a distinct type. Check existing webhook consumers (n8n flows in `/extra`) for type filtering.

---

### Phase 4 — Extract `dispatch/service` from `models`

> **Warning:** This is the highest-impact change. It requires restructuring the
> `models.DispatchingHandler` lifecycle methods. Do NOT start this phase until
> Phase 2 is fully complete and tested.

#### Task 4.1 — Map all `dispatch/service` call sites inside `models`
**File:** `src/models/dispatching_handler.go`

Run: `grep -rn "dispatchservice\." src/models/`

Expected sites:
- `dispatchservice.PublishRealtimeLifecycle(...)` in `LoggedOut`, `OnConnected`, `OnDisconnected`
- `dispatchservice.GetInstance().DispatchOutbound(...)` (from `QPDispatchingHandler` — removed in Phase 2)

Confirm the list before proceeding.

---

#### Task 4.2 — Move lifecycle events to `runtime`
**Files:**
- `src/models/dispatching_handler.go` (source)
- `src/runtime/` (destination — create `lifecycle_handler.go`)

Create `src/runtime/lifecycle_handler.go` with a `LifecycleHandler` type:
```go
type LifecycleHandler struct {
    server *models.QpWhatsappServer
}

func (h *LifecycleHandler) OnConnected() { ... }
func (h *LifecycleHandler) OnDisconnected() { ... }
func (h *LifecycleHandler) LoggedOut(reason string) { ... }
```

These methods call `dispatchservice.PublishRealtimeLifecycle` — valid since `runtime` IS allowed
to import `dispatch/service`.

Then in `models.DispatchingHandler`:
- Keep the method signatures (`OnConnected`, `LoggedOut`, etc.) as they are part of the
  `IWhatsappHandlers` interface.
- Replace the body with a delegation to an injected `LifecyclePublisher` interface:
  ```go
  type LifecyclePublisher interface {
      OnConnected(token, wid, phone, user string, verified bool)
      LoggedOut(token, wid, phone, user, reason string, verified bool)
  }
  ```
- Inject via `DispatchingHandler` constructor.

This removes the `dispatch/service` import from `models` entirely.

**Test:** Full build. QA: disconnect a server manually, verify lifecycle events appear in
WebSocket/SignalR clients and in webhook `kind: connected/logged_out` payloads.

---

### Phase 5 — Remove Direct Transport Imports from `models`

#### Task 5.1 — Remove `signalr` import from `models`
**File:** `src/models/qp_whatsapp_server.go`

Current usage:
```go
signalr.SignalRHub.HasActiveConnections(server.Token)
```

Steps:
1. Define a `RealtimePresenceChecker` interface in `models`:
   ```go
   type RealtimePresenceChecker interface {
       HasActiveConnections(token string) bool
   }
   ```
2. Inject via `QpWhatsappServer` constructor or as a package-level setter:
   ```go
   var GlobalPresenceChecker RealtimePresenceChecker
   ```
3. Replace the direct call with `GlobalPresenceChecker.HasActiveConnections(server.Token)`.
4. In `main.go` (or webserver init), wire: `models.GlobalPresenceChecker = signalr.SignalRHub`.
5. Remove the `signalr` import from `models`.

**Test:** Build. QA: verify SignalR hub connection status is still correctly reflected.

---

#### Task 5.2 — Remove `rabbitmq` import from `models`
**File:** `src/models/qp_whatsapp_server.go`

Identify all rabbitmq call sites inside `models`. Likely usage: `InitializeRabbitMQConnections`.

Apply the same interface injection pattern as Task 5.1.

**Test:** Build. QA: configure a RabbitMQ dispatching, verify messages are published.

---

### Completion Checklist

| Phase | Task | Status |
|---|---|---|
| 1 | 1.1 Fix `IsConnecting` atomic | ⬜ |
| 1 | 1.2 Fix timestamp wrap | ⬜ |
| 1 | 1.3 Remove dead `DispatchingHandler` field | ⬜ |
| 2 | 2.1 Remove `QPDispatchingHandler` | ⬜ |
| 2 | 2.2 Remove duplicated dispatch functions from `models` | ⬜ |
| 3 | 3.1 Fix `Receipt` / `TriggerReceipt` | ⬜ |
| 4 | 4.1 Map `dispatch/service` sites in `models` | ⬜ |
| 4 | 4.2 Move lifecycle events to `runtime` | ⬜ |
| 5 | 5.1 Remove `signalr` import from `models` | ⬜ |
| 5 | 5.2 Remove `rabbitmq` import from `models` | ⬜ |

---

*Analysis date: 2026-04-29*
*Branch: develop*
