# Models Remodeling Audit

## Objective

Document why the current `src/models` package became a maintenance bottleneck
and define a practical refactor path that improves boundaries without requiring
one risky rewrite.

This audit is meant to run in parallel with the PR `#39` review so backend
cleanup and SPA extraction can move in the same direction.

## Executive Summary

The current `models` package is not acting as a domain-model package anymore.
It is acting as a catch-all package for at least six different concerns:

- domain entities
- runtime/service orchestration
- persistence and migrations
- transport DTOs for HTTP/form
- infrastructure integrations
- environment/config helpers

That shape is the main reason the package now feels disorganized: the problem is
not the amount of code alone, but the number of responsibilities sharing the
same namespace and dependency surface.

## Main Findings

### 1. `models` mixes core and edge concerns

Representative examples:

- [src/models/qp_server.go](/abs/path/Z:/Desenvolvimento/nocodeleaks-quepasa/src/models/qp_server.go)
  Domain entity.
- [src/models/qp_whatsapp_service.go](/abs/path/Z:/Desenvolvimento/nocodeleaks-quepasa/src/models/qp_whatsapp_service.go)
  Application/runtime service.
- [src/models/qp_database.go](/abs/path/Z:/Desenvolvimento/nocodeleaks-quepasa/src/models/qp_database.go)
  Persistence and migrations.
- [src/models/qp_form_account_data.go](/abs/path/Z:/Desenvolvimento/nocodeleaks-quepasa/src/models/qp_form_account_data.go)
  Legacy form view-model.
- [src/models/qp_send_request.go](/abs/path/Z:/Desenvolvimento/nocodeleaks-quepasa/src/models/qp_send_request.go)
  HTTP request DTO plus conversion logic.
- [src/models/qp_contact_manager.go](/abs/path/Z:/Desenvolvimento/nocodeleaks-quepasa/src/models/qp_contact_manager.go)
  Adapter around WhatsMeow-facing contact access.

Best-practice consequence:

- the package name stops communicating intent
- every import of `models` pulls mental overhead from unrelated concerns
- the easiest way to avoid cycles becomes “put it into models too”

### 2. Transport-specific DTOs are stored in the core package

Examples:

- `qp_form_*`
- `qp_*response*`
- `qp_send_request*`
- `qp_receive_response*`

Best-practice consequence:

- API versioning and template rendering details leak into the core package
- changing one transport shape risks unnecessary recompilation and coupling
- versioned payloads become harder to retire

### 3. `models` imports infrastructure directly

Observed imports from `models`:

- `whatsmeow`
- `rabbitmq`
- `signalr`
- `environment`
- `library`

Best-practice consequence:

- the dependency direction is inverted
- core logic depends on delivery/transport choices
- testing and mocking become harder because the domain layer knows too much

### 4. Persistence is mixed with runtime orchestration

Examples:

- `qp_database.go`
- `qp_data_*`
- `qp_whatsapp_service.go`
- `qp_whatsapp_server.go`

Best-practice consequence:

- database concerns and lifecycle concerns evolve together
- repository/store abstractions exist, but they are still buried inside the same
  package as runtime state and message dispatch logic

### 5. Legacy compatibility shapes are staying alive in the wrong layer

Examples:

- form-specific template structs

Best-practice consequence:

- compatibility concerns occupy central package space instead of being confined
  to the transport that needs them

## Current Package Roles Hidden Inside `models`

The package currently contains at least these logical groups:

### Domain

- `qp_user.go`
- `qp_server.go`
- `qp_dispatching.go`
- `qp_webhook.go`
- `qp_rabbitmq_config.go`
- `qp_timestamps.go`

### Runtime / Application Services

- `qp_whatsapp_service.go`
- `qp_whatsapp_server.go`
- `dispatching_handler.go`
- `qp_contact_manager.go`
- `qp_group_manager.go`
- `qp_status_manager.go`
- `realtime_publishers.go`

### Persistence

- `qp_database.go`
- `qp_database_config.go`
- `qp_data_*`
- `qp_migration_file.go`
- `seed.go`

### Transport DTOs

- `qp_send_request*.go`
- `qp_send_response*.go`
- `qp_receive_response*.go`
- `qp_health_response_item.go`

### Form View Models

- `qp_form_*`

### Integration Adapters

- `qp_whatsapp_extensions_whatsmeow.go`
- `qp_contact_manager.go`
- `qp_group_manager.go`
- `qp_dispatching_handler.go`

## Target Direction

The goal is not to create many tiny packages for the sake of it. The goal is to
separate the system by responsibility and dependency direction.

Recommended target split:

### `core`

Pure business/domain concepts:

- users
- servers
- dispatching configuration
- timestamps
- domain errors

Rules:

- should not import `api`, `form`, `signalr`, `cable`, `rabbitmq`, `whatsmeow`
- may depend on narrow abstractions/interfaces defined close to the core

