# ADR-0001: Modular Monolith With Incremental Refactoring

## Status

Accepted

## Context

QuePasa is a large integration-heavy system with active production behavior,
legacy API compatibility, multiple runtime transports, and a growing internal
architecture refactoring effort.

The codebase already behaves as a modular monolith:

- one main deployable backend process
- multiple internal packages with bounded responsibilities
- several transport and protocol integrations
- shared runtime state across modules

At the same time, the architecture still contains overloaded package centers,
especially in `models`, `api`, and `whatsmeow`.

The project needs structural improvement, but a disruptive redesign would carry
high regression risk because:

- session lifecycle behavior is central to the product
- HTTP compatibility breadth is still significant
- startup composition and transport integration are operationally sensitive

## Decision

QuePasa will continue to evolve as a modular monolith, and architectural
improvement will be performed through incremental refactoring rather than a full
rewrite or premature service decomposition.

## Consequences

### Positive

- refactors can be validated slice by slice
- compatibility can be preserved where needed
- operational risk stays lower than in disruptive redesigns
- package boundaries can be improved without changing deployment topology

### Negative

- transition code and compatibility layers may exist for some time
- architectural cleanup requires discipline over multiple branches
- temporary duplication may be tolerated during migration phases

## Implications

- prefer extraction over replacement
- prefer compatibility layers over breaking broad call-site changes
- prefer package-boundary hardening over system-level decomposition
- evaluate architectural work by coupling reduction, not by novelty

## Related Documents

- `ARCHITECTURE-CURRENT-STATE.md`
- `ARCHITECTURE-TARGET-STATE.md`
- `ARCHITECTURE-ROADMAP.md`