# PLAN: MCP (Model Context Protocol) Improvements

**Created:** 2025-01-15
**Status:** Proposed
**Priority:** High
**Target Module:** `/src/mcp/`

## Context

The QuePasa project has a Model Context Protocol (MCP) server implementation that enables AI agents to interact with WhatsApp API endpoints through a standardized JSON-RPC interface. The implementation uses:

- Swagger annotation auto-discovery (`ParseSwaggerAnnotations`)
- Manual handler registry (`HandlerRegistry` map)
- JSON-RPC 2.0 protocol support
- SSE (Server-Sent Events) for streaming
- Tool execution with authentication context (`MCPToolContext`)

While the architecture is solid and functional, there are maintenance bottlenecks, robustness gaps, and missing features that limit long-term usability.

## Problems Identified

### 1. Manual Handler Registry (Blocking)
- `mcp_handler_registry.go` contains a hardcoded `map[string]http.HandlerFunc`
- Every new/renamed endpoint requires manual registry update
- No validation between Swagger discovery and registry entries
- Violates auto-discovery principle already in use

### 2. Tool Naming Limitations (Required)
- Tool names auto-generated from method + path (e.g., `get_v1_5_0_contacts`)
- No support for custom friendly names
- Names can be long and non-intuitive for end users
- No way to override via Swagger annotations

### 3. Incomplete SSE Implementation (Required)
- Basic SSE endpoint exists (`HandleSSE`)
- Missing standard MCP notification types:
  - `notifications/progress` for long-running tools
  - `notifications/message` for async updates
  - `notifications/` logging support
- No test coverage for SSE behavior

### 4. Missing Test Coverage (Blocking)
- No `*_test.go` files in `/src/mcp/`
- Critical paths untested:
  - `GenerateInputSchema` schema generation
  - `ParseSwaggerAnnotations` parsing logic
  - `ExecuteWithContext` context injection
  - Error handling and JSON-RPC error codes
- Regression risk on changes

### 5. Fragile Swagger Parser (Required)
- Depends on string parsing and regex
- Silent failures on malformed annotations
- No line-by-line error reporting for debugging
- No validation that endpoints exist in Chi router

### 6. Non-Standard Error Handling (Required)
- HTTP errors not mapped to JSON-RPC error codes
- Missing standard error codes:
  - `-32602` (Invalid params)
  - `-32603` (Internal error)
  - Custom error domains for QuePasa-specific failures
- Inconsistent error formats across tools

### 7. Performance Considerations (Conditional)
- Uses `httptest.NewRecorder` for every tool execution
- Potential memory allocation overhead under high load
- No response recorder pooling
- Debug logs may leak sensitive headers

### 8. Schema Validation Gaps (Conditional)
- `InputSchema` generated but not validated
- No check that required Swagger params match generated schema
- No explicit validation against handler signatures
- Risk of mismatched expectations

---

## Detailed Implementation Plan

### Phase 1: Eliminate Manual Handler Registry (Blocking)

**Objective:** Remove maintenance bottleneck by automating handler discovery.

**Tasks:**

1. [ ] **Implement Handler Name Validation**
   - Add `ValidateHandlerRegistry()` function in `mcp_api_registry.go`
   - Called after `ParseSwaggerAnnotations()` completes
   - Warn on Swagger entries missing from registry
   - Provide clear error message with file:line context

2. [ ] **Add Runtime Handler Lookup**
   - Import `api` package and use reflection to discover handlers
   - Or implement active registration pattern where controllers register themselves
   - Fallback to existing registry for compatibility

3. [ ] **Document Registration Convention**
   - Add comment to `mcp_handler_registry.go` explaining registration rules
   - Create `docs/USAGE-mcp-handler-registration.md` with guidelines
   - Example: "Controller functions must be exported and match exact name in Swagger `@Summary`"

**Files to Modify:**
- `src/mcp/mcp_api_registry.go`
- `src/mcp/mcp_handler_registry.go`
- `docs/USAGE-mcp-handler-registration.md` (new)

**Validation:**
- Run `ParseSwaggerAnnotations()` with intentionally missing handler
- Verify warning appears in logs with clear guidance
- No runtime crashes when registry mismatch occurs

---

### Phase 2: Add Custom Tool Naming (Required)

**Objective:** Allow human-friendly tool names via Swagger annotations.

**Tasks:**

1. [ ] **Extend Swagger Parser**
   - Add `@MCPName` annotation support in `mcp_swagger_parser.go`
   - Store custom name in `SwaggerEndpoint.MCPName` field
   - Fallback to auto-generated name if annotation missing

2. [ ] **Update Tool Name Generation**
   - Modify `GenerateToolName()` in `mcp_swagger_parser.go`
   - Priority: `@MCPName` > `@Summary` (lowercased) > auto-generated
   - Validate name uniqueness across all tools

