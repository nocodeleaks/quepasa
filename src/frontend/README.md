Vue 3 frontend for QuePasa

Setup:
  cd src/frontend
  npm install
  npm run dev

Build:
  npm run build  # outputs to src/assets/frontend

Dev notes:
  - Vite dev server proxies /api and /form to backend (including websocket upgrades for /form/verify/ws). Configure backend with environment vars:

- VITE_BACKEND_URL (ex: http://localhost:32000) OR
- VITE_BACKEND_PORT (ex: 32000) — fallback to env WEBAPIPORT or 32000
- VITE_DEV_PORT (optional) — vite dev port (default 5173)

Examples:
  VITE_BACKEND_PORT=32000 npm run dev
  VITE_BACKEND_URL=http://backend.local:32000 npm run dev
  - API client is in src/services/api.ts
  - WebSocket service is in src/services/ws.ts

Dev helpers:
  Use project scripts to run backend + frontend together (from project root):

  ./scripts/start-dev.sh  # builds backend and starts backend + vite
  ./scripts/stop-dev.sh   # stop both services

Or let the Go binary spawn the frontend automatically by setting:

  QUEPASA_DEV_FRONTEND=1 ./src/.dist/quepasa
