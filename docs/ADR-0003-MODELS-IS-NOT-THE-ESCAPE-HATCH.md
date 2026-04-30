# ADR-0003: Models Is Not The Escape Hatch

## Status

Accepted

## Context

The `src/models` package grew into the main architectural pressure point of the
codebase.

Historically, it accumulated several kinds of concerns:

- domain entities
- runtime lifecycle behavior
- persistence-adjacent mutations
- compatibility helpers
- transport-adjacent shapes
- manager composition

The main architectural risk is not only file size. The deeper problem is that
`models` becomes the default placement when ownership is unclear.

That pattern causes two long-term problems:

1. package meaning becomes vague
2. future layering work becomes harder because ambiguity keeps growing faster
   than refactors can reduce it

## Decision

`src/models` must no longer be treated as the default escape hatch for new code
whose ownership has not been designed.

Only code that is clearly domain state or tightly bound to domain/runtime state
should be added there.

## Consequences

### Positive

- package ownership decisions become explicit earlier
- refactoring pressure on `models` can actually go down over time
- the application layer can emerge without competing with new accidental growth

### Negative

- some changes will require short design decisions before implementation
- contributors may need to create or strengthen adjacent packages instead of
  taking the most convenient local path

## Implications

- uncertain ownership must be resolved before placing new code
- workflow orchestration should move toward an explicit application layer
- transport DTOs should stay near transport layers
- compatibility helpers should remain visibly bounded and temporary where
  possible

## Related Documents

- `MODELS_REMODELING_AUDIT.md`
- `ARCHITECTURE-CURRENT-STATE.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`