3. [ ] **Update Handler Registry Doc**
   - Document `@MCPName` annotation in Swagger comments
   - Provide examples for different tool naming strategies

**Example Usage:**
```go
// @Summary Send a text message to a WhatsApp contact
// @Description Sends a message to the specified phone number with optional media
// @MCPName send_message
// @Param token header string true "Bot authentication token"
// @Param body body SendMessageRequest true "Message payload"
// @Router /v1.5.0/send [post]
func SendMessageController(w http.ResponseWriter, r *http.Request) { ... }
```

**Files to Modify:**
- `src/mcp/mcp_swagger_parser.go`
- `docs/USAGE-mcp-swagger-annotations.md` (new)

**Validation:**
- Add `@MCPName` to a controller
- Verify tool name in MCP `tools/list` response
- Ensure fallback still works for endpoints without annotation

---

### Phase 3: Complete SSE Implementation (Required)

**Objective:** Support standard MCP notifications for streaming and long-running tools.

**Tasks:**

1. [ ] **Define Notification Types**
   - Create `SSENotification` struct in `mcp_server.go`
   - Implement `notifications/progress` (for progress updates)
   - Implement `notifications/message` (for async responses)
   - Implement `notifications/` logging (optional)

2. [ ] **Update HandleSSE Method**
   - Accept notification channel parameter
   - Stream notifications as they arrive
   - Handle client disconnect gracefully

3. [ ] **Add Progress Reporting to Long-Running Tools**
   - Identify tools that benefit from progress (e.g., bulk message send)
   - Modify `ExecuteWithContext` to accept progress callback
   - Example: `SendBulkMessagesTool` reports progress per message

4. [ ] **Add SSE Tests**
   - Create `mcp_server_test.go`
   - Test SSE connection establishment
   - Test notification streaming
   - Test client disconnect handling

**Files to Modify:**
- `src/mcp/mcp_server.go`
- `src/mcp/mcp_tools.go` (extend `MCPTool` interface if needed)
- `src/mcp/mcp_server_test.go` (new)

**Validation:**
- Connect via SSE to `/mcp` endpoint
- Trigger a long-running tool
- Verify progress notifications stream in real-time
- Test disconnection does not cause panic

---

### Phase 4: Add Comprehensive Test Coverage (Blocking)

**Objective:** Ensure robustness and enable confident refactoring.

**Tasks:**

1. [ ] **Create mcp_swagger_parser_test.go**
   - Test `ParseSwaggerAnnotations()` with valid annotations
   - Test parsing with malformed/mixed annotations
   - Test parameter extraction (`@Param`, `@MCPName`, etc.)
   - Test edge cases: empty files, missing annotations, Unicode

2. [ ] **Create mcp_api_tool_test.go**
   - Test `GenerateInputSchema()` for all Go types (string, int, bool, object, array)
   - Test `ExecuteWithContext()` with valid params
   - Test error handling for invalid params
   - Test authentication context injection (master vs bot token)

3. [ ] **Create mcp_server_test.go**
   - Test JSON-RPC request/response format
   - Test `tools/list` endpoint
   - Test `tools/call` execution
   - Test error mapping (HTTP → JSON-RPC codes)
   - Test SSE connection and streaming

4. [ ] **Create mcp_registry_test.go**
   - Test tool registration and retrieval
   - Test duplicate name handling
   - Test handler validation

5. [ ] **Add Test Coverage Badge**
   - Run tests with coverage: `go test -cover ./src/mcp/...`
   - Target: Minimum 80% coverage

**Files to Create:**
- `src/mcp/mcp_swagger_parser_test.go`
- `src/mcp/mcp_api_tool_test.go`
- `src/mcp/mcp_server_test.go`
- `src/mcp/mcp_registry_test.go`

**Validation:**
- Run `go test ./src/mcp/... -v`
- Verify all tests pass
- Check coverage report (`go test -coverprofile=coverage.out && go tool cover -html=coverage.out`)

---

### Phase 5: Robust Swagger Parser (Required)

**Objective:** Improve parser reliability and debugging experience.

**Tasks:**

1. [ ] **Add Line-by-Line Error Context**
   - Store file position in `SwaggerEndpoint`
   - Include file path, line number in error messages
   - Format: `Error in src/api/contacts.go:45: missing @Param for token`

2. [ ] **Add Annotation Validation**
   - Validate required fields present (`@Summary`, `@Router`)
   - Validate `@Router` format: `METHOD path`
   - Validate parameter names match handler signature
   - Return structured errors with fix suggestions

3. [ ] **Add Router Existence Check**
   - After parsing, query Chi router for each endpoint path+method
   - Warn on Swagger endpoints not found in router
   - Help detect annotation typos or missing route registration

