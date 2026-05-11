# Frontend Apps Architecture

This document defines how QuePasa frontend applications are organized under `src/apps/` and how the backend discovers and serves them.

## Purpose

QuePasa supports multiple frontend applications for the same API and backend runtime.

Each app may represent:

- a different product experience
- a different UX strategy
- a different operational workflow
- a different interpretation of the same system

The backend must treat each app as an independent frontend, not as a variant or alias of another app.

## Core Rule

Every directory under `src/apps/<slug>` is its own application.

Examples:

- `src/apps/vuejs`
- `src/apps/console`
- `src/apps/mobile-admin`

The `<slug>` is the public app identifier and maps directly to the URL prefix:

```text
/apps/<slug>
```

Examples:

```text
/apps/vuejs/
/apps/console/
/apps/mobile-admin/
```

## Isolation Rules

Frontend apps must stay isolated by slug.

That means:

- no implicit redirect from one app slug to another
- no automatic fallback from one app slug to another
- no semantic coupling such as `console -> vuejs`
- no backend logic that assumes one app replaces another app
- no route resolution based on app name meaning

If `src/apps/console` exists, `/apps/console` must serve that app.

If `src/apps/vuejs` exists, `/apps/vuejs` must serve that app.

If both exist, both must coexist independently.

## Standard URL Contract

The backend exposes frontend apps using exact slug matching:

```text
/apps/<slug>
/apps/<slug>/
/apps/<slug>/*
```

Client-side navigation falls back to that app's `index.html`.

Static assets are resolved inside the selected app only.

Example:

- `/apps/console/assets/index.js` must come from `console`
- `/apps/vuejs/assets/index.js` must come from `vuejs`

## App Discovery Rules

The backend app discovery lives in:

- `src/webserver/webserver.go`

It scans:

```text
src/apps/*
```

For each app directory, the backend resolves the public directory using the following order:

1. `dist/index.*`
2. root `index.*` for legacy prebuilt bundles
3. `client/index.*` when frontend dev proxy mode is enabled

## Recommended Source App Layout

For actively developed frontend apps, use this layout:

```text
src/apps/<slug>/
  client/
    index.html
    src/
  dist/
  package.json
  vite.config.ts
  README.md
```

Meaning:

- `client/` contains source files
- `dist/` contains the published build output
- `package.json` marks the app as a source-based frontend project

## Legacy Bundle Layout

Older or imported frontend apps may exist only as prebuilt static bundles:

```text
src/apps/<slug>/
  index.html
  assets/
```

This is still supported.

In that case, the root directory itself is the published app.

## Dev Proxy Behavior

When `QUEPASA_DEV_FRONTEND=1` is enabled, the backend may proxy requests for a source-based app to its frontend dev server.

This is intended for active frontend development.

There is no default alias for app slugs.

Each app must be accessed explicitly through its own slug under `/apps/<slug>`.

## Backend Responsibilities

The backend is responsible for:

- discovering available apps
- serving each app under `/apps/<slug>`
- serving static files from the selected app
- falling back to the selected app's `index.html` for SPA navigation

The backend is not responsible for:

- deciding that one app should behave like another
- rewriting one app into another based on slug naming
- sharing frontend state or assets across apps

## Frontend Responsibilities

Each app is responsible for its own:

- routing base
- asset paths
- build output
- UI semantics
- navigation model
- API consumption strategy

If an app loads the wrong UI, first verify the actual contents of that app directory.

In many cases the backend is serving the correct app directory, but the app bundle itself is a copied or outdated build.

## Operational Guidance

When creating a new frontend app:

1. create a new directory under `src/apps/<slug>`
2. keep the app self-contained
3. configure the app to build and run under `/apps/<slug>/`
4. do not reuse another app slug
5. do not add backend alias logic to make the app work

When migrating an app:

1. move its source into `src/apps/<slug>`
2. update its build base to `/apps/<slug>/`
3. keep any compatibility alias explicit and temporary
4. avoid treating another app as its replacement unless the user explicitly wants that alias

## Current Reference Points

Relevant files:

- `src/webserver/webserver.go`
- `src/webserver/webserver_test.go`
- `src/apps/vuejs/README.md`
- `.github/copilot-instructions.md`

## Non-Goals

This document does not define:

- the internal architecture of a specific frontend app
- shared component libraries between apps
- branding strategy across apps
- frontend framework choices

Those decisions belong to each app unless another project document defines them.
