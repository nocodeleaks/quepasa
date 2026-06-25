# ADR-0002: Session As The Primary Runtime Concept

## Status

Accepted

## Context

The historical central runtime object for one WhatsApp connected identity has
been named with `server` terminology.

That naming became misleading over time because the object does not represent an
infrastructure server. It represents the lifecycle of one WhatsApp identity
within the process.

This ambiguity creates confusion with real infrastructure concerns such as:

- web server
- SignalR server or hub host
- SIP server
- general process-level hosting concerns

Recent code evolution already introduced a compatibility bridge through
`QpWhatsappSession` naming while preserving the existing runtime surface.

## Decision

QuePasa will treat `session` as the preferred domain and runtime term for one
WhatsApp connected identity.

`server` terminology remains valid only for infrastructure concerns and
temporary compatibility layers during migration.

## Consequences

### Positive

- runtime terminology becomes more semantically correct
- architecture discussions become clearer
- future multi-session reasoning becomes less ambiguous

### Negative

- compatibility wrappers and aliases are needed during transition
- some documentation and code will remain mixed until migration is completed

## Implications

- move responsibility first, rename second
- do not perform broad cosmetic renames that preserve the same structural
  confusion
- keep infrastructure packages using `server` where that term is accurate
- prefer `session` in new runtime documentation and new runtime-facing APIs when
  safe

## Related Documents

- `ARCHITECTURE-TARGET-STATE.md`
- `ARCHITECTURE-ROADMAP.md`
- `PLAN-ARCHITECTURE-REFACTORING.md`