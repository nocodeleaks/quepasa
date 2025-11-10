# QuePasa AI Agent Instructions

## Communication Guidelines
- **Response Language**: Always respond in the same language as the user's query (Portuguese for Portuguese queries, English for English queries)
- **Code and Comments**: All code, comments, documentation, and technical content must be in English
- **Consistency**: Maintain language consistency within each response type

## Architecture
- Go-based WhatsApp bot platform with HTTP APIs
- Whatsmeow library integration (`go.mau.fi/whatsmeow`)
- Message flow: `WhatsmeowHandlers` → `QPWhatsappHandlers` → `Webhook/RabbitMQ/Dispatching`
- Multi-layered APIs: v1, v2, v3 + non-versioned (latest) routes
- Modular packages in `src/` subdirectories

## Core Components
1. Connection: `whatsmeow/WhatsmeowConnection`
2. Handlers: `WhatsmeowHandlers` → `QPWhatsappHandlers`
3. Processing: Cache → Trigger → Webhooks/RabbitMQ
4. API: REST endpoints in `api/` with `api_handlers+*Controller.go` pattern
5. Server: `QpWhatsappServer` coordinates all operations

## Identifiers
- **JId**: `types.JID` from whatsmeow
- **WId**: String format
- **LId**: Local identifier (default, hides phone numbers)

## Documentation Structure
- **AGENTS.md**: Module-specific AI agent instructions (check each package)
- **README.md**: Human-readable documentation
- **USAGE-*.md**: Usage instructions for scripts/specific code
- **copilot-instructions.md**: Global AI agent guidelines (this file)

## Key Files
- `src/main.go`: System initialization
- `src/models/qp_defaults.go`: `QpVersion` constant
- `src/environment/`: Environment variable management
- `src/api/api_handlers.go`: Latest REST endpoints
- `src/whatsmeow/whatsmeow_handlers.go`: WhatsApp event processing
## Packages
- **api**: REST API, GraphQL, gRPC endpoints and controllers
- **audio**: Media processing (conversion, extraction)
- **environment**: Environment variables and configuration (9 categories, 47 vars)
- **form**: Form handling and validation
- **library**: Reusable utilities (Go packages only, no third-party)
- **metrics**: Prometheus monitoring and metrics
- **models**: Data structures and business logic
- **rabbitmq**: Message queueing and async processing
- **sipproxy**: SIP proxy server
- **webserver**: HTTP server, routing, middleware, forms, websockets (check AGENTS.md for details)
- **whatsapp**: WhatsApp abstractions and interfaces
- **whatsmeow**: Whatsmeow library integration

## Naming Conventions
- Extension methods: use `source` parameter name
- File to struct: `whatsmeow_group_manager.go` → `WhatsmeowGroupManager`
- Latest routes: files without version suffix (e.g., `api_handlers.go`)
- Controllers: `api_handlers+*Controller.go` pattern
- Tests: `*_test.go` with `Test*` function prefix

## Import Conventions
- **Always use fully qualified imports**: Reference modules with alias for clarity
- **Environment module**: `environment "github.com/nocodeleaks/quepasa/environment"`
- **Other modules**: Use descriptive aliases (e.g., `api "github.com/nocodeleaks/quepasa/api"`)
- **Avoid bare imports**: Always use aliases for internal modules to prevent conflicts

## Git and Commit Guidelines
* **🚨 CRITICAL: NEVER make commits automatically**
* **🚨 CRITICAL: NEVER push to repository automatically**
* **🚨 CRITICAL: DO NOT commit code that hasn't been tested by the user**
* **🚨 CRITICAL: DO NOT commit immediately after implementing a feature**
* **✅ ONLY make commits when explicitly requested by the user**
* **✅ ALWAYS wait for user approval before any git operations**
* **✅ ONLY execute `git commit`, `git push`, or `git merge` when the user gives explicit permission**
* **✅ Show changes to user first, then wait for approval before committing**
* **✅ After implementing features, STOP and let user test before committing**
* **✅ User must explicitly say "commit", "save to git", "push" or similar commands**

