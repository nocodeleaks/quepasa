# Code Organization Guidelines

This document defines the project rules for file size, responsibility distribution, and method placement.

The goal is to keep the codebase easier to read, review, and evolve without concentrating too much behavior in a single file.

## Core Rule

Source files should be kept at **400 lines maximum**.

Recommended target:

- ideal range: `150-300` lines
- acceptable range: `300-400` lines
- refactor required: `> 400` lines

This is a structural rule, not just a style preference.

## Responsibility Rule

A file should group code by **one clear responsibility**.

Examples:

- state calculation
- deletion flow
- dispatching flow
- persistence helpers
- serialization helpers
- validation helpers

Avoid "god files" that mix:

- state
- persistence
- lifecycle
- webhook dispatching
- formatting
- helper utilities

in the same file.

## Method Placement Rule

Keep a method attached to a receiver only when it truly depends on:

- receiver state
- receiver invariants
- private lifecycle coordination
- behavior that conceptually belongs to that type

Move code out of the receiver file when it only needs:

- primitive values such as `wid`, `token`, `timestamp`
- slices, maps, or payloads passed as arguments
- generic transformation logic
- shared package-level behavior

If a function does not need the full receiver, do not pass the full receiver.

Example:

- if a dispatch helper only needs `wid` and `[]*QpDispatching`, it should not depend on `*QpWhatsappServer`

## File Splitting Rule

When a type grows, split it by behavior, not arbitrarily.

Preferred pattern:

- `qp_whatsapp_server.go`: core type definition and minimal constructor/wiring
- `qp_whatsapp_server_state.go`: state calculation and status helpers
- `qp_whatsapp_server_delete.go`: delete flow and delete-specific helpers
- `qp_whatsapp_server_send.go`: send/message flow
- `qp_whatsapp_server_dispatching.go`: dispatching-related behavior already owned by the server
- `qp_whatsapp_server_extensions.go`: package-level helpers that are not exclusive to the type

Do not create generic files such as:

- `utils.go`
- `helpers.go`
- `misc.go`
- `common.go`

unless the name is explicitly scoped to a bounded responsibility.

## Extraction Criteria

Extract code from a file when any of the following becomes true:

- the file exceeds `400` lines
- a method depends more on arguments than on receiver state
- multiple unrelated flows exist in the same file
- helper functions are reusable by more than one caller
- a reviewer must scroll too much to understand a single concern

## Receiver Design Guidance

Prefer receiver methods for:

- `Start()`
- `Stop()`
- `Delete()`
- methods that mutate server state
- methods that coordinate connection, handler, and persistence together

Prefer package-level functions for:

- payload builders
- cloning helpers
- filtering helpers
- dispatch loops that operate from explicit arguments
- formatting/translation helpers

## Naming Guidance

Names should reveal responsibility directly.

Prefer:

- `BuildServerDeletedEvent`
- `CloneDispatchings`
- `PostToDispatchings`
- `RestoreDispatchingsAfterDeleteFailure`

Avoid vague names such as:

- `Handle`
- `Process`
- `Helper`
- `DoStuff`

unless the surrounding type already gives precise context.

## Review Checklist

Before merging a change, verify:

- the touched file stays at or below `400` lines
- the file has one clear responsibility
- methods that do not need the full receiver were extracted
- helpers were placed in a responsibility-specific file
- naming reflects behavior, not implementation accident
- new logic did not increase coupling unnecessarily

## Refactoring Policy

When touching an oversized file:

1. do not add more unrelated behavior into it
2. extract the new behavior first if practical
3. prefer small responsibility-driven files over one large central file
4. reduce coupling while preserving existing behavior

Existing oversized files may be refactored incrementally, but new work should move the codebase toward this standard, not away from it.
