# QuePasa AI Agent Instructions

## AI Memory First - EXECUTE BEFORE BROAD REPOSITORY EXPLORATION

- If the `Sufficit AI Memory` MCP server is available in VS Code, run memory recall before broad repository exploration, repo mapping, or speculative architecture reading.
- Build the first recall query from the current repository, branch, active file, user goal, and the most stable local anchors available: symbol, failing command, failing test, or concrete error.
- Use compact retrieval order: `memory_search` → `memory_timeline` → `memory_get_observations`.
- Save structured memory with `memory_save` using the schema in `schemas/vscode-memory-anchor-checkpoint.schema.json`.
- Save one `task-anchor` after initial routing and a `task-checkpoint` at major milestones: first local hypothesis, first substantive edit, focused validation, deploy/runtime verification, and handoff/summary boundaries.
- Persist compact task state only. Never dump raw chat transcripts, secrets, or oversized logs into memory.
- If the MCP server is unavailable or the current workspace is not configured for it, continue normally and do not block the task.

## AI Agent Startup Checklist - EXECUTE FIRST

**At the start of EVERY conversation or when resuming after summary:**

1. **Run the `AI Memory First` workflow above** whenever the `Sufficit AI Memory` MCP server is available
2. **Review current task context files** if available in the current directory
3. **Read relevant project documentation** based on task type
4. **Read `/.github/copilot-chat-vocabulary.md`** and apply its normalization rules in the full conversation

## Communication Guidelines
- Response Language: Always respond in the same language as the user's query (Portuguese for Portuguese queries, English for English queries)
- Code and Comments: All code, comments, documentation, and technical content must be in English
- Consistency: Maintain language consistency within each response type

## Architecture
- Go-based WhatsApp bot platform with HTTP APIs
- Whatsmeow library integration (go.mau.fi/whatsmeow)
- Message flow: WhatsmeowHandlers → QPWhatsappHandlers → Webhook/RabbitMQ/Dispatching
- Multi-layered APIs: v1, v2, v3 + non-versioned (latest) routes
- Modular packages in src/ subdirectories
- Each directory in `src/apps/<slug>` is an independent frontend app for the same QuePasa API; keep apps isolated by slug with no implicit aliasing, fallback, or semantic coupling between apps

## Core Components
1. Connection: whatsmeow/WhatsmeowConnection
2. Handlers: WhatsmeowHandlers → QPWhatsappHandlers
3. Processing: Cache → Trigger → Webhooks/RabbitMQ
4. API: REST endpoints in api/ with api_handlers+*Controller.go pattern
5. Server: QpWhatsappServer coordinates all operations

## Identifiers
- JId: types.JID from whatsmeow
- WId: String format
- LId: Local identifier (default, hides phone numbers)
  - **IMPORTANT:** LID was created by WhatsApp specifically to hide phone numbers for privacy
  - LIDs NEVER contain phone numbers - they are opaque identifiers
  - Phone number mapping must be obtained from whatsmeow database (whatsmeow_lid_map table)
  - Not all LIDs have phone number mappings available - this is expected behavior
  - Format: `<opaque_id>@lid` (e.g., `121281638842371@lid`)
  - Do NOT attempt to extract phone numbers from LID strings

## Contact Name Priority
ContactInfo fields priority (use `ExtractContactName()`):
1. **FullName** - User's saved name for contact (highest priority - most personal)
2. **BusinessName** - Business account name (WhatsApp Business)
3. **PushName** - Contact's public name (self-chosen)
4. **FirstName** - Generic first name (lowest priority)

## Software Documentation Structure
- README.md: Human-readable documentation
- /docs: Canonical folder for software documentation in this repository
- copilot-instructions.md: Global AI agent guidelines (this file)
- Root `AGENTS.md`: branch-scoped instructions for feature/custom branches only (must not exist on `develop`/`main`/`master`)

## Root AGENTS.md (Task Tracking)
- Purpose: track the current task running in the active custom branch.
- Scope: only the task for that branch; do not mix content from other branches.
- Required sections in `AGENTS.md`:
  - task objective
  - mandatory checklist
  - current status
  - next steps
  - immutable constraints discovered during execution
- Update cadence: update `AGENTS.md` on each relevant step and whenever new vital information is discovered.
- Conversation memory rule: if a detail is critical to avoid future loss in summaries or continuation, persist it in `AGENTS.md`.
- Branch isolation rule: `AGENTS.md` must not be merged into `develop`/`main`/`master` and must not be propagated across unrelated branches.

