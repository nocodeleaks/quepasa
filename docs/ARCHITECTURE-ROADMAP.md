# QuePasa Architecture Roadmap

## Purpose

This document translates the current architectural analysis into an execution
order that is realistic for an active production codebase.

It complements the refactoring plan by organizing the work into decision-ready
program steps.

## Guiding Principle

The roadmap should reduce coupling without pausing feature delivery.

That means:

- prefer iterative extractions over disruptive rewrites
- preserve compatibility when the surface is externally consumed
- pay down architectural debt where it reduces future change cost the most

## Priority Model

Changes should be prioritized by this order:

1. dependency direction improvements
2. ownership clarification of overloaded packages
3. API surface containment
4. naming cleanup after ownership is stable

## Phase A: Stop Structural Regression

Goal:

- prevent new work from increasing ambiguity in package ownership

Actions:

- enforce `docs/CODE_ORGANIZATION.md` for new and touched files
- avoid adding new workflow logic to `src/models` unless it is undeniably domain
  state behavior
- avoid adding new transport DTOs outside transport-facing packages
- keep new compatibility layers visibly scoped and documented

Expected outcome:

- the codebase stops drifting further away from the intended architecture

## Phase B: Make The Application Layer Real

Goal:

- move session workflows out of implicit ownership and into explicit use-case
  orchestration

Candidate workflow extractions:

- start session
- stop session
- restart session
- pair session
- delete session
- send message
- restore or sync history

Expected outcome:

- `models` becomes less overloaded
- behavior becomes easier to test by workflow rather than by oversized entity
  surface

## Phase C: Shrink `models` By Responsibility

Goal:

- reduce conceptual overload inside the current runtime core package

Actions:

- keep only domain state and clearly domain-bound helpers in `models`
- move persistence-heavy behavior behind store-facing components
- move transport projection and compatibility helpers out where possible
- reduce manager creation and orchestration pressure when explicit services can
  own the flow instead

Expected outcome:

- importing `models` becomes cheaper mentally and structurally

## Phase D: Simplify Composition Root Wiring

Goal:

- reduce reliance on broad global startup assignments in `main.go`

Actions:

- introduce grouped dependency structs for subsystems
- replace one-off global assignments gradually where constructor wiring is
  feasible
- keep compatibility seams only where migration cost is still too high

Expected outcome:

- startup dependencies become more explicit
- test setup becomes more localized

## Phase E: Contain HTTP Compatibility Surface

Goal:

- reduce maintenance cost of API legacy breadth without breaking users blindly

Actions:

- document canonical routes versus compatibility aliases
- concentrate versioned DTOs and compatibility behavior in API-facing packages
- retire aliases only when actual consumers are known and migration is safe

Expected outcome:

- API maintenance burden becomes easier to reason about
- route behavior becomes easier to validate systematically

## Phase F: Keep Adapters Thin

Goal:

- ensure `whatsmeow`, realtime, and outbound transport packages remain boundary
  modules rather than alternate business cores

Actions:

- continue event-handler extractions inside `src/whatsmeow`
- keep delivery policy separate from delivery mechanics
- translate external events into application behaviors instead of embedding more
  business coordination into adapter packages

Expected outcome:

- integration packages become easier to change or test independently

## Phase G: Finish The Server To Session Transition

Goal:

- make the runtime language match the true domain model

Actions:

- keep compatibility aliases while migration remains active
- move responsibility first, rename second
- retire `server` terminology from the WhatsApp identity lifecycle only after
  call sites and ownership are stable

Expected outcome:

- clearer architecture vocabulary
- less confusion between infrastructure servers and WhatsApp sessions

## Decision Guidance

When choosing the next architectural task, prefer work that satisfies at least
two of the following:

- removes a dependency-direction violation
- shrinks `models`
- reduces startup global wiring
- clarifies ownership of a major workflow
- reduces legacy compatibility spread

## Suggested Near-Term Sequence

The highest-value near-term sequence is:

1. stop new growth in `models`
2. extract explicit session use cases
3. reduce startup globals through grouped service wiring
4. continue API compatibility quarantine
5. continue adapter slimming in `whatsmeow`
6. finish semantic migration from `server` to `session`

This order matters because naming changes create durable value only after the
underlying ownership model is cleaner.

## Roadmap Conclusion

QuePasa does not need a platform rewrite.

It needs disciplined architectural follow-through.

The most important work is making the intended layering visible in package
ownership, dependency flow, and workflow orchestration.

## Related Documents

- `ARCHITECTURE-INDEX.md` is the entry point for this documentation set
- `ARCHITECTURE-CURRENT-STATE.md` explains the present architectural shape
- `ARCHITECTURE-TARGET-STATE.md` defines the desired destination
- `ARCHITECTURE-EXECUTION-CHECKLIST.md` converts roadmap items into execution
  checkpoints
- `ARCHITECTURE-PACKAGE-MAP.md` identifies the current package responsibilities
- `PLAN-ARCHITECTURE-REFACTORING.md` contains concrete refactor candidates and
  previously completed items