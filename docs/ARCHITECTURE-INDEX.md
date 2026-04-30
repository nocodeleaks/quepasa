# QuePasa Architecture Index

## Purpose

This document is the entry point for the architecture documentation set under
`docs/`.

Use it when you need to understand:

- what the current architecture looks like
- what the target architecture should be
- what each main package currently owns
- what refactoring work should happen first

## Recommended Reading Order

### 1. Current Situation

Read first:

- `ARCHITECTURE-CURRENT-STATE.md`

Use this document when you want to understand the codebase as it is today.

### 2. Package-Level Map

Read next:

- `ARCHITECTURE-PACKAGE-MAP.md`

Use this document when you want to locate ownership, package pressure, and
structural risk by package.

### 3. Target Direction

Read after that:

- `ARCHITECTURE-TARGET-STATE.md`

Use this document when you need the recommended end state and dependency model.

### 4. Execution Sequence

Then read:

- `ARCHITECTURE-ROADMAP.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`

Use these documents when you are preparing actual implementation work.

### 5. Refactoring Inventory

For detailed refactor findings and already completed structural items, read:

- `PLAN-ARCHITECTURE-REFACTORING.md`
- `MODELS_REMODELING_AUDIT.md`
- `CODE_ORGANIZATION.md`

### 6. Formal Decisions

For stable architecture decisions and their rationale, read:

- `ARCHITECTURE-DECISIONS.md`
- `ADR-0001-MODULAR-MONOLITH-INCREMENTAL-REFACTORING.md`
- `ADR-0002-SESSION-AS-RUNTIME-CONCEPT.md`
- `ADR-0003-MODELS-IS-NOT-THE-ESCAPE-HATCH.md`
- `ADR-0004-EXPLICIT-APPLICATION-LAYER.md`
- `ADR-0005-GROUPED-COMPOSITION-ROOT-WIRING.md`

## Document Roles

### `ARCHITECTURE-CURRENT-STATE.md`

Role:

- explains the current architectural reality
- identifies the main structural bottlenecks
- highlights what is already working well

### `ARCHITECTURE-PACKAGE-MAP.md`

Role:

- maps the practical responsibility of the main packages
- identifies strengths, risks, and recommended direction package by package

### `ARCHITECTURE-TARGET-STATE.md`

Role:

- defines the desired layer model
- defines the intended dependency direction
- clarifies the future role of session, transport, store, and adapters

### `ARCHITECTURE-ROADMAP.md`

Role:

- defines the recommended order of change
- keeps the work incremental rather than disruptive

### `ARCHITECTURE-EXECUTION-CHECKLIST.md`

Role:

- translates the roadmap into concrete execution checkpoints
- can be used during implementation planning and branch-level task tracking

### `PLAN-ARCHITECTURE-REFACTORING.md`

Role:

- detailed technical inventory of refactor items
- tracks already completed architecture improvements

### `ARCHITECTURE-DECISIONS.md`

Role:

- entry point for formal architecture decisions
- links the active ADR set

### `MODELS_REMODELING_AUDIT.md`

Role:

- deep dive focused specifically on the `models` package problem

### `CODE_ORGANIZATION.md`

Role:

- coding-time structural discipline
- file size, responsibility, extraction, and placement rules

## Quick Navigation By Intent

If you want to understand the system now:

- read `ARCHITECTURE-CURRENT-STATE.md`

If you want to know where a package fits:

- read `ARCHITECTURE-PACKAGE-MAP.md`

If you want to know what architecture the project should move toward:

- read `ARCHITECTURE-TARGET-STATE.md`

If you want to choose the next refactor:

- read `ARCHITECTURE-ROADMAP.md`
- then use `ARCHITECTURE-EXECUTION-CHECKLIST.md`

If you are working specifically on `models`:

- read `MODELS_REMODELING_AUDIT.md`

If you are touching large files during implementation:

- read `CODE_ORGANIZATION.md`

If you need the rationale behind an architectural constraint:

- read `ARCHITECTURE-DECISIONS.md`
- then read the relevant ADR

## Conclusion

This architecture documentation set is meant to support incremental structural
improvement, not a one-time redesign.

When in doubt:

1. start from the current state
2. validate package ownership
3. check the target state
4. execute the roadmap through the checklist