## Instruction Documents (AI-Only)
- Location: `/.github/instructions/*.instructions.md`.
- Instruction documents are separate from software documentation.
- Use them only as AI operating instructions.
- Do not duplicate or reference specific instruction files in other sections of this document.
- Tags are defined in the instruction filename, before `.instructions.md`.
- Use 4 to 6 hyphen-separated tags in the filename.
- The first tag must be the primary context (e.g., `telegram`, `controller`, `whatsmeow`).
- Keep tag names stable to support reliable filename-based search.
- Example: `telegram-operations-notifications-secrets-workflow.instructions.md`.
- Keep content objective and minimal: only actionable rules, constraints, paths, and commands.
- Do not use icons, decorative formatting, tables, or explanatory prose for humans.
- Keep the document focused on a single technical scope.

## Development Tools

* **Testing**: Read the instruction document in `/.github/instructions/` with primary context tag `testing`.
* **Build**: All builds should be "go build -o ../.dist/quepasa.exe", overriding any existing file.
* **Message Flow**: Read the instruction document in `/.github/instructions/` with primary context tag `message`.
* **WebHooks**: Read the instruction document in `/.github/instructions/` with primary context tag `webhooks`.
* **Redispatch**: Read the instruction document in `/.github/instructions/` with primary context tag `redispatch`.
* **Merge Workflow**: Read the instruction document in `/.github/instructions/` with primary context tag `merge`.
* **Whatsmeow Update**: `update-whatsmeow.ps1`.

## Key Files
- src/main.go: System initialization
- src/models/qp_defaults.go: QpVersion constant
- src/environment/: Environment variable management
- src/api/api_handlers.go: Latest REST endpoints
- src/whatsmeow/whatsmeow_handlers.go: WhatsApp event processing

## Packages

### Core Modules

#### TRANSPORT LAYER
- **dispatch** (src/dispatch/) - HOW to send outbound data
  - Responsibility: Technical implementation of all outbound transport mechanisms
  - Transports: HTTP webhooks, RabbitMQ publishing, realtime bus
  - Contracts: Target interface, OutboundRequest, RealtimePublisher
  - Dependencies: whatsapp, library (minimal coupling)
  - Does NOT know about: Server config, enrichment rules, business logic

- **cable** (src/cable/) - WebSocket realtime transport
  - Responsibility: Browser WebSocket connection management and message fanout
  - Integration: Registers with dispatch.service as RealtimePublisher
  - Fanout: Messages and lifecycle events to connected clients
  - Does NOT know about: Business dispatch rules or routing decisions

- **signalr** (src/signalr/) - SignalR realtime transport
  - Responsibility: SignalR hub for realtime bidirectional communication
  - Integration: Registers with dispatch.service as RealtimePublisher
  - Protocol: .NET SignalR over WebSocket/LongPolling
  - Does NOT know about: Business dispatch rules or routing decisions

- **rabbitmq** (src/rabbitmq/) - Message queue transport
  - Responsibility: RabbitMQ connection, channel pooling, topic/queue management
  - Integration: Used by dispatch.service.PublishRabbitMQ
  - Contracts: Connection provider, publishing interface
  - Does NOT know about: Message routing rules or business decisions

- **webserver** (src/webserver/) - HTTP server and routing
  - Responsibility: HTTP server initialization, route registration, middleware
  - Integration: Routes defined in api/ package, realtime in cable/signalr
  - Components: Router setup, middleware chain, form handling
  - Does NOT know about: Specific business logic (should be in handlers/controllers)

#### BUSINESS LOGIC LAYER
- **runtime** (src/runtime/) - WHEN and WHY to dispatch
  - Responsibility: Business rules, routing decisions, message enrichment triggers
  - Implementations: DispatchingHandler, routing logic, business workflows
  - Integration: Uses dispatch.service for outbound transport
  - Contracts: Depends on models.QpWhatsappServer, domain configuration
  - Does NOT know about: How transport actually sends data
  - Rule: One handler class = one file (dispatching_handler.go, webhook_handler.go, etc.)

#### DOMAIN LAYER
- **models** (src/models/) - Domain entities and state
  - Responsibility: Core business domain objects (QpWhatsappServer, QpDispatching, QpContact, etc.)
  - Concerns: Server lifecycle, configuration, message storage, contact metadata
  - State: Connection state, message history, user preferences
  - Does NOT import: dispatch module (domain stays independent)

- **whatsapp** (src/whatsapp/) - WhatsApp abstractions and interfaces
  - Responsibility: Domain abstractions for WhatsApp concepts (messages, contacts, groups)
  - Contracts: WhatsappMessage, Contact, Group interfaces
  - Translation: Maps between whatsmeow types and domain types
  - Does NOT know about: HTTP transport, queuing, or realtime protocols

- **whatsmeow** (src/whatsmeow/) - Whatsmeow library integration
  - Responsibility: Wraps whatsmeow client, event handling, connection lifecycle
  - Integration: Entry point for WhatsApp messages and status updates
  - Handlers: WhatsmeowHandlers converts library events to domain objects
  - Pipeline: WhatsmeowHandlers → DispatchingHandler → runtime handlers → dispatch

