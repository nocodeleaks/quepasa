# QuePasa Architecture Target State

## Purpose

This document defines the recommended target architecture for QuePasa after the
current transition phase is complete.

The target is not a rewrite. It is an incremental end state that the codebase
can approach through bounded refactors.

## Design Goals

The target architecture should optimize for five outcomes:

1. clear package responsibility
2. explicit dependency direction
3. predictable runtime workflows
4. transport isolation
5. safe incremental migration from the current codebase

## Recommended Layer Model

QuePasa should evolve toward five stable layers.

### 1. Core Domain Layer

Suggested responsibility:

- business entities
- state representations
- domain errors
- narrow domain interfaces
- value objects and invariants

Examples of concepts that belong here:

- session identity
- dispatch target configuration
- timestamps and metadata
- domain-level validation rules

Rules:

- must not depend on HTTP, RabbitMQ, SignalR, cable, or whatsmeow
- should expose minimal interfaces when external capabilities are needed
- should remain small and conceptually stable

### 2. Application Layer

Suggested responsibility:

- use cases
- lifecycle orchestration
- outbound dispatch workflows
- coordination between domain, persistence, and adapters

Examples of use cases:

- start session
- stop session
- reconnect session
- pair session
- delete session
- send outbound message
- restore messages

Rules:

- depends on the core domain layer
- depends on interfaces implemented by adapters and stores
- does not own HTTP request parsing or external SDK details

This is the layer currently missing in explicit form.

### 3. Store Layer

Suggested responsibility:

- database bootstrap
- migrations
- repository implementations
- persistence adapters
- low-level storage coordination

Rules:

- owns SQL and physical persistence concerns
- exposes repository behavior needed by the application layer
- does not own HTTP DTOs or frontend view models

### 4. Transport Layer

Suggested responsibility:

- HTTP API input/output
- form rendering contracts
- versioned route compatibility
- realtime transport endpoints
- external delivery transports

Subareas may include:

- `transport/api`
- `transport/form`
- `transport/realtime`
- `transport/outbound`

Rules:

- converts transport-specific data into application use cases
- does not own the business lifecycle of sessions

### 5. External Adapter Layer

Suggested responsibility:

- whatsmeow integration
- RabbitMQ concrete integration
- SignalR concrete integration
- websocket fanout implementation
- media tools and other infrastructure bindings

Rules:

- depends inward on application/domain contracts
- should be replaceable without changing the core model

## Recommended Package Direction

The dependency direction should move inward.

Preferred direction:

- adapters and transport depend on application contracts
- application depends on domain and store interfaces
- store depends on domain where necessary
- domain depends on nothing transport-specific

Avoid the opposite direction:

- domain importing RabbitMQ
- domain importing SignalR
- domain importing HTTP concerns
- domain importing whatsmeow directly

## Recommended Ownership Of Current Concerns

### Session lifecycle

Current owner in practice:

- `src/models`

Recommended owner:

- application layer orchestration with a smaller domain session model beneath it

### Persistence mutations

Current owner in practice:

- `src/models`

Recommended owner:

- store/repository layer

### HTTP request and response contracts

Current owner in practice:

- split between `src/api` and residual legacy shapes historically tied to models

Recommended owner:

- transport API layer only

### WhatsApp SDK event handling

Current owner in practice:

- `src/whatsmeow`

Recommended owner:

- adapter layer, translated into application-layer events or use-case calls

### Outbound routing policy

Current owner in practice:

- `src/dispatch/service` with extracted policy abstraction

Recommended owner:

- policy contract close to the orchestration boundary, with transport remaining
  delivery-focused

## Session As The Central Runtime Concept

The runtime should treat `session` as the primary business concept for one
WhatsApp-connected identity.

This means:

- `server` remains valid only for infrastructure concerns such as web servers,
  SIP servers, and transport hosts
- `session` becomes the preferred domain/runtime term for WhatsApp identity
  lifecycle

This naming shift is valuable only if responsibility follows the name.

The target end state is not just:

- `QpWhatsappSession = QpWhatsappServer`

The target end state is:

- a small session core model
- explicit session lifecycle use cases
- explicit persistence and transport boundaries around that model

## Recommended Composition Model

The composition root should become more explicit over time.

Target characteristics:

- startup assembles a small number of service groups
- dependencies are wired through constructors or explicit setup structs
- global indirection is reduced to the minimum needed for compatibility

Good transitional pattern:

- one aggregated services struct per major subsystem

Examples:

- session services
- outbound transport services
- realtime services
- repository services

## Recommended Quality Gates

The architecture should be protected by a few explicit rules.

### File size discipline

Keep following `docs/CODE_ORGANIZATION.md`.

New feature work should not expand already overloaded files when extraction is
practical.

### No new escape-hatch growth in `models`

`models` should not keep growing as the default place for uncertain ownership.

When ownership is unclear, the design question should be answered before placing
new code.

### Legacy compatibility must stay quarantined

Compatibility shims are valid, but they should remain visibly temporary and stay
near the transport layer whenever possible.

### Adapters should translate, not orchestrate everything

Boundary packages should translate external events and invoke application
behavior, not become alternative centers of business workflow.

## Target Architectural Conclusion

The desired QuePasa architecture is a modular monolith with explicit
application-layer orchestration, slimmer domain ownership, and adapter-driven
integrations.

The important shift is not from monolith to microservices.

The important shift is from package-level convenience toward responsibility-led
boundaries.

## Related Documents

- `ARCHITECTURE-INDEX.md` is the entry point for this documentation set
- `ARCHITECTURE-CURRENT-STATE.md` describes the current structural reality
- `ARCHITECTURE-ROADMAP.md` translates the target into execution phases
- `ARCHITECTURE-EXECUTION-CHECKLIST.md` provides implementation checkpoints
- `ARCHITECTURE-PACKAGE-MAP.md` maps the current package landscape
- `PLAN-ARCHITECTURE-REFACTORING.md` contains the deeper refactor inventory