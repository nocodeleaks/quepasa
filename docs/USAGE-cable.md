# Cable

## Objective

Define the main realtime transport as a websocket cable focused on a
simple pair of concepts:

- commands sent by the client
- events pushed by the backend

This module is intentionally separate from the legacy HTTP API and from the
legacy SignalR transport so we can evolve the SPA protocol without dragging
older integration constraints into the new channel.

## Route

- `GET /cable`

The transport is a backend capability independent of any specific frontend app.

There is only one official route:

- `/cable`

## Authentication

The cable endpoint reuses the same JWT already used by the form and frontend
app routes:

- cookie `jwt`
- or `Authorization: BEARER <token>`
- or `?jwt=<token>` when needed for non-browser clients

This works because `go-chi/jwtauth` verifier is used before the websocket
upgrade. The signing secret is still `SIGNING_SECRET`.

## Connection Model

The hub is connection-oriented, not user-oriented.

That means:

- one user can open many simultaneous websocket connections
- the same browser can open many tabs at the same time
- each connection keeps its own subscription set
- server events are fanned out only to subscribed connections
- lifecycle events are fanned out both to:
  - subscribers of the specific server
  - every live connection of the owning user

This keeps the transport scalable and predictable:

- account dashboards refresh on session lifecycle changes without manual polling
- heavy message streams stay opt-in through subscriptions

## Protocol

Inbound frames are commands:

```json
{
  "id": "cmd-123",
  "command": "subscribe",
  "data": {
    "tokens": ["server-token-1", "server-token-2"]
  }
}
```

Outbound frames are either responses or events.

Response:

```json
{
  "type": "response",
  "id": "cmd-123",
  "command": "subscribe",
  "ok": true,
  "data": {
    "subscriptions": ["server:server-token-1", "server:server-token-2"]
  },
  "timestamp": "2026-04-22T12:00:00Z"
}
```

Event:

```json
{
  "type": "event",
  "event": "server.message",
  "topic": "server:server-token-1",
  "data": {
    "token": "server-token-1",
    "user": "admin",
    "wid": "5511999999999",
    "state": "ready",
    "message": {
      "id": "ABCD",
      "text": "hello"
    }
  },
  "timestamp": "2026-04-22T12:00:00Z"
}
```

Error response:

```json
{
  "type": "response",
  "id": "cmd-123",
  "command": "server.enable",
  "ok": false,
  "error": {
    "code": "command_error",
    "message": "server token not owned by user"
  },
  "timestamp": "2026-04-22T12:00:00Z"
}
```

## Commands

### `ping`

Purpose:

- verify liveness
- recover current connection metadata

Response fields:

- `connectionId`
- `user`
- `subscriptions`

### `subscribe`

Purpose:

- subscribe the current connection to one or more server streams

Accepted payload shapes:

```json
{ "token": "server-token" }
```

```json
{ "tokens": ["server-token-a", "server-token-b"] }
```

```json
{ "topic": "server:server-token" }
```

Rules:

- ownership is validated against the persisted server record
- subscription is connection-local, not global to the user

### `unsubscribe`

Purpose:

- detach the current connection from one or more server streams

Same payload rules as `subscribe`.

### `server.enable`

Purpose:

- start or reconnect a server from the realtime channel

Payload:

```json
{ "token": "server-token" }
```

Rules:

- ownership is validated before start
- if the server is not currently loaded in memory, the backend can recreate it
  from the database through `GetOrCreateServerFromToken`

### `server.disable`

Purpose:

- stop a running server from the realtime channel

Payload:

```json
{ "token": "server-token" }
```

Rules:

- ownership is validated before stop
- the command targets the in-memory server instance

### `message.send`

Purpose:

- send a WhatsApp message through a server without going through HTTP

Payload example:

```json
{
  "token": "server-token",
  "chatId": "5511999999999@s.whatsapp.net",
  "text": "hello from cable"
}
```

Supported fields in the current implementation:

