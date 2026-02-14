# Webhooks Instruction

## Scope
- Webhook dispatch behavior, health checks, and operational monitoring.

## Runtime Inputs
- Environment variable: `WEBHOOK_TIMEOUT`.
- Webhook endpoints and dispatching configuration come from server config.

## Operational Endpoints
- `GET /health`
- `GET/POST/DELETE /webhook`

## Metrics Focus
- `quepasa_webhooks_sent_total`
- `quepasa_webhook_send_errors_total`
- `quepasa_webhook_timeouts_total`
- `quepasa_webhook_success_total`

## Mandatory Rules
- Keep webhook processing direct and deterministic.
- Preserve timeout behavior and error counters.
- Keep health endpoint behavior stable for operations.
- Validate webhook behavior after dispatching changes.
