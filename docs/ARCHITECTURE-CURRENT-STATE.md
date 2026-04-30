# QuePasa Architecture Current State

## Purpose

This document describes the current architectural shape of QuePasa as it exists
today, focusing on actual package responsibilities, dependency pressure, and the
main structural bottlenecks that affect maintenance.

It is intentionally different from a refactoring plan:

- this document describes what the system is
- `PLAN-ARCHITECTURE-REFACTORING.md` describes what should change

## Executive Summary

QuePasa is currently a modular monolith with clear integration boundaries but an
incomplete internal layering model.

The codebase already contains strong architectural decisions:

- transport modules are separated from each other
- frontend apps are isolated by slug
- route registration is extensible through webserver configurators
- the central runtime naming is moving from `server` to `session`

The main structural issue is not the lack of architecture. The main issue is
that the intended architecture is only partially enforced in the current package
boundaries.

In practice, `models`, `api`, and `whatsmeow` still act as the three main
gravity centers of the codebase.

## Current High-Level Shape

The runtime can be understood as five major zones.

### 1. Composition Root

Main bootstrap happens in:

- `src/main.go`

Responsibilities currently owned there:

- environment and log level initialization
- database migration startup
- centralized cache initialization
- whatsmeow startup
- transport adapter wiring into `models`
- background migration handlers
- WhatsApp service startup
- web server startup

This makes `main.go` the effective composition root for the whole process.

### 2. WhatsApp Transport Boundary

Main packages:

- `src/whatsmeow`
- `src/whatsapp`

Current role:

- `whatsapp` defines domain-facing abstractions and option contracts
- `whatsmeow` adapts the external library into QuePasa runtime behavior

Important evolution already present:

- event dispatch is no longer a single giant switch only
- `whatsmeow_event_router.go` introduces a more explicit event routing model

Remaining pressure:

- `whatsmeow_connection.go` is still very large
- `whatsmeow_handlers.go` still concentrates too many runtime concerns for a
  boundary adapter package

### 3. Session Runtime Core

Main package:

- `src/models`

Current role in practice:

- session identity and state
- dispatching registration and orchestration
- lifecycle transitions
- persistence helpers and save/delete flows
- compatibility surface for the `server` to `session` migration
- manager composition for contacts, groups, and status

Important evolution already present:

- `qp_whatsapp_session.go` introduces the preferred runtime term `session`
- `session_intent.go` replaces paired lifecycle booleans with an explicit intent
- `server_connection.go`, `server_messaging.go`, and `server_persistence.go`
  split a previously overloaded server file into bounded responsibilities

Remaining pressure:

- `models` is still the default landing zone for orchestration code
- runtime concerns and persistence concerns still share the same package surface
- the package name still hides too many different conceptual roles

### 4. Outbound Dispatch and Realtime Delivery

Main packages:

- `src/dispatch/service`
- `src/rabbitmq`
- `src/cable`
- `src/signalr`

Current role:

- `dispatch/service` coordinates outbound message delivery
- transport-specific packages implement actual delivery mechanisms
- realtime transports register themselves without forcing the webserver package
  to know transport details

Important evolution already present:

- `DispatchPolicy` was extracted from raw delivery logic
- the dispatch service is closer to a transport coordinator than a business-rule
  owner

Remaining pressure:

- some dependency wiring still uses globals configured at startup
- transport inversion has improved, but the composition model is still
  transitional rather than fully constructor-driven

### 5. HTTP and Frontend Delivery

Main packages:

- `src/webserver`
- `src/api`
- `src/apps/*`

Current role:

- `webserver` hosts route registration and frontend discovery
- `api` exposes HTTP endpoints and compatibility aliases
- `apps/*` contains isolated frontend applications mounted by slug

Important evolution already present:

- backend modules register routes through `RegisterRouterConfigurator`
- frontend apps are discovered and served by exact slug matching
- backend-managed apps can coexist with static SPA apps under `/apps/<slug>`

Remaining pressure:

- the API package is still very large
- legacy aliases continue to widen the surface area of the HTTP layer

## Current Structural Strengths

### Strong extensibility at the web boundary

The route registration model in `src/webserver/webserver.go` is one of the
strongest parts of the codebase.

It gives the system a plugin-like shape without creating a separate plugin
system:

