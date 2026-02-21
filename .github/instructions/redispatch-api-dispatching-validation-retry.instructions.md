# Redispatch Instruction

## Scope
- Forced re-dispatch of cached messages via API.
- Endpoint: `POST /redispatch/{messageid}`.

## Source Paths
- `src/api/api_handlers+RedispatchController.go`
- `src/models/qp_whatsapp_server_extensions.go`
- `src/models/qp_dispatching_handler.go`

## Mandatory Validation Rules
- Reuse original dispatching pipeline validations.
- Apply TrackId/ForwardInternal loop prevention rules.
- Apply message type filters (groups, broadcasts, calls, read receipts).
- Respect per-dispatching endpoint configuration.

## Error Handling
- Missing `messageid`: return bad request.
- Message not in cache: return not found.
- Server unavailable: return service unavailable.

## Operational Notes
- Message IDs are normalized to uppercase in handler path.
- Redispatch must preserve message metadata and dispatch filters.