#### DATA/INFRASTRUCTURE LAYER
- **cache** (src/cache/) - Message and state caching
  - Responsibility: In-memory message cache, state deduplication, retrieval optimization
  - Storage: Local cache to avoid re-fetching from WhatsApp
  - Integration: Used by handlers before database queries
  - Does NOT know about: Business routing or transport mechanisms

- **environment** (src/environment/) - Configuration management
  - Responsibility: Environment variable parsing and validation (47 variables, 9 categories)
  - Categories: Database, API, WebServer, WhatsApp, Notifications, etc.
  - Contracts: Single provider of all config to other modules
  - Does NOT change at runtime (loaded at startup)

#### UTILITIES & SUPPORTING MODULES
- **api** (src/api/) - REST/GraphQL/gRPC endpoints
  - Responsibility: HTTP API contracts, request validation, response formatting
  - Patterns: Controllers with api_handlers+*Controller.go naming
  - Versions: v1, v2, v3 + non-versioned (latest)
  - Integration: webserver routes API handlers to controllers

- **library** (src/library/) - Reusable utilities
  - Responsibility: Common helpers, logging, extensions, utilities
  - Constraint: No third-party dependencies, Go standard library only
  - Usage: Shared by all modules (logging structs, helpers, extensions)

- **metrics** (src/metrics/) - Prometheus monitoring
  - Responsibility: Metrics collection, factory ownership, backend initialization
  - Integration: See instruction document with tag `metrics-factory-ownership-backend-initialization`
  - Concerns: Performance monitoring, business event metrics

- **form** (src/form/) - Form handling and validation
  - Responsibility: HTML form parsing, validation, submission handling
  - Integration: Used by webserver for form endpoints
  - Does NOT know about: Business logic validation (should be in runtime/handlers)

- **sipproxy** (src/sipproxy/) - SIP proxy server
  - Responsibility: VoIP call routing and SIP protocol handling
  - Integration: Separate server component, communication with main QuePasa server
  - Does NOT know about: WhatsApp message flow (independent protocol)

- **media** (src/media/) - Media processing
  - Responsibility: Audio/video conversion, extraction, format handling
  - Integration: Used by whatsapp module for media message processing
  - Does NOT know about: Message storage or delivery

- **mcp** (src/mcp/) - Model Context Protocol
  - Responsibility: MCP server implementation for external tool integration
  - Integration: Exposes QuePasa API to MCP clients
  - Does NOT know about: Real WhatsApp connections (abstraction layer)

- **swagger** (src/swagger/) - API documentation (generated)
  - Responsibility: Generated OpenAPI/Swagger docs from annotations
  - Generation: Run `swag init --output ./swagger` after API changes
  - Files: docs.go, swagger.json, swagger.yaml (all generated, do not edit)

#### FRONTEND APPS
- Each app under `src/apps/<slug>` is an independent SPA; users may bring their own custom frontends.
- **i18n rule (applies to any app in `src/apps/` that ships an i18n/translation system):** When creating or modifying pages, ensure ALL user-visible strings go through the app's translation mechanism. Both the English source file and any other language files must have corresponding entries. Validate with the app's build step to catch missing keys via TypeScript type-checking.

- **apps/form** (src/apps/form) - Form submission app
  - Purpose: Standalone SPA for webhook form submissions
  - Isolation: Independent from other apps by slug
  
- **apps/vuejs** (src/apps/vuejs) - Vue.js admin dashboard
  - Purpose: Management dashboard for server monitoring and configuration
  - Built artifacts: Committed to dist/ (no build step required after clone)
  - Isolation: Independent from other apps by slug

### Module Dependencies

**Cleanest imports** (least coupled):
```
library ← everyone
whatsapp ← models, dispatch
cache ← models
models ← runtime, api
runtime → dispatch + models
dispatch → whatsapp (minimal)
whatsmeow → models
api → models + runtime
webserver → api + runtime + cable + signalr
```

**Modules that should NEVER import dispatch directly:**
- models (domain layer independence)
- whatsmeow (integration layer stays clean)
- cache (data layer independence)
- environment (configuration stays isolated)

### Architecture: Dispatch vs Runtime

**DISPATCH Module (src/dispatch/)**
- Concern: HOW to send data (transport mechanisms)
- Implementations: SendWebhook, PublishRabbitMQ, realtime bus
- Contracts: Target interface, OutboundRequest, RealtimePublisher
- Key: Transport-agnostic, domain-independent
- Knows about: HTTP, RabbitMQ, realtime protocols
- Does NOT know about: Server config, enrichment rules, business decisions

