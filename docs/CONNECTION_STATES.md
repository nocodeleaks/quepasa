# Connection States

This document explains the WhatsApp connection states exposed by QuePasa, how they are used by the health endpoint, and which states are currently emitted by the active runtime path.

## Overview

Connection states are defined in `src/whatsapp/whatsapp_connection_state.go`.

Terminology note:

- this document uses `session` as the preferred runtime term for one WhatsApp
	connected identity
- some code and compatibility surfaces still use historical `server` naming
	while the migration remains in progress

They are used to:

- represent the runtime status of each WhatsApp session
- expose session health in API responses
- distinguish intentional stop states from transport or authentication failures

## Health Semantics

QuePasa currently treats only the following states as healthy:

- `Ready`
- `Stopped`

This means:

- `Ready` is healthy because the session is connected, authenticated, and fully operational
- `Stopped` is healthy because it represents an intentional and stable stopped state, not a failure condition

All other states are treated as non-healthy by the health endpoint.

## State Reference

### `Unknown`

Fallback for invalid, missing, or unmapped values.

Current usage:

- returned by the session wrapper when the runtime reference itself is invalid or nil

### `UnPrepared`

The session exists, but there is no active connection object attached.

Typical situations:

- before the first start attempt
- after a connection object was disposed

Current usage:

- emitted by the session wrapper
- emitted by the whatsmeow status provider

### `UnVerified`

The session is not authenticated with WhatsApp yet.

Typical situations:

- before pairing
- before login completes
- when the session is no longer verified

Important:

- this is not, by itself, a transport failure

Current usage:

- emitted by the session wrapper
- emitted by the whatsmeow status provider when no authenticated client/session is available

### `Starting`

Reserved state intended for finer-grained lifecycle reporting during startup.

Current usage:

- defined in the enum
- not currently emitted by the active status calculation path

### `Connecting`

The client is trying to establish a session with WhatsApp servers.

Current usage:

- emitted by the whatsmeow status provider while `IsConnecting` is true

### `Stopping`

An intentional stop was requested, but the active connection is still being released.

Important:

- transitional state
- not a failure state

Current usage:

- emitted by the session wrapper when an intentional stop or delete flow is active and the transport is still connected

### `Stopped`

The session is intentionally offline after a stop request completed.

Important:

- stable state
- restartable state
- healthy state for the health endpoint
- not a failure state

Typical situations:

- manual user stop
- toggle stop
- controlled internal flow that calls `Stop()`, such as restart

Current usage:

- emitted by the session wrapper
- treated as healthy by the health endpoint

### `Restarting`

Reserved state intended for finer-grained lifecycle reporting during restart sequences.

Current usage:

- defined in the enum
- not currently emitted by the active status calculation path

### `Reconnecting`

Reserved state intended for a future explicit auto-reconnect status.

Current usage:

- defined in the enum
- not currently emitted by the active status calculation path

### `Connected`

The transport is established, but the session is not yet fully ready.

Typical situations:

- credentials are being loaded
- login is still completing
- the client may still be waiting for full authentication state

Current usage:

- emitted by the whatsmeow status provider when the transport is connected but the client is not fully logged in yet

### `Fetching`

Reserved state intended for finer-grained lifecycle reporting during initial synchronization or history fetch.

Current usage:

- defined in the enum
- not currently emitted by the active status calculation path

### `Ready`

The session is connected, authenticated, and fully operational.

Important:

- this is the main healthy runtime state

Current usage:

- emitted by the whatsmeow status provider
- treated as healthy by the health endpoint

### `Halting`

Reserved state intended for finer-grained lifecycle reporting during final shutdown.

Current usage:

- defined in the enum
- not currently emitted by the active status calculation path

### `Disconnected`

The connection to WhatsApp servers was lost or ended outside the intentional stopped flow.

Important:

- this is not considered healthy
- it differs from `Stopped`, which is intentional

Current usage:

- emitted by the whatsmeow status provider for non-intentional offline states

### `Failed`

The session entered an error state that prevented normal operation.

Current usage:

- emitted by the whatsmeow status provider when a connection or token failure is flagged

## States Currently Emitted

The active status calculation path currently emits these states:

- `Unknown`
- `UnPrepared`
- `UnVerified`
- `Connecting`
- `Stopping`
- `Stopped`
- `Connected`
- `Ready`
- `Disconnected`
- `Failed`

## States Currently Reserved

These states exist in the public enum but are not currently emitted by the active status calculation path:

- `Starting`
- `Restarting`
- `Reconnecting`
- `Fetching`
- `Halting`

They are kept to preserve room for more detailed lifecycle reporting in the future.

## Runtime Sources

The current state is composed from two layers:

1. The whatsmeow provider calculates transport/session states.
2. The QuePasa session wrapper adds intentional stop semantics on top of that provider.

This is why:

- `Connected`, `Ready`, `Disconnected`, and `Failed` come from the whatsmeow provider
- `Stopping` and `Stopped` are added by the QuePasa session wrapper through session intent handling

## Practical Interpretation

If you are integrating with QuePasa:

- treat `Ready` as online and fully available
- treat `Stopped` as intentionally offline but operationally OK
- treat `Stopping` as a temporary transitional state
- treat `Disconnected` and `Failed` as problem states requiring attention
- do not rely on reserved states until the runtime starts emitting them