- `token`
- `id`
- `chatId` or `chatid`
- `trackId` or `trackid`
- `text`
- `inReply` or `inreply`
- `fileName` or `filename`
- `fileLength` or `filelength`
- `mime` or `mimeType`
- `seconds`
- `typingDuration`
- `mediaType`
- `url`
- `content`
- `poll`
- `location`
- `contact`

Rules:

- ownership is validated first
- the server must be in `ready` state
- URL and embedded base64 content are converted using the same model helpers used
  by the HTTP send flow

### `message.edit`

Purpose:

- edit a cached message through the realtime channel

Payload:

```json
{
  "token": "server-token",
  "messageId": "ABCD",
  "content": "updated text"
}
```

Rules:

- ownership is validated first
- the server must be in `ready` state
- the command reuses the current message edit semantics from `develop`

### `message.revoke`

Purpose:

- revoke a cached message through the realtime channel

Payload:

```json
{
  "token": "server-token",
  "messageId": "ABCD"
}
```

Rules:

- ownership is validated first
- the server must be in `ready` state

### `chat.archive`

Purpose:

- archive or unarchive a chat through the realtime channel

Payload:

```json
{
  "token": "server-token",
  "chatId": "5511999999999@s.whatsapp.net",
  "archive": true
}
```

Rules:

- ownership is validated first
- the server must be in `ready` state
- the command uses the same archive logic as the current HTTP API

### `chat.presence`

Purpose:

- send typing/presence updates through the realtime channel

Payload:

```json
{
  "token": "server-token",
  "chatId": "5511999999999@s.whatsapp.net",
  "type": "text",
  "duration": 3000
}
```

Rules:

- ownership is validated first
- the server must be in `ready` state
- non-paused presence keeps the same timeout-based auto-pause behavior used by the backend

## Events

### `session.ready`

Emitted immediately after the websocket is accepted.

Purpose:

- confirm authentication worked
- expose the assigned connection id
- let the client know the command surface expected by the backend

### `server.message`

Emitted for subscribed server topics whenever the backend dispatching handler
accepts a WhatsApp message/event for live delivery.

Topic:

- `server:{token}`

### `server.connected`
### `server.disconnected`
### `server.logged_out`
### `server.stopped`
### `server.deleted`

Emitted from lifecycle hooks in `models.DispatchingHandler`.

Terminology note:

- the websocket protocol keeps `server.*` event names for compatibility
- these lifecycle events refer to the WhatsApp session runtime, not to the web
  server process

Delivery rules:

- to every connection of the owning user
- to every connection subscribed to `server:{token}`

These events are separate from `server.message` because lifecycle transitions are
important enough to drive UI state even when there is no raw chat message to show.

## Backend Architecture

The transport is intentionally split in two layers:

### Model Layer

The model layer now exposes a transport-neutral realtime publisher registry:

- `RegisterRealtimePublisher`
- `PublishRealtimeServerMessage`
- `PublishRealtimeLifecycle`

This keeps WhatsApp core logic independent from websocket implementation
details.

### Cable Transport

The `cable` module:

- registers the `/cable` route in `webserver`
- authenticates the websocket upgrade
- manages active connections and subscriptions
- registers itself as a realtime publisher in the model layer

That separation is what allows the module to stay isolated without creating a
package cycle between `models`, `api`, and the transport.

## Why Not Reuse The Earlier Websocket Implementation

An earlier websocket implementation exists, but it is not a good base for the
current objective.

Reasons:

- it is centered on QR/verification flow, not on a durable command/event bus
- it does not define a stable protocol for session commands
- it is not structured for multiple concurrent user connections as the main
  realtime transport
- it would need deep adaptation anyway, so a dedicated module is cleaner

## Current Limits

This first version is intentionally narrow:

- it does not replace the legacy SignalR path yet
- it still does not expose every HTTP mutation via websocket commands
- it currently focuses on session lifecycle plus the main message/chat actions

That is deliberate. The protocol is now stable enough to expand without another
transport rewrite.