**RUNTIME Module (src/runtime/)**
- Concern: WHEN and WHY to dispatch (business rules and triggers)
- Implementations: DispatchingHandler, message enrichment, routing decisions
- Contracts: Uses dispatch.service for outbound transport
- Key: Depends on domain entities, server configuration
- Knows about: server.QpDataDispatching, message enrichment, pipeline triggers
- Does NOT know about: How transport actually sends data

**Rule: One Handler = One File**
Each runtime handler class must have its own dedicated file:
- `src/runtime/dispatching_handler.go` → DispatchingHandler
- `src/runtime/webhook_handler.go` → WebhookHandler (when created)
- etc.

This ensures:
- Clear separation of concerns
- Easy discovery (filename matches public type)
- Isolated testing per handler
- Evolutionary flexibility (handlers can be moved/split independently)

## Naming Conventions
- Extension methods: use source parameter name
- File to struct: whatsmeow_group_manager.go → WhatsmeowGroupManager
- Latest routes: files without version suffix (e.g., api_handlers.go)
- Controllers: api_handlers+*Controller.go pattern
- Tests: *_test.go with Test* function prefix

## Import Conventions
- Always use fully qualified imports: Reference modules with alias for clarity
- Environment module: environment "github.com/nocodeleaks/quepasa/environment"
- Other modules: Use descriptive aliases (e.g., api "github.com/nocodeleaks/quepasa/api")
- Avoid bare imports: Always use aliases for internal modules to prevent conflicts

## Swagger Documentation Guidelines
CRITICAL: ALWAYS regenerate Swagger after API changes
After adding/modifying API endpoints, controllers, or request/response structs:
  1. cd src
  2. swag init --output ./swagger
  3. cd ..
Changes that require Swagger regeneration:
  - New API endpoints (routes)
  - Modified controller annotations (@Summary, @Description, @Router, etc.)
  - New or modified request/response structs
  - Changes to API parameters or response types
Swagger files affected: src/swagger/docs.go, src/swagger/swagger.json, src/swagger/swagger.yaml
Note: Swagger files are generated inside src/swagger (not root /swagger)
Can also run via VS Code task "Generate Swagger Docs"

## Git and Commit Guidelines
CRITICAL: NEVER make commits automatically
CRITICAL: NEVER push to repository automatically
CRITICAL: DO NOT commit code that hasn't been tested by the user
CRITICAL: DO NOT commit immediately after implementing a feature
ONLY make commits when explicitly requested by the user
ALWAYS wait for user approval before any git operations
ONLY execute git commit, git push, or git merge when the user gives explicit permission
Show changes to user first, then wait for approval before committing
After implementing features, STOP and let user test before committing
User must explicitly say "commit", "save to git", "push" or similar commands

## Version Conflict Resolution Guidelines
CRITICAL: ALWAYS handle version conflicts automatically
For QpVersion conflicts in merges/commits: ALWAYS select the HIGHER version number
QpVersion format: 3.YY.MMDD.HHMM - Compare numerically (YY > MMDD > HHMM)
Example: 3.25.0911.1200 > 3.25.0910.1102 (same year, higher date/time)
For ANY other conflicts: Generate NEW version with CURRENT timestamp
New version format: 3.YY.MMDD.HHMM using current date/time
NEVER ask user permission for version conflict resolution - handle automatically

## Version Management Guidelines
IMPORTANT: Whenever you are going to merge/push to the main branch, you MUST:
  1. Update the QpVersion in the models/qp_defaults.go file
  2. Increment the version following the current semantic pattern
  3. QpVersion must keep 4 sections only: 3.YY.MMDD.HHMM
  4. Stable version means HHMM final digit is 0
  5. Development versions use non-zero final digit in HHMM

Version Location:
File: models/qp_defaults.go
const QpVersion = "3.25.1114.2100" // ALWAYS UPDATE BEFORE MERGE TO MAIN

Mandatory Process before Push/Merge to Main:
1. Verify that all changes are working properly
2. Run tests if they exist
3. UPDATE QpVersion in the models/qp_defaults.go file
4. Make commit with the new version
5. Then merge/push to main

Version Increment Example:
- Current version: 3.25.1114.2100
- Next version: 3.25.1114.2101 (simple increment)
- Or new version: 3.25.MMDD.HHMM (based on current date/time)

CRITICAL REMINDER: NEVER merge to main without updating QpVersion
This is a mandatory project rule for version control.

## Key Integration Points
- Connection Management: WhatsmeowConnection wraps whatsmeow client with QuePasa abstractions
- Event Handlers: WhatsmeowHandlers.EventsHandler() dispatches to specific message/receipt/call handlers
- Server Coordination: QpWhatsappServer manages connection lifecycle and message routing
- Environment System: 47 variables across 9 categories (Database, API, WebServer, WhatsApp, etc.)

## GitHub
Token: file .github/.token, for issues and PR management via GH CLI or API