4. [ ] **Add Parser Statistics**
   - Log counts: files scanned, endpoints found, skipped, errors
   - Example: `MCP: Scanned 15 files, found 42 endpoints, 3 errors, 2 skipped (@MCPHidden)`

**Files to Modify:**
- `src/mcp/mcp_swagger_parser.go`
- `src/mcp/mcp_api_registry.go` (add validation after parsing)

**Validation:**
- Introduce a malformed annotation in a controller
- Verify error message includes file:line
- Verify parser continues despite individual errors
- Check router validation catches missing routes

---

### Phase 6: Standardize Error Handling (Required)

**Objective:** Map HTTP errors to JSON-RPC 2.0 error codes consistently.

**Tasks:**

1. [ ] **Define JSON-RPC Error Constants**
   - Add constants in `mcp_server.go`:
     ```go
     const (
         JSONRPCInvalidRequest = -32600
         JSONRPCMethodNotFound = -32601
         JSONRPCInvalidParams  = -32602
         JSONRPCInternalError  = -32603
         QuePasaAuthError      = -32001
         QuePasaRateLimitError = -32002
     )
     ```

2. [ ] **Create Error Mapping Function**
   - Add `MapHTTPToJSONRPCError(statusCode int, err error) JSONRPCError`
   - Map 400 → -32602, 401 → -32001, 500 → -32603, etc.
   - Include original HTTP error in `data` field

3. [ ] **Update Tool Execution Error Handling**
   - In `mcp_api_tool.go`, wrap handler errors with `MapHTTPToJSONRPCError`
   - Ensure response format matches JSON-RPC spec:
     ```json
     {
       "jsonrpc": "2.0",
       "error": { "code": -32602, "message": "Invalid params", "data": {...} },
       "id": 1
     }
     ```

4. [ ] **Add Error Tests**
   - Test error mapping for all HTTP status codes
   - Test error format matches JSON-RPC spec
   - Test custom error codes (QuePasa-specific)

**Files to Modify:**
- `src/mcp/mcp_server.go` (add constants)
- `src/mcp/mcp_api_tool.go` (update error handling)
- `src/mcp/mcp_api_tool_test.go` (add error tests)

**Validation:**
- Trigger tool with invalid params → verify `-32602` response
- Trigger tool without auth → verify `-32001` response
- Test unknown method → verify `-32601` response

---

### Phase 7: Performance Optimization (Conditional)

**Objective:** Reduce overhead under high load (only if profiling shows need).

**Tasks:**

1. [ ] **Profile Current Implementation**
   - Run load test: `ab -n 1000 -c 10 -p body.json http://localhost/mcp`
   - Use `pprof` to identify bottlenecks
   - Only proceed if `httptest.NewRecorder` is a hotspot

2. [ ] **Implement Response Recorder Pool**
   - Create `recorderPool := sync.Pool{ New: func() interface{} { return httptest.NewRecorder() } }`
   - Use `recorderPool.Get()` before execution
   - Call `recorderPool.Put(rec)` after execution

3. [ ] **Sanitize Debug Logs**
   - Filter sensitive headers (`Authorization`, `X-Api-Key`) from logs
   - Use `log.Debugf("[MCP] %s %s (auth omitted)", method, path)`

4. [ ] **Add Performance Benchmarks**
   - Create `mcp_benchmark_test.go`
   - Benchmark tool execution with/without pool
   - Benchmark Swagger parsing

**Files to Modify:**
- `src/mcp/mcp_api_tool.go` (add pool)
- `src/mcp/mcp_benchmark_test.go` (new)

**Validation:**
- Run benchmarks: `go test -bench=. -benchmem ./src/mcp/...`
- Compare allocations before/after pooling
- Verify logs do not expose sensitive data

---

### Phase 8: Schema Validation (Conditional)

**Objective:** Ensure generated schemas match actual handler expectations.

**Tasks:**

1. [ ] **Add Schema-to-Signature Validation**
   - Use reflection to inspect handler signature
   - Compare generated schema with actual parameter types
   - Warn on mismatches (e.g., schema says `int`, handler expects `string`)

2. [ ] **Validate Required Fields**
   - Check that Swagger `@Param(required=true)` results in schema `"required": [...]`
   - Verify required params match handler validation logic

3. [ ] **Add Schema Tests**
   - Test schema generation for all parameter types
   - Test required/optional field correctness
   - Test complex nested objects and arrays

**Files to Modify:**
- `src/mcp/mcp_api_tool.go` (add validation)
- `src/mcp/mcp_api_tool_test.go` (add schema tests)

**Validation:**
- Add handler with intentional schema mismatch
- Verify warning appears in logs
- Fix mismatch and verify warning disappears

