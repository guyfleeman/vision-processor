# vision-processor-gui-frontend

Browser UI for the vision-processor GUI. Svelte 5 + TypeScript + Vite.

Currently a skeleton: connects to a WebSocket at `/ws` and asks the
backend (via `GET /api/snapshots`) which debug images currently exist
on disk (only meaningful when the browser and the vision processor
share a filesystem -- see the root [README.md](../../README.md)).

## Run

Two options:

```
# Standalone dev server with HMR (proxies /api and /ws to the Go host on :8085)
cd gui/frontend
npm install
npm run dev
```

```
# Through the Go host itself, on its own port, no proxy involved
cd gui
make run
```

Either way the Go host in `gui/` must be running for the connection
badge and snapshot grid to show anything.

## Architecture

- `src/lib/wrapper-bus.ts` — single `WebSocket` client. Exposes
  `connectionState` (Svelte store) and `topic<T>(name)` (returns a
  store of the latest message). Subscribes to a topic lazily on first
  reader, unsubscribes when the last reader goes away. Reconnects on
  close with exponential backoff (1s → 30s).
- `src/App.svelte` — placeholder UI: connection badge + a grid of
  `<img>` tags, one per entry returned by `GET /api/snapshots`. The list is
  refreshed every 5 s; each `<img>` is refreshed once per second via a
  cache-busting `?t=<ms>` query. Plus a dev panel with a subscribe
  toggle + JSON dump for `wrapper_packet.out`.
- `src/main.ts` — mounts `App` into `#app`.

The WS wire format (`gui/internal/hub`):

```jsonc
// client -> server
{ "action": "subscribe",   "topic": "wrapper_packet.out" }
{ "action": "unsubscribe", "topic": "wrapper_packet.out" }
// server -> client
{ "topic": "wrapper_packet.out", "data": { ... } }
```

`wrapper_packet.out` carries the current `SSL_WrapperPacket` as
canonical protojson, republished once per second.

Snapshot endpoints are plain HTTP: `GET /api/snapshots` returns the
list of available `{cam_id, view}` entries as JSON;
`GET /api/snapshot/<cam_id>/<view>` returns the actual `image/jpeg` or
`image/png` (or 404 if missing).

## Scripts

```
npm run dev           # Vite dev server with HMR
npm run build         # production build to dist/
npm run preview       # serve the production build locally
npm run check         # svelte-check + tsc (type-check everything)
npm run lint          # eslint over src/
npm run format        # prettier --write .
npm run format:check  # prettier --check . (CI-style)
```

TypeScript is configured strict (`strict`,
`noUncheckedIndexedAccess`, `noImplicitOverride`,
`noPropertyAccessFromIndexSignature`, `noImplicitReturns`,
`noFallthroughCasesInSwitch`).

## Production serving

`npm run build` produces `dist/`, which the Go host in `gui/` embeds
directly into its binary (`gui/frontend/embed.go`) and serves on its
own port, with a fallback to `index.html` for client-side routes. See
`make run` in the root [README.md](../../README.md).
