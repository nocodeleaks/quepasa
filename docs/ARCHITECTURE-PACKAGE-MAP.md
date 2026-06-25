# QuePasa Architecture Package Map

## Purpose

This document maps the main QuePasa packages to their practical architectural
roles.

It is not a directory listing. It is a responsibility map intended to make
package ownership easier to reason about during refactoring and feature work.

## Reading Guide

Each package is classified by:

- primary responsibility
- current strength
- current risk
- recommended direction

## Composition Root

### `src`

Primary responsibility:

- process bootstrap
- startup wiring
- binary entry point

Key files:

- `main.go`

Current strength:

- centralized startup order is easy to locate

Current risk:

- too much subsystem wiring lives directly in the entry point

Recommended direction:

- keep `main.go` as the composition root
- reduce ad-hoc global assignments over time through grouped service wiring

## HTTP and Frontend Host

### `src/webserver`

Primary responsibility:

- router creation
- middleware chain
- frontend app discovery and mounting
- route registration host for feature packages

Current strength:

- strong extension model through configurator registration
- clean support for backend-managed and static frontend apps

Current risk:

- `webserver.go` is still large enough to accumulate unrelated behavior

Recommended direction:

- preserve the current configurator-based design
- keep frontend discovery and router composition bounded

### `src/apps/*`

Primary responsibility:

- isolated frontend applications mounted by slug

Current strength:

- explicit slug-based isolation
- no implicit aliasing requirement

Current risk:

- app bundles may drift operationally if deployment artifacts are not kept in
  sync with source expectations

Recommended direction:

- keep app isolation strict
- treat each slug as a first-class app surface

## HTTP API Surface

### `src/api`

Primary responsibility:

- HTTP controllers
- request parsing
- response formatting
- version and compatibility route exposure

Current strength:

- broad functional coverage
- active separation between canonical and legacy route registration exists

Current risk:

- package is large and operationally central
- legacy compatibility breadth increases maintenance cost
- several controllers remain near or above the preferred file-size threshold

Recommended direction:

- continue isolating transport DTOs and version-specific behavior inside API
  ownership
- keep shrinking compatibility spread into bounded registration paths

### `src/api/legacy`

Primary responsibility:

- compatibility route registration

Current strength:

- legacy behavior is at least being concentrated instead of remaining fully mixed

Current risk:

- large alias tables widen the supported HTTP surface

Recommended direction:

- keep this package as quarantine for compatibility behavior
- retire aliases only with evidence and migration safety

### `src/api/v5`

Primary responsibility:

- canonical version registration and alias mounting

Current strength:

- clearer modern route registration point

Current risk:

- still depends on a large controller surface in the root API package

Recommended direction:

- strengthen v5 as the canonical transport entry point over time

## Runtime Core

### `src/models`

Primary responsibility in practice:

- session runtime state
- domain entities
- lifecycle behavior
- persistence-adjacent mutations
- compatibility wrappers
- manager composition

Current strength:

- contains the business center of the system
- already undergoing responsibility-driven decomposition

Current risk:

- still acts as the default escape hatch for mixed concerns
- package meaning is broader than the name suggests
- orchestration and persistence still overlap conceptually

Recommended direction:

- keep only domain state and truly domain-bound behavior here
- move orchestration pressure outward into an explicit application layer
- avoid adding new transport or compatibility shapes here unless unavoidable

### `src/runtime`

Primary responsibility in practice:

- currently very small lifecycle publishing support

Current strength:

- package name reserves the right conceptual place for future use-case
  orchestration

Current risk:

- intended application-layer responsibilities still live mostly elsewhere

Recommended direction:

- grow this package only with explicit workflow orchestration
- do not turn it into another miscellaneous package

## WhatsApp Boundary

### `src/whatsapp`

Primary responsibility:

- abstractions and contracts around WhatsApp-facing concepts

Current strength:

- provides a cleaner seam than binding everything directly to the SDK layer

Current risk:

- can become blurred if business logic keeps leaking into abstraction helpers