---

## Implementation Checklist

- [ ] Phase 1: Eliminate Manual Handler Registry
  - [ ] Add validation function
  - [ ] Implement runtime handler lookup
  - [ ] Document registration convention
  - [ ] Test validation with missing handlers

- [ ] Phase 2: Add Custom Tool Naming
  - [ ] Extend parser with `@MCPName`
  - [ ] Update `GenerateToolName` logic
  - [ ] Document annotation usage
  - [ ] Test custom names and fallback

- [ ] Phase 3: Complete SSE Implementation
  - [ ] Define notification types
  - [ ] Update `HandleSSE` method
  - [ ] Add progress reporting to tools
  - [ ] Add SSE tests

- [ ] Phase 4: Add Comprehensive Test Coverage
  - [ ] Create `mcp_swagger_parser_test.go`
  - [ ] Create `mcp_api_tool_test.go`
  - [ ] Create `mcp_server_test.go`
  - [ ] Create `mcp_registry_test.go`
  - [ ] Achieve 80%+ coverage

- [ ] Phase 5: Robust Swagger Parser
  - [ ] Add line-by-line error context
  - [ ] Add annotation validation
  - [ ] Add router existence check
  - [ ] Add parser statistics logging

- [ ] Phase 6: Standardize Error Handling
  - [ ] Define JSON-RPC error constants
  - [ ] Create error mapping function
  - [ ] Update tool execution error handling
  - [ ] Add error tests

- [ ] Phase 7: Performance Optimization (Conditional)
  - [ ] Profile current implementation
  - [ ] Implement response recorder pool
  - [ ] Sanitize debug logs
  - [ ] Add performance benchmarks

- [ ] Phase 8: Schema Validation (Conditional)
  - [ ] Add schema-to-signature validation
  - [ ] Validate required fields
  - [ ] Add schema tests

---

## Validation Criteria

After implementing all phases, the MCP module should:

1. **Zero Manual Maintenance Required**
   - New controllers appear as MCP tools automatically
   - No `HandlerRegistry` updates needed for standard endpoints
   - Registry validation catches mismatches early

2. **Intuitive Tool Names**
   - Tools have human-friendly names via `@MCPName`
   - Fallback logic ensures backward compatibility
   - Name conflicts detected and reported

3. **Complete SSE Support**
   - Long-running tools report progress via `notifications/progress`
   - Async updates supported via `notifications/message`
   - SSE connections handle disconnects gracefully

4. **Test Coverage >80%**
   - All critical paths tested
   - Edge cases covered
   - No regressions on changes

5. **Robust Swagger Parsing**
   - Malformed annotations report file:line context
   - Parser continues despite individual errors
   - Router validation catches missing routes

6. **Standard Error Format**
   - All errors follow JSON-RPC 2.0 spec
   - HTTP status codes mapped to correct JSON-RPC codes
   - Custom error codes for QuePasa-specific failures

7. **Performance (If Needed)**
   - Response recorder pooling reduces allocations
   - Benchmarks show improvement under load
   - Logs do not expose sensitive data

8. **Schema Accuracy**
   - Generated schemas match handler signatures
   - Required fields validated correctly
   - Mismatches detected and warned

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing MCP tools | High | Ensure backward compatibility for tool names, add validation warnings before removing registry |
| Test coverage low → regressions | Medium | Prioritize Phase 4 early, write tests alongside implementation |
| SSE implementation complex | Medium | Start with basic streaming, add progress support incrementally |
| Reflection-based handler lookup slower | Low | Cache handler lookups after first discovery, measure impact before/after |
| Swagger parser changes may break existing annotations | Medium | Add deprecation warnings, support both old and new formats during transition |

---

## Dependencies

None. All phases are self-contained within `/src/mcp/`.

---

## Next Steps

1. Review and approve this plan
2. Prioritize phases (suggested: 1, 4, 2, 3, 5, 6, 7, 8)
3. Create feature branch: `feature/mcp-improvements`
4. Implement Phase 1 (eliminate manual registry)
5. Implement Phase 4 (test coverage) in parallel with other phases
6. Incrementally implement remaining phases
7. Run full test suite after each phase
8. Update Swagger annotations with `@MCPName` as needed
9. Merge to `develop` after all phases complete and validated

---

## References

- **MCP Specification:** https://spec.modelcontextprotocol.io/
- **JSON-RPC 2.0:** https://www.jsonrpc.org/specification
- **Swagger in Go (swaggo):** https://github.com/swaggo/swag
- **QuePasa MCP Implementation:** `/src/mcp/`
- **Existing MCP Tools:** `tool_health.go`, `tool_list_servers.go`

---

**Last Updated:** 2025-01-15
**Author:** AI Agent (based on code analysis)