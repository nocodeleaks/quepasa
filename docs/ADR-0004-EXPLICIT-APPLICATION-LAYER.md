# ADR-0004: Explicit Application Layer For Session Workflows

## Status

Accepted

## Context

QuePasa already documents a target architecture in which session lifecycle and
related workflows should be orchestrated by an explicit application layer.

Today, that layer is still mostly implicit.

In practice, workflow behavior is still concentrated across:

- `src/models`
- `src/whatsmeow`
- transport-facing packages that invoke runtime behavior directly

This makes the system harder to reason about because a reviewer often has to
reconstruct a use case from entity methods, adapter callbacks, and transport
entry points instead of finding one workflow owner.

The main workflows affected are:

- start session
- stop session
- restart session
- pair session
- delete session
- send message
- restore or sync history

## Decision

QuePasa will establish an explicit application layer for session workflows.

This layer will own orchestration and use-case flow, while `models` remains
focused on domain/runtime state and `whatsmeow` remains focused on external SDK
integration.

## Consequences

### Positive

- workflow ownership becomes easier to discover
- testing can target use cases more directly
- `models` can stop absorbing orchestration by convenience
- adapter packages become easier to keep thin

### Negative

- some workflow code will be temporarily split across old and new ownership
  during migration
- contributors will need to make package-boundary decisions earlier

## Implications

- new workflow logic should prefer the explicit application layer over `models`
  when the code coordinates multiple dependencies
- entity/state methods should stay close to invariants and runtime state
- transport packages should call application behavior rather than becoming the
  primary workflow owners
- migration should happen one use case at a time

## Related Documents

- `ARCHITECTURE-TARGET-STATE.md`
- `ARCHITECTURE-ROADMAP.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`
- `ADR-0003-MODELS-IS-NOT-THE-ESCAPE-HATCH.md`