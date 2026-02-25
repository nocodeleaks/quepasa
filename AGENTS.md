# SPA Sync From chat.aireset.com.br

## task objective
Sync SPA updates from `chat.aireset.com.br/src` into `quepasa/src` with full scope (backend SPA + frontend + webserver fallback), versioning frontend assets.

## mandatory checklist
- [x] Create dedicated feature branch
- [x] Sync SPA backend API routes/controllers/login
- [x] Sync form handlers for SPA-first flow
- [x] Sync webserver SPA fallback behavior
- [x] Sync environment SPA-related settings
- [x] Sync frontend source (`src/frontend`)
- [x] Sync built assets (`src/assets/frontend`)
- [x] Regenerate Swagger docs
- [x] Validate build

## current status
Branch `feature/spa-sync-from-chat-20260224` with SPA sync completed. Backend build, frontend build, and Swagger generation executed successfully. Local runtime smoke tests executed with app running on port `31000`.

## next steps
Await user validation/QA in runtime and follow-up adjustments if necessary. Keep temporary runtime env override notes for local startup (`USER`/`PASSWORD`) until seed config is aligned.

## immutable constraints discovered during execution
- Do not commit or push without explicit user request.
- Keep API prefix aligned for SPA usage (`/api`).
- Preserve SPA auth flow consistency (login/session/cookie JWT).
- Keep swagger in sync after API route/controller changes.
