# QuePasa AI Agent Instructions

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
- api: REST API, GraphQL, gRPC endpoints and controllers
- audio: Media processing (conversion, extraction)
- environment: Environment variables and configuration (9 categories, 47 vars)
- form: Form handling and validation
- library: Reusable utilities (Go packages only, no third-party)
- metrics: Prometheus monitoring and metrics
- models: Data structures and business logic
- rabbitmq: Message queueing and async processing
- sipproxy: SIP proxy server
- webserver: HTTP server, routing, middleware, forms, websockets (check instruction documents for module-specific rules)
- whatsapp: WhatsApp abstractions and interfaces
- whatsmeow: Whatsmeow library integration

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
