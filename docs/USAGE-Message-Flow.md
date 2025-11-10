# Message Processing Flow

This document describes how messages flow through the QuePasa system from WhatsApp to webhooks/RabbitMQ.

## Flow Overview

1. **Raw WhatsApp Events** → `WhatsmeowHandlers.Message()`
2. **Message Processing** → `WhatsmeowHandlers.Follow()` → `QPWhatsappHandlers.Message()`
3. **Caching & Dispatch** → `appendMsgToCache()` → `Trigger()` → Webhooks/RabbitMQ
4. **API Response** → Various v1/v2/v3 endpoints transform and return messages

## Components

### WhatsmeowHandlers
- Receives raw events from WhatsApp via whatsmeow library
- Entry point: `WhatsmeowHandlers.Message()`
- Processes events and forwards to QuePasa handlers

### QPWhatsappHandlers
- Processes messages after whatsmeow layer
- Entry point: `QPWhatsappHandlers.Message()`
- Applies business logic and filtering

### Caching & Triggering
- `appendMsgToCache()`: Stores messages in cache
- `Trigger()`: Dispatches messages to configured webhooks and RabbitMQ

### API Endpoints
- v1, v2, v3 endpoints provide different message formats
- Non-versioned endpoints (latest) use current format
- Transform internal message format to API responses