Recommended direction:

- preserve it as the contract-facing layer between domain and SDK integration

### `src/whatsmeow`

Primary responsibility:

- external SDK integration
- event translation
- connection management

Current strength:

- adapter package is now more structured than before
- event routing is more explicit with the event router model

Current risk:

- still contains very large files
- boundary code is carrying too much workflow weight

Recommended direction:

- keep splitting by capability and event family
- prefer translating external events into application behavior instead of growing
  orchestration here

## Outbound and Realtime Transport

### `src/dispatch/service`

Primary responsibility:

- outbound dispatch coordination
- delivery fanout to external targets
- policy-aware transport orchestration

Current strength:

- now cleaner after policy extraction
- good candidate for preserving transport neutrality

Current risk:

- must resist reabsorbing business filtering logic directly into delivery code

Recommended direction:

- keep policy and delivery concerns separate
- preserve the service as an outbound coordinator, not a business-rule sink

### `src/rabbitmq`

Primary responsibility:

- queue transport integration

Current strength:

- clear external transport ownership

Current risk:

- transport modules can easily accumulate cross-cutting operational helpers

Recommended direction:

- keep queue-specific behavior local to this package

### `src/cable`

Primary responsibility:

- websocket realtime fanout

Current strength:

- dedicated realtime transport package

Current risk:

- command and hub files are already moderately large

Recommended direction:

- keep browser realtime transport concerns isolated here

### `src/signalr`

Primary responsibility:

- SignalR realtime transport

Current strength:

- clear protocol ownership

Current risk:

- avoid letting application workflow drift into hub-level handlers

Recommended direction:

- remain a protocol adapter, not a lifecycle owner

## Persistence and Infrastructure Support

### `src/cache`

Primary responsibility:

- centralized cache strategy and backends

Current strength:

- one of the cleaner internal subsystem designs
- backend strategy is explicit and practical

Current risk:

- cache injection points must remain consistent with the centralized design

Recommended direction:

- preserve the single service entry pattern

### `src/environment`

Primary responsibility:

- configuration loading and environment interpretation

Current strength:

- strong central ownership of runtime settings

Current risk:

- environment logic can become a hidden dependency driver if too much behavior is
  conditioned globally

Recommended direction:

- keep this package configuration-focused

### `src/library`

Primary responsibility:

- reusable helpers and common support utilities

Current strength:

- shared utility surface for multiple packages

Current risk:

- utility packages can become unbounded if not kept disciplined

Recommended direction:

- keep helper scope explicit and avoid turning this into a generic dumping ground

## Specialized Subsystems

### `src/media`

Primary responsibility:

- media processing helpers and tooling

Recommended direction:

- keep media concerns local and adapter-facing

### `src/mcp`

Primary responsibility:

- Model Context Protocol exposure

Recommended direction:

- treat it as an external protocol surface, not a domain owner

### `src/metrics`

Primary responsibility:

- metrics collection and registration

Recommended direction:

- keep metrics passive and observational rather than letting it shape business
  flow

### `src/sipproxy`

Primary responsibility:

- SIP and VoIP-related runtime support

Recommended direction:

- keep it isolated from WhatsApp session semantics as much as possible

## Practical Conclusion

The current package map shows a codebase with a solid modular shell and an
unfinished inner layering model.

The main architectural work ahead is not inventing new packages. It is making
ownership more honest inside the packages that already dominate the system:

- `models`
- `api`
- `whatsmeow`

## Related Documents

- `ARCHITECTURE-INDEX.md` is the entry point for this documentation set
- `ARCHITECTURE-CURRENT-STATE.md` explains the system-level current state
- `ARCHITECTURE-TARGET-STATE.md` defines the recommended destination
- `ARCHITECTURE-ROADMAP.md` defines the preferred execution order
- `ARCHITECTURE-EXECUTION-CHECKLIST.md` provides implementation checkpoints
- `PLAN-ARCHITECTURE-REFACTORING.md` contains specific refactor findings and
  completed items