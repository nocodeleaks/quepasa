# QuePasa Architecture Execution Checklist

## Purpose

This document converts the architecture roadmap into an implementation-oriented
checklist.

It is meant to be practical:

- short enough to use during active work
- explicit enough to guide branch-level execution
- strict enough to prevent architectural backsliding

## How To Use This Checklist

For each architecture task:

1. confirm which phase it belongs to
2. verify the entry criteria
3. implement only the bounded scope for that phase
4. validate the done criteria before moving on

## Phase A: Stop Structural Regression

### Objective

- stop making package ownership worse during ordinary feature work

### Entry Criteria

- a file or package is being touched for feature or bugfix work
- ownership is known to be overloaded or unclear

### Execution Checklist

- [ ] Do not add new workflow logic to `src/models` unless it is clearly domain
      state behavior
- [ ] Do not add new HTTP DTOs outside API-facing ownership
- [ ] Do not add new generic helper files such as `utils.go` or `helpers.go`
- [ ] Keep touched files within the file-size discipline from
      `CODE_ORGANIZATION.md` when practical
- [ ] If compatibility is required, isolate it visibly instead of mixing it into
      the main implementation path

### Done Criteria

- no new ambiguity was introduced into package ownership
- no new growth moved the codebase away from the documented target

## Phase B: Make The Application Layer Explicit

### Objective

- move lifecycle workflows out of implicit ownership and into explicit
      orchestration

### Candidate Work Items

- [ ] start session use case
- [ ] stop session use case
- [ ] restart session use case
- [ ] pair session use case
- [ ] delete session use case
- [ ] send message use case
- [ ] restore or sync history use case

### Entry Criteria

- workflow logic currently lives in oversized runtime entity files or mixed
      packages

### Execution Checklist

- [ ] Define the workflow boundary before moving code
- [ ] Move orchestration first, not low-level helpers first
- [ ] Keep entity state and pure domain logic in `models`
- [ ] Move infrastructure calls behind explicit dependencies where practical
- [ ] Validate behavior with narrow tests before broadening scope

### Done Criteria

- the extracted workflow can be explained without saying it "just lives in
      models"
- the new code has one clear orchestration responsibility

## Phase C: Shrink `models` By Responsibility

### Objective

- reduce conceptual overload inside `src/models`

### Candidate Work Items

- [ ] move persistence-heavy behavior behind store-facing components
- [ ] remove residual transport DTO ownership from `models`
- [ ] reduce compatibility wrappers that no longer need to stay in `models`
- [ ] isolate manager creation or coordination where explicit services can own it

### Entry Criteria

- package ownership is clear enough to move code safely

### Execution Checklist

- [ ] Confirm the moved code is not actually domain state behavior
- [ ] Move one responsibility slice at a time
- [ ] Preserve backward compatibility only where active call sites still require it
- [ ] Avoid large cross-package moves without focused validation

### Done Criteria

- `models` contains less mixed orchestration and less transport-adjacent logic

## Phase D: Simplify Composition Root Wiring

### Objective

- reduce broad startup global wiring in `main.go`

### Candidate Work Items

- [ ] group RabbitMQ-related startup wiring
- [ ] group realtime-related startup wiring
- [ ] group dispatch-related startup wiring
- [ ] replace scattered global assignment with explicit setup structs where safe

### Entry Criteria

- the subsystem dependencies are known and stable enough to group

### Execution Checklist

- [ ] Start with one subsystem group, not all at once
- [ ] Prefer explicit grouped wiring over ad-hoc global sprawl
- [ ] Preserve compatibility where constructor migration is not yet feasible
- [ ] Verify startup behavior after each extraction

### Done Criteria

- startup dependencies for the extracted subsystem are easier to identify and
      test

## Phase E: Contain HTTP Compatibility Surface

### Objective

- reduce the long-term cost of legacy API breadth

### Candidate Work Items

- [ ] document canonical routes versus aliases
- [ ] keep compatibility logic inside API-specific ownership
- [ ] retire obsolete aliases when consumers are known and safe to migrate

### Entry Criteria

- an API surface is being changed or documented

### Execution Checklist

- [ ] Do not spread compatibility logic into unrelated packages
- [ ] Keep canonical behavior easy to identify
- [ ] Validate route behavior narrowly with route or controller tests

### Done Criteria

- compatibility remains contained rather than expanding further into the system

## Phase F: Keep Adapters Thin

### Objective

- ensure boundary packages stay as adapters, not alternate business cores

### Candidate Work Items

- [ ] continue splitting `src/whatsmeow` by event family or capability
- [ ] keep delivery policy separate from delivery mechanics
- [ ] translate external events into application behaviors rather than embedding
      more workflow there

### Entry Criteria

- a boundary package is growing in size or responsibility

### Execution Checklist

- [ ] Confirm whether the code is translation logic or orchestration logic
- [ ] Keep adapter files scoped to external protocol or SDK concerns
- [ ] Move business coordination inward when it grows beyond simple translation

### Done Criteria

- adapter packages are easier to explain as boundaries rather than runtime cores

## Phase G: Finish The Server To Session Transition

### Objective

- complete the semantic migration only after boundaries are cleaner

### Candidate Work Items

- [ ] keep compatibility aliases during active migration
- [ ] update primary runtime terminology to `session`
- [ ] remove legacy `server` terminology from WhatsApp identity flow when safe

### Entry Criteria

- ownership is stable enough that the rename will express a real structural
      truth

### Execution Checklist

- [ ] move responsibility first, rename second
- [ ] avoid cosmetic renames that leave the same architectural confusion behind
- [ ] preserve public compatibility where still required

### Done Criteria

- runtime vocabulary matches actual responsibility
- `server` primarily refers to infrastructure, not WhatsApp identity lifecycle

## Cross-Cutting Validation Checklist

Use this after every architecture task.

- [ ] The change reduced or contained coupling
- [ ] The change did not widen `models` again
- [ ] The change did not spread transport DTOs into the wrong layer
- [ ] The change preserved documented architecture direction
- [ ] The change was validated with the narrowest useful tests or build checks
- [ ] Any new compatibility layer was made explicit and scoped

## Practical Rule Of Thumb

If a change does not improve at least one of these, it is probably not a strong
architecture task:

- dependency direction
- ownership clarity
- package pressure
- compatibility containment
- workflow explicitness

## Related Documents

- `ARCHITECTURE-INDEX.md`
- `ARCHITECTURE-CURRENT-STATE.md`
- `ARCHITECTURE-PACKAGE-MAP.md`
- `ARCHITECTURE-TARGET-STATE.md`
- `ARCHITECTURE-ROADMAP.md`
- `PLAN-ARCHITECTURE-REFACTORING.md`
- `CODE_ORGANIZATION.md`