### `app`

Use cases and orchestration:

- start/stop server
- send message
- pair device
- dispatch runtime events

Rules:

- may depend on `core` and on interfaces implemented by adapters
- should contain workflow logic, not HTTP/template details

### `store`

Persistence implementations:

- sqlite connection/bootstrap
- migrations
- user/server/dispatching repositories

Rules:

- owns SQL and schema concerns
- should not contain HTTP or template DTOs

### `transport/api`

HTTP-specific contracts:

- request/response DTOs
- versioned payload shapes
- response projection helpers

Rules:

- API versioning stays here
- no reason for transport-specific payload structs to stay in `core`

### `transport/form`

Template/view data:

- page data
- template helpers

Rules:

- form rendering state belongs here, not in `models`

### `integrations`

Boundary adapters:

- WhatsMeow adapter
- webhook publishing
- RabbitMQ publishing
- realtime publishing

Rules:

- depends inward on interfaces/use cases
- core should not import these packages directly

## Recommended Refactor Order

### Phase 1: Remove obvious transport pollution

Safe moves:

- move `qp_form_*` out of `models`
- move API-only request/response DTOs out of `models`
- leave aliases only if needed for transition

Status:

- started in this turn for legacy form page data
- continued in this turn for API-only response DTOs
- completed for the first batch of API response DTOs by moving them into
  `src/api/models` and deleting their duplicated definitions from `src/models`
- continued by moving the HTTP send request contracts into `src/api/models`
- continued by moving the health response item projection into `src/api/models`

### Phase 2: Isolate persistence

Move toward:

- `store/sqlite`
- repository implementations
- migration helpers

Benefits:

- shrinks `models`
- makes DB dependencies explicit

### Phase 3: Isolate runtime service layer

Move toward:

- `app` or `runtime`
- server lifecycle orchestration
- message dispatch orchestration

Benefits:

- runtime code stops competing for space with DTOs and schema code

### Phase 4: Invert integration dependencies

Priority hotspots:

- remove direct `signalr` import from core runtime code
- keep realtime publishing behind transport-neutral interfaces
- reduce direct `whatsmeow` calls from broad packages to focused adapters

### Phase 5: Clean naming and compatibility layers

After boundaries are improved:

- rename generic package names to capability names
- simplify prefixes where package names already provide context

## First Safe Slice Chosen

The first extraction chosen here is:

- move form page/view-model structs out of `models`

Reason:

- they are used only by the `form` module
- they do not carry domain rules
- they are template-facing edge data
- this change improves package boundaries with minimal behavioral risk

## Best-Practice Guidance For This Codebase

### 1. Stop using `models` as the escape hatch

Rule:

- if a type exists only for one transport or one adapter, it should live there

### 2. Keep dependency direction inward

Rule:

- transport/infrastructure may depend on core
- core should not depend on transport/infrastructure implementation packages

### 3. Use packages named by capability, not by generic category

Good:

- `store`
- `webhook`
- `dispatch`
- `viewmodel`
- `runtime`

Weak:

- `models`
- `utils`
- `helpers`

### 4. Put versioned contracts at the boundary

Rule:

- `spa`, `form`, and any future versioned shapes belong to transport layers

### 5. Prefer phased extraction over big-bang rewrites

Rule:

- move leaf types first
- then repositories
- then services
- then dependency inversion for integrations

## Next Recommended Slice

After the form view-model extraction, the first API response DTO extraction,
and the removal of obsolete `v1`/`v2` compatibility types from `models`, the
next safest move is:

- continue moving API-only request payloads from `models` into `src/api/models`
  or a new `src/api/dto`

Priority candidates:

This keeps the transport boundary moving in the right direction before touching
the heavier runtime/persistence split.

## Immediate Recommendation

The next high-value extraction should not start with the runtime services yet.
It should start with transport contracts that are still leaking into shared
layers, especially because the new realtime cable module will depend on clean
command/event contracts.

Recommended order from here:

1. Move API request DTOs that are currently shared only because of convenience.
2. Define explicit cable command/event envelopes in `src/cable` instead of
   reusing HTTP request payloads.
3. Keep `models` focused on domain and runtime state while planning a later
   split of persistence and orchestration.

Why this order:

- it reduces accidental coupling between HTTP and websocket/cable flows
- it gives the cable module stable contracts for multiple simultaneous clients
- it prepares the codebase for a future `runtime` and `store` separation without
  mixing transport concerns again

Progress note:

- the websocket `message.send` flow now uses a cable-local request contract
  instead of reusing `models.QpSendAnyRequest`
- obsolete `v1`/`v2` compatibility files were removed from `src/models`
- `QpSendRequest` and `QpSendAnyRequest` were replaced by API-local request
  contracts in `src/api/models`
- `QpHealthResponseItem` was replaced by an API-local health item projection
