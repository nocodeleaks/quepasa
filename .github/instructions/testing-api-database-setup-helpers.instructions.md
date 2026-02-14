# Testing Instruction

## Scope
- Test setup for API handlers and database-backed tests.
- Primary files:
  - `src/api/testing_setup.go`
  - `src/api/api_handlers+*_test.go`

## Mandatory Setup
- Call `SetupTestService(t)` at test start.
- Call `defer CleanupTestDatabase(t)` in each test.
- Keep tests isolated and independent.

## Test Data Helpers
- `CreateTestUser(t, username, password)`
- `CreateTestServer(t, token, username)`
- `SetupTestMasterKey(t, masterKey)` and `defer cleanup()`

## Validation Rules
- Validate HTTP status code.
- Validate response payload structure.
- Cover success and failure cases.
- Avoid external dependencies when helper setup is available.

## Execution Commands
- `cd src/api`
- `go test -v`
- `go test -v -run TestName`
- `go test -v -cover`