- API registers its own routes
- SignalR registers its own routes
- cable registers its own routes
- Swagger registers its own routes
- form registers its own routes

This avoids central route sprawl inside the webserver package.

### Clean frontend app isolation

The frontend mounting rules under `src/apps/<slug>` are structurally sound.

Benefits:

- no hidden aliasing between apps
- no semantic fallback from one app to another
- each app owns its own assets and SPA fallback
- future frontends can be added without rewriting server routing assumptions

### Transport separation is directionally correct

The separation between transport-specific packages and outbound coordination is
good and should be preserved.

Notable examples:

- dispatch service coordinates delivery
- RabbitMQ owns queue transport behavior
- cable owns websocket fanout
- SignalR owns hub transport behavior

### Architecture is already evolving incrementally

Recent extractions show the codebase can be improved without a rewrite:

- event router added in whatsmeow
- dispatch policy extracted
- session intent introduced
- server responsibilities split across multiple files

This is important because it means the correct strategy is iterative structural
hardening, not a risky full redesign.

## Current Structural Weaknesses

### `models` remains the largest conceptual bottleneck

The package still acts as a mixed home for:

- domain entities
- runtime orchestration
- persistence mutations
- compatibility wrappers
- manager composition
- transport-neutral adapter seams

The main consequence is cognitive load.

Every dependency on `models` brings more conceptual weight than the package name
implies.

### The application layer is still mostly implicit

The project documentation and package naming indicate a desired distinction
between domain/runtime concepts and higher-level orchestration.

In practice, the runtime or application-service layer is still thin.

This means workflows such as:

- start session
- stop session
- reconnect session
- send outbound message
- restore history
- pair identity

are still too close to the entity/state package surface.

### Composition still relies on startup globals

The current startup wiring improves dependency direction compared to direct
imports from core packages, but it still relies on global assignments in
`main.go`.

That approach is acceptable as a transition, but it has costs:

- dependencies are less explicit
- test setup is less localized
- lifecycle assumptions are encoded in startup order

### API compatibility surface is broad

The root API layer still carries a large historical surface through aliases and
legacy route tables.

That increases:

- test surface
- documentation burden
- maintenance overhead
- risk of subtle divergence across equivalent endpoints

### Boundary adapters are still heavier than ideal

The `whatsmeow` package is clearly an integration boundary, but some of its
files are still large enough to suggest hidden workflow concentration.

That matters because boundary packages should normally become easier to reason
about over time, not another center of business flow.

## Package Pressure Snapshot

At the time of this analysis, the largest package totals in `src/` were:

- `src/api` ≈ 10k lines
- `src/models` ≈ 8k lines
- `src/whatsmeow` ≈ 5k lines

Notable large files included:

- `src/whatsmeow/whatsmeow_connection.go`
- `src/whatsmeow/whatsmeow_handlers.go`
- `src/webserver/webserver.go`
- several API controllers above the preferred file-size threshold

This does not automatically mean those packages are wrong, but it does confirm
where future structural effort will have the highest return.

## What The Codebase Is Optimized For Today

The current architecture is optimized for:

- practical feature delivery in a single deployable process
- backwards compatibility at the HTTP layer
- incremental refactoring without destabilizing runtime behavior
- integration-heavy workflows over strict academic layering

That is a valid set of priorities.

The downside is that the codebase now needs stronger architectural governance to
avoid returning to package sprawl as new features are added.

## Current Architectural Conclusion

QuePasa should be understood as a modular monolith in transition.

It already contains the right architectural ideas, but some of the most
important ones are not yet fully reflected in package ownership and dependency
direction.

The next stage of improvement should focus less on renaming and more on making
the application layer and dependency boundaries explicit.

## Related Documents

- `ARCHITECTURE-INDEX.md` is the entry point for this documentation set
- `ARCHITECTURE-TARGET-STATE.md` defines the recommended end state
- `ARCHITECTURE-ROADMAP.md` defines the execution order
- `ARCHITECTURE-EXECUTION-CHECKLIST.md` converts the roadmap into implementation
  checkpoints
- `ARCHITECTURE-PACKAGE-MAP.md` describes the current package roles package by
  package
- `PLAN-ARCHITECTURE-REFACTORING.md` tracks concrete refactoring items already
  identified