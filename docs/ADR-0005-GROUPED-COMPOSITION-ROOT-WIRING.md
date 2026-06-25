# ADR-0005: Grouped Composition Root Wiring Over Broad Startup Globals

## Status

Accepted

## Context

The current startup path in `src/main.go` assembles the process correctly, but
it still relies on a broad set of global assignments to wire transport and
infrastructure behavior into runtime packages.

This approach improved dependency direction compared to direct imports inside the
runtime core, but it still has costs:

- dependencies are discovered by startup order rather than by local constructor
  signatures
- test setup can become global and scattered
- subsystem ownership is less visible than it should be

The current transitional wiring is acceptable, but it should not become the
long-term composition model.

## Decision

QuePasa will gradually replace broad startup global wiring with grouped
composition-root wiring using explicit setup structs or subsystem service groups
where migration is practical.

Compatibility globals may remain temporarily, but they should be treated as a
transition seam rather than as the final architecture.

## Consequences

### Positive

- subsystem dependencies become easier to inspect
- startup behavior becomes easier to evolve intentionally
- tests can configure one subsystem group at a time
- runtime packages become less dependent on implicit global environment

### Negative

- some transitional duplication may exist while old and new wiring coexist
- constructor and setup surfaces may temporarily grow during migration

## Implications

- wiring refactors should proceed subsystem by subsystem, not all at once
- grouped setup should be preferred over one-off individual global assignments
- global compatibility hooks should be removed only after the grouped path is
  validated
- `main.go` remains the composition root, but its job should become more
  declarative over time

## Related Documents

- `ARCHITECTURE-TARGET-STATE.md`
- `ARCHITECTURE-ROADMAP.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`
- `ADR-0001-MODULAR-MONOLITH-INCREMENTAL-REFACTORING.md`