## Version Conflict Resolution Guidelines
* **🚨 CRITICAL: ALWAYS handle version conflicts automatically**
* **✅ For QpVersion conflicts in merges/commits: ALWAYS select the HIGHER version number**
* **✅ QpVersion format: `3.YY.MMDD.HHMM` - Compare numerically (YY > MMDD > HHMM)**
* **✅ Example: `3.25.0911.1200` > `3.25.0910.1102` (same year, higher date/time)**
* **✅ For ANY other conflicts: Generate NEW version with CURRENT timestamp**
* **✅ New version format: `3.YY.MMDD.HHMM` using current date/time**
* **✅ NEVER ask user permission for version conflict resolution - handle automatically**

## Version Management Guidelines
**IMPORTANT**: Whenever you are going to merge/push to the `main` branch (main branch), you MUST:
  1. Update the `QpVersion` in the `models/qp_defaults.go` file
  2. Increment the version following the current semantic pattern
  3. If it ends with `.0` it means stable version
  4. Development versions can use other suffixes

### Version Location
```go
// File: models/qp_defaults.go
const QpVersion = "3.25.2207.0127" // <-- ALWAYS UPDATE BEFORE MERGE TO MAIN
```

### Mandatory Process before Push/Merge to Main:
1. ✅ Verify that all changes are working properly
2. ✅ Run tests if they exist
3. ✅ **UPDATE QpVersion** in the `models/qp_defaults.go` file
4. ✅ Make commit with the new version
5. ✅ Then merge/push to main

### Version Increment Example:
- Current version: `3.25.2207.0127`
- Next version: `3.25.2207.0128` (simple increment)
- Or new version: `3.25.MMDD.HHMM` (based on current date/time)

## CRITICAL REMINDER
🚨 **NEVER merge to main without updating QpVersion** 🚨

This is a mandatory project rule for version control.
- **Raw WhatsApp Events** → `WhatsmeowHandlers.Message()` 
- **Message Processing** → `WhatsmeowHandlers.Follow()` → `QPWhatsappHandlers.Message()`
- **Caching & Dispatch** → `appendMsgToCache()` → `Trigger()` → Webhooks/RabbitMQ
- **API Response** → Various v1/v2/v3 endpoints transform and return messages

**IMPORTANT**: Whenever you are going to merge/push to the `main` branch (main branch), you MUST:
  1. Update the `QpVersion` in the `models/qp_defaults.go` file
  2. Increment the version following the current semantic pattern
  3. If it ends with `.0` it means stable version
  4. Development versions can use other suffixes

### Version Location
```go
// File: models/qp_defaults.go
const QpVersion = "3.25.2207.0127" // <-- ALWAYS UPDATE BEFORE MERGE TO MAIN
```

### Mandatory Process before Push/Merge to Main:
1. ✅ Verify that all changes are working properly
2. ✅ Run tests if they exist
3. ✅ **UPDATE QpVersion** in the `models/qp_defaults.go` file
4. ✅ Make commit with the new version
5. ✅ Then merge/push to main

### Version Increment Example:
- Current version: `3.25.2207.0127`
- Next version: `3.25.2207.0128` (simple increment)
- Or new version: `3.25.MMDD.HHMM` (based on current date/time)

## CRITICAL REMINDER
🚨 **NEVER merge to main without updating QpVersion** 🚨

This is a mandatory project rule for version control.

## Message Processing Flow (Critical Understanding)
- **Raw WhatsApp Events** → `WhatsmeowHandlers.Message()` 
- **Message Processing** → `WhatsmeowHandlers.Follow()` → `QPWhatsappHandlers.Message()`
- **Caching & Dispatch** → `appendMsgToCache()` → `Trigger()` → Webhooks/RabbitMQ
- **API Response** → Various v1/v2/v3 endpoints transform and return messages

## Key Integration Points
- **Connection Management**: `WhatsmeowConnection` wraps whatsmeow client with QuePasa abstractions
- **Event Handlers**: `WhatsmeowHandlers.EventsHandler()` dispatches to specific message/receipt/call handlers
- **Server Coordination**: `QpWhatsappServer` manages connection lifecycle and message routing
- **Environment System**: 47 variables across 9 categories (Database, API, WebServer, WhatsApp, etc.)
