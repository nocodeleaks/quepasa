# Merge Workflow Instruction

## Scope
- Branch flow policy for publishing changes.
- Conflict resolution policy for version and dependencies.

## Branching Rules
- Default flow: implement and validate in `develop` first.
- Only after validation, merge `develop` into `main`.
- Do not publish directly to `main` unless explicitly requested by user.

## Version Rules
- File: `src/models/qp_defaults.go`
- Stable release must end with `.0`.
- For merge/version conflicts in `QpVersion`, keep the newest version by current task date/time.
- Do not downgrade `QpVersion` during merge resolution.

## Conflict Resolution Rules
- Prefer current task intent over unrelated incoming changes.
- If merge imports unrelated dependency updates, isolate and avoid publishing them with the task unless requested.
- Resolve only conflicts required by the current scope when possible.

## Dependency and WhatsApp Update Rules
- Do not bundle broad dependency updates with unrelated task merges.
- Treat whatsmeow/dependency version updates as separate integration scope.
- Integrate dependency updates in `develop` first, validate build/tests, then merge to `main`.
- Prefer controlled integration through PR review when dependency drift is large.

## Publish Rules
- Build before publish: `cd src` then `go build -o ../.dist/quepasa.exe`.
- Publish sequence:
  1. commit/push `develop`
  2. merge to `main`
  3. push `main`

## Instruction Priority
- For merge operations, this file is the source of truth.
- Keep this file updated when branch policy changes.
