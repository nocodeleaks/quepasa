# QuePasa Architecture Decisions

## Purpose

This document is the entry point for formal architecture decisions recorded for
QuePasa.

Unlike the broader architecture documents, these records focus on one decision
at a time:

- the context behind the decision
- the selected direction
- the consequences of choosing it

## When To Use ADRs

Use an architecture decision record when:

- multiple valid directions exist
- the chosen direction affects several packages
- the team needs a stable explanation for future refactors
- a future reviewer would otherwise ask why a structural constraint exists

Do not use ADRs for:

- routine implementation details
- temporary branch-only notes
- low-impact local refactors with no lasting architectural meaning

## Current ADR Set

- `ADR-0001-MODULAR-MONOLITH-INCREMENTAL-REFACTORING.md`
- `ADR-0002-SESSION-AS-RUNTIME-CONCEPT.md`
- `ADR-0003-MODELS-IS-NOT-THE-ESCAPE-HATCH.md`
- `ADR-0004-EXPLICIT-APPLICATION-LAYER.md`
- `ADR-0005-GROUPED-COMPOSITION-ROOT-WIRING.md`

## Recommended Reading Order

1. `ADR-0001-MODULAR-MONOLITH-INCREMENTAL-REFACTORING.md`
2. `ADR-0002-SESSION-AS-RUNTIME-CONCEPT.md`
3. `ADR-0003-MODELS-IS-NOT-THE-ESCAPE-HATCH.md`
4. `ADR-0004-EXPLICIT-APPLICATION-LAYER.md`
5. `ADR-0005-GROUPED-COMPOSITION-ROOT-WIRING.md`

## Related Documents

- `ARCHITECTURE-INDEX.md`
- `ARCHITECTURE-CURRENT-STATE.md`
- `ARCHITECTURE-TARGET-STATE.md`
- `ARCHITECTURE-ROADMAP.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`

## Maintenance Rule

When an architecture decision is still active, update or supersede its ADR
instead of silently drifting away from it in code or planning documents.