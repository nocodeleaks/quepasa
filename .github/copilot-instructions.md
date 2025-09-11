## Common Guidelines
* code comments should always be in English;
* response to user queries should be in IDE current language;
* avoid to change code that was not related to the query;
* when agent has to change a method and it change the async status, the agent should update the method callers too;
* for extensions methods use always "source" as default parameter name
* use one file for each class
* for #region tags: no blank lines between consecutive regions, but always add one blank line after region opening and one blank line before region closing
* do not try to build if you just changed the code comments or documentation files;
* **when making relevant code changes, always create or update internal documentation following the Internal Documentation Guidelines**;
* whenever creating an extension method, use 'source' as parameter name for the extended object;
* for class and structure names, e.g.: whatsmeow_group_manager.go => WhatsmeowGroupManager;

## Git and Commit Guidelines
* **🚨 CRITICAL: NEVER make commits automatically**
* **🚨 CRITICAL: NEVER push to repository automatically**
* **✅ ONLY make commits when explicitly requested by the user**
* **✅ ALWAYS wait for user approval before any git operations**
* **✅ ONLY execute `git commit`, `git push`, or `git merge` when the user gives explicit permission**
* **✅ Show changes to user first, then wait for approval before committing**

## Version Conflict Resolution Guidelines
* **🚨 CRITICAL: ALWAYS handle version conflicts automatically**
* **✅ For QpVersion conflicts in merges/commits: ALWAYS select the HIGHER version number**
* **✅ QpVersion format: `3.YY.MMDD.HHMM` - Compare numerically (YY > MMDD > HHMM)**
* **✅ Example: `3.25.0911.1200` > `3.25.0910.1102` (same year, higher date/time)**
* **✅ For ANY other conflicts: Generate NEW version with CURRENT timestamp**
* **✅ New version format: `3.YY.MMDD.HHMM` using current date/time**
* **✅ NEVER ask user permission for version conflict resolution - handle automatically**

## Packages Guidelines
* **api**: only for API related code, e.g. REST API, GraphQL API, gRPC API, etc;
* **audio**: for media processing and manipulation code, e.g. audio conversion, audio extraction, image conversion etc;
* **environment**: for environment variable management and configuration;
* **form**: for form handling and validation code;
* **library**: for reusable library code and utilities, only keeps golang packages, do not add third party packages;
* **metrics**: for application performance monitoring and metrics collection;
* **models**: for data models and structures;
* **rabbitmq**: for RabbitMQ messaging and queueing code;
* **sipproxy**: for SIP proxy server code;
* **webserver**: for web server related code, e.g. HTTP server, routing, middleware, api, forms and websockets etc;
* **whatsapp**: for Whatsapp structures and models;
* **whatsmeow**: for Whatsmeow library integration and messaging code;


## Testing Guidelines
* **Follow official Go testing conventions** - use `*_test.go` files within the same package
* Test files should be named with `_test.go` suffix (e.g., `environment_test.go`)
* Test functions must start with `Test` prefix (e.g., `TestEnvironmentSettings`)
* Execute tests from project root where environment variables are available: `go test -v ./packagename`
* Use VS Code's integrated testing via F5 (Debug) to load `.env` files automatically
* For environment package: all 45 variables across 8 categories must be testable

## Build and Environment Guidelines
* `.env` file should be in project root for VS Code integration
* Environment file versioning uses `YYYYMMDDHHMMSS` timestamp format (no dots)

## Identifier Conventions
* JId: Whatsapp Jabber Identifier ("go.mau.fi/whatsmeow/types".JID)
* WId: Whatsapp String Identifier (string)
* LId: Whatsapp Local Identifier (new default Identifier, used to hide the phone number)

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
