# Message Flow Instruction

## Scope
- End-to-end message path from WhatsApp event ingestion to dispatch outputs.

## Core Flow
- Entry: `WhatsmeowHandlers` receives raw events.
- Processing: `QPWhatsappHandlers` applies business rules.
- Cache: `DispatchingHandler.appendMsgToCache(...)` stores message.
- Dispatch: `DispatchingHandler.Trigger(...)` sends to webhooks/RabbitMQ.
- API layers transform internal model for v1/v2/v3/latest responses.

## Key Source Paths
- `src/whatsmeow/whatsmeow_handlers.go`
- `src/models/dispatching_handler.go`
- `src/api/`

## Mandatory Rules
- Preserve dispatching flow order when modifying handlers.
- Keep cache append and trigger behavior aligned with existing path.
- Validate downstream dispatch side effects after handler changes.
