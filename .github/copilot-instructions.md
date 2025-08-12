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
