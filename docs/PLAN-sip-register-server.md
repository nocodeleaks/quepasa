# PLAN: SIP Register Server Implementation for QuePasa

**Date**: 2026-06-25
**Author**: AI Agent
**Status**: Planning
**Priority**: High

## Objective

Implement a SIP Register Server for QuePasa that allows users to register from any softphone directly to their WhatsApp instance using the QuePasa server, with authentication based on:
- **Username**: Instance identifier (token)
- **Password**: Optional (default: no password, token alone is secure)
- **Custom Password**: Can be set via API endpoints for enhanced security

## Current State

### Existing SIP Proxy
- QuePasa already has a functional SIP proxy for receiving calls from WhatsApp
- This proxy handles inbound calls from WhatsApp to SIP endpoints
- Current implementation is focused on proxy functionality, not registration

### Technical Context
- QuePasa is a Go-based project
- Already integrates with WhatsApp via whatsmeow library
- Uses SIP for VoIP communication (inbound calls)
- Current implementation: SIP proxy (stateless or stateful)

## Requirements

### Functional Requirements

1. **SIP Registration Support**
   - Implement RFC 3261 REGISTER method handling
   - Support multi-tenant registration (multiple WhatsApp instances)
   - Maintain registration state and expiration timers
   - Handle re-registration and de-registration

2. **Authentication**
   - User: Instance identifier (token) - unique per instance
   - Password: Optional (can be empty by default)
   - Use SIP Digest Authentication (RFC 3261, RFC 8760) when custom password is set
   - Default mode: Allow registration without password (token is already secure)
   - Custom password mode: Verify password hash from sip_credentials table
   - Verify instance exists in QuePasa database

3. **Password Management (New)**
   - Default: No password required (token alone is sufficient)
   - Optional: Set custom SIP password via API
   - Support password rotation via API
   - Passwords hashed with bcrypt

3. **Softphone Compatibility**
   - Support standard SIP softphones (Zoiper, MicroSIP, Linphone, X-Lite, etc.)
   - Compatible with SIP clients on Windows, macOS, Linux, iOS, Android
   - Standard SIP UDP/TCP transport (5060 or custom port)
   - Support for SRV records and DNS-based discovery

4. **Call Flow**
   - Registered softphone can receive inbound WhatsApp calls via SIP
   - Registration binds softphone to specific WhatsApp instance
   - Support multiple concurrent registrations per instance
   - Handle NAT traversal (STUN/TURN if needed)

### Non-Functional Requirements

1. **Security**
   - Prevent unauthorized registrations
   - Rate limiting to prevent registration flooding
   - TLS support for secure SIP (SIPS) if required
   - Log all registration attempts

2. **Performance**
   - Support multiple concurrent registrations
   - Minimal latency in registration processing
   - Efficient state management
   - Graceful handling of registration expiration

3. **Compatibility**
   - Must not break existing SIP proxy functionality
   - Backward compatible with current call flows
   - Follow SIP standards strictly

## Architecture

### Recommended Go SIP Libraries

Based on research:

1. **emiago/sipgo** (Recommended)
   - High-performance SIP stack for Go
   - Supports RFC 3261, RFC 3581, RFC 6026
   - Optimized for fast parsing and handling
   - Can implement stateful proxy and registration
   - Active development and community
   - Example implementations available

2. **Alternative: 1lann/go-sip**
   - Simpler implementation
   - Server package includes authentication
   - Less feature-rich but easier to integrate

**Recommendation**: Use `emiago/sipgo` for production-grade implementation with better performance and feature support.

### System Components

```
┌─────────────────┐
│   Softphone     │
│  (Zoiper, etc.) │
└────────┬────────┘
         │ SIP REGISTER
         │ User: <instance-token>
         │ Pass: (empty or custom)
         ▼
┌─────────────────────────────────┐
│  QuePasa SIP Register Server     │
│  (Go + sipgo)                    │
│  - Handle REGISTER method        │
│  - Verify instance token         │
│  - Check custom password (if set)│
│  - Registration state storage    │
│  - Expiration timers             │
└────────┬────────────────────────┘
         │ Verify token exists
         ▼
┌─────────────────────────────────┐
│  QuePasa Instance Database       │
│  - Instance (token) lookup       │
│  - Verification status           │
└────────┬────────────────────────┘
         │ Check custom password (optional)
         ▼
┌─────────────────────────────────┐
│  sip_credentials Table          │
│  - Token → Password hash mapping│
│  - Bcrypt hashed passwords      │
└─────────────────────────────────┘
         │ Update registration
         ▼
┌─────────────────────────────────┐
│  Registration State Storage      │
│  - In-memory or Redis            │
│  - Contact URIs per instance     │
│  - Expiration timestamps         │
└─────────────────────────────────┘
         │ For incoming calls
         ▼
┌─────────────────────────────────┐
│  Existing SIP Proxy              │
│  - Route calls to registered     │
│    softphones                    │
└─────────────────────────────────┘
```

### Data Flow

#### Registration Flow

#### Registration Flow (Default - No Password)

1. **REGISTER Request (No Authorization)**
   ```
   REGISTER sip:quepasa-server:5060 SIP/2.0
   Via: SIP/2.0/UDP 192.168.1.100:5060;branch=z9hG4bK123
   From: <sip:abc123-instance-id@quepasa-server:5060>;tag=abc123
   To: <sip:abc123-instance-id@quepasa-server:5060>
   Call-ID: call-id-123
   CSeq: 1 REGISTER
   Contact: <sip:192.168.1.100:5060>
   Expires: 3600
   Content-Length: 0
   ```

2. **200 OK (Success - No Password Required)**
   ```
   SIP/2.0 200 OK
   Via: SIP/2.0/UDP 192.168.1.100:5060;branch=z9hG4bK123;received=192.168.1.100
   From: <sip:abc123-instance-id@quepasa-server:5060>;tag=abc123
   To: <sip:abc123-instance-id@quepasa-server:5060>;tag=xyz789
   Call-ID: call-id-123
   CSeq: 1 REGISTER
   Contact: <sip:192.168.1.100:5060>;expires=3600
   Date: Thu, 01 Jan 2026 00:00:00 GMT
   Content-Length: 0
   ```

#### Registration Flow (Custom Password)

1. **REGISTER Request**
   ```
   REGISTER sip:quepasa-server:5060 SIP/2.0
   Via: SIP/2.0/UDP 192.168.1.100:5060;branch=z9hG4bK123
   From: <sip:abc123-instance-id@quepasa-server:5060>;tag=abc123
   To: <sip:abc123-instance-id@quepasa-server:5060>
   Call-ID: call-id-123
   CSeq: 1 REGISTER
   Contact: <sip:192.168.1.100:5060>
   Expires: 3600
   Content-Length: 0
   ```

2. **401 Unauthorized (Challenge)**
   ```
   SIP/2.0 401 Unauthorized
   Via: SIP/2.0/UDP 192.168.1.100:5060;branch=z9hG4bK123;received=192.168.1.100
   From: <sip:+5511999999999@quepasa-server:5060>;tag=abc123
   To: <sip:+5511999999999@quepasa-server:5060>;tag=xyz789
   Call-ID: call-id-123
   CSeq: 1 REGISTER
   WWW-Authenticate: Digest realm="quepasa", nonce="abc123", algorithm=MD5
   Content-Length: 0
   ```

3. **REGISTER with Auth**
   ```
   REGISTER sip:quepasa-server:5060 SIP/2.0
   Via: SIP/2.0/UDP 192.168.1.100:5060;branch=z9hG4bK456
   From: <sip:+5511999999999@quepasa-server:5060>;tag=abc123
   To: <sip:+5511999999999@quepasa-server:5060>
   Call-ID: call-id-123
   CSeq: 2 REGISTER
   Contact: <sip:192.168.1.100:5060>
   Expires: 3600
   Authorization: Digest username="+5511999999999", realm="quepasa",
     nonce="abc123", uri="sip:quepasa-server:5060", response="xyz456",
     algorithm=MD5, cnonce="abc", qop=auth, nc=00000001
   Content-Length: 0
   ```

4. **200 OK (Success)**
   ```
   SIP/2.0 200 OK
   Via: SIP/2.0/UDP 192.168.1.100:5060;branch=z9hG4bK456;received=192.168.1.100
   From: <sip:abc123-instance-id@quepasa-server:5060>;tag=abc123
   To: <sip:abc123-instance-id@quepasa-server:5060>;tag=xyz789
   Call-ID: call-id-123
   CSeq: 2 REGISTER
   Contact: <sip:192.168.1.100:5060>;expires=3600
   Date: Thu, 01 Jan 2026 00:00:00 GMT
   Content-Length: 0
   ```

#### Call Routing to Registered Softphone

1. WhatsApp call arrives → Existing SIP proxy
2. Proxy looks up registration state for target instance
3. Proxy forwards INVITE to registered Contact URI(s)
4. Softphone rings and answers
5. Call established

### Storage Strategy

#### Options:

1. **In-memory (Simple)**
   - Use Go maps or sync.Map
   - Simple and fast
   - No persistence across restarts
   - Suitable for development/small scale

2. **Redis (Recommended for Production)**
   - Persistent registration state
   - Supports expiration natively
   - Shared state across multiple server instances
   - Horizontal scaling possible
   - Fast lookups

3. **Database**
   - PostgreSQL or SQLite
   - Persistent storage
   - Slower than Redis
   - Overkill for registration state

**Recommendation**: Start with in-memory for development, migrate to Redis for production.

### Registration Data Model

```go
type Registration struct {
    Token         string    `json:"token"`          // Instance identifier (username)
    ContactURI    string    `json:"contact_uri"`    // Softphone contact URI
    ExpiresAt     time.Time `json:"expires_at"`     // Registration expiration
    RegisteredAt  time.Time `json:"registered_at"`  // Registration timestamp
    LastRefresh   time.Time `json:"last_refresh"`   // Last refresh timestamp
    CallID        string    `json:"call_id"`        // SIP Call-ID
    UserAgent     string    `json:"user_agent"`     // Softphone user-agent
    RemoteIP      string    `json:"remote_ip"`      // Client IP address
}
```

### SIP Credentials Data Model

```go
type QpSipCredentials struct {
    Token        string    `db:"token" json:"token"`
    PasswordHash string    `db:"password_hash" json:"-"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Database schema
CREATE TABLE sip_credentials (
    token TEXT PRIMARY KEY,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (token) REFERENCES servers(token) ON DELETE CASCADE
);
```

## Implementation Plan

### Phase 1: Research & Setup (Week 1)

**Task 1.1: Library Selection**
- [ ] Evaluate `emiago/sipgo` in detail
- [ ] Review documentation and examples
- [ ] Check compatibility with existing code
- [ ] Prototype basic REGISTER handler

**Task 1.2: Authentication Design**
- [ ] Define credential verification process
- [ ] Design digest authentication implementation
- [ ] Plan integration with existing instance database
- [ ] Document security considerations

**Task 1.3: Development Environment**
- [ ] Set up development branch
- [ ] Install required dependencies
- [ ] Create test environment with softphone
- [ ] Verify existing SIP proxy operation

### Phase 2: Core Registration Logic (Week 2)

**Task 2.1: REGISTER Method Handler**
- [ ] Implement basic REGISTER request parser
- [ ] Create response handler (200 OK, 401, 403, 404)
- [ ] Add Via header processing
- [ ] Handle Contact header parsing

**Task 2.2: Digest Authentication**
- [ ] Implement challenge generation (401)
- [ ] Implement response validation
- [ ] Add MD5 hash calculation
- [ ] Support nonce and qop

**Task 2.3: Credential Verification**
- [ ] Connect to instance database via existing QpServer model
- [ ] Verify token exists using FindByToken()
- [ ] Check instance is verified
- [ ] Check sip_credentials table for custom password
- [ ] If custom password exists, verify bcrypt hash
- [ ] If no custom password, allow without password
- [ ] Add error handling

**Task 2.4: SIP Credentials Database (New)**
- [ ] Create migration for sip_credentials table
- [ ] Implement QpDataSipCredentialsSql repository
- [ ] Add CRUD operations (Create, Read, Update, Delete)
- [ ] Implement password hashing with bcrypt
- [ ] Add foreign key cascade delete

**Task 2.4: Registration State Management**
- [ ] Implement in-memory storage
- [ ] Add registration creation
- [ ] Add registration update (refresh)
- [ ] Add registration deletion
- [ ] Implement expiration timer

### Phase 3: Integration (Week 3)

**Task 3.1: SIP Proxy Integration**
- [ ] Connect register server to existing proxy
- [ ] Implement call routing to registered contacts
- [ ] Add registration lookup in INVITE handling
- [ ] Test end-to-end call flow

**Task 3.2: Multi-registration Support**
- [ ] Support multiple contacts per instance
- [ ] Handle concurrent registrations
- [ ] Implement load balancing across contacts
- [ ] Add fallback logic

**Task 3.3: NAT Traversal**
- [ ] Implement STUN/TURN support if needed
- [ ] Add Via received/rport handling
- [ ] Test with clients behind NAT
- [ ] Document NAT requirements

### Phase 4: Testing & Validation (Week 4)

**Task 4.1: Softphone Testing**
- [ ] Test with Zoiper
- [ ] Test with MicroSIP
- [ ] Test with Linphone
- [ ] Test with X-Lite
- [ ] Test with mobile clients (iOS/Android)

**Task 4.2: Automated Tests**
- [ ] Unit tests for registration logic
- [ ] Unit tests for authentication
- [ ] Integration tests for call flow
- [ ] Load tests for concurrent registrations

**Task 4.3: Security Testing**
- [ ] Test unauthorized registration attempts
- [ ] Test credential spoofing
- [ ] Test registration flooding
- [ ] Verify rate limiting

### Phase 5: API Endpoints (Week 5)

**Task 5.1: SIP Credentials Management API**
- [ ] Create API handler for SIP credentials
- [ ] POST /api/v1/sipcredentials - Set/update SIP password
- [ ] GET /api/v1/sipcredentials - Get current SIP password status
- [ ] DELETE /api/v1/sipcredentials - Remove custom SIP password
- [ ] Add request validation
- [ ] Add authentication middleware (require master key or instance token)
- [ ] Add Swagger documentation

**Task 5.2: Registration Status API**
- [ ] GET /api/v1/sipregistrations - List all active registrations
- [ ] GET /api/v1/sipregistrations/{token} - Get registration for specific instance
- [ ] DELETE /api/v1/sipregistrations/{token} - Force unregister specific instance
- [ ] Add pagination support
- [ ] Add filtering (by instance, by status)
- [ ] Add Swagger documentation

**Task 5.3: Production Readiness**
- [ ] Redis integration for registration storage
- [ ] Monitoring & logging
- [ ] Configuration management
- [ ] Documentation

### API Endpoints Specification

#### POST /api/v1/sipcredentials
Set or update SIP password for an instance.

**Request**:
```json
{
  "token": "abc123-instance-id",
  "password": "secure-password-123"
}
```

**Response**:
```json
{
  "success": true,
  "message": "SIP password updated successfully"
}
```

**Validation**:
- Token must exist and be verified
- Password minimum 8 characters
- Bearer token authentication required

#### GET /api/v1/sipcredentials
Check if instance has custom SIP password.

**Request**:
```
GET /api/v1/sipcredentials?token=abc123-instance-id
```

**Response**:
```json
{
  "token": "abc123-instance-id",
  "has_custom_password": true,
  "created_at": "2026-06-25T10:00:00Z",
  "updated_at": "2026-06-25T10:00:00Z"
}
```

#### DELETE /api/v1/sipcredentials
Remove custom SIP password (revert to no-password mode).

**Request**:
```
DELETE /api/v1/sipcredentials?token=abc123-instance-id
```

**Response**:
```json
{
  "success": true,
  "message": "SIP password removed successfully"
}
```

#### GET /api/v1/sipregistrations
List all active SIP registrations.

**Request**:
```
GET /api/v1/sipregistrations?page=1&limit=50
```

**Response**:
```json
{
  "registrations": [
    {
      "token": "abc123-instance-id",
      "contact_uri": "sip:192.168.1.100:5060",
      "expires_at": "2026-06-25T11:00:00Z",
      "registered_at": "2026-06-25T10:00:00Z",
      "user_agent": "Zoiper 5.6.3",
      "remote_ip": "192.168.1.100"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 50
}
```

#### GET /api/v1/sipregistrations/{token}
Get registration status for specific instance.

**Request**:
```
GET /api/v1/sipregistrations/abc123-instance-id
```

**Response**:
```json
{
  "token": "abc123-instance-id",
  "contact_uri": "sip:192.168.1.100:5060",
  "expires_at": "2026-06-25T11:00:00Z",
  "registered_at": "2026-06-25T10:00:00Z",
  "last_refresh": "2026-06-25T10:30:00Z",
  "user_agent": "Zoiper 5.6.3",
  "remote_ip": "192.168.1.100",
  "status": "active"
}
```

#### DELETE /api/v1/sipregistrations/{token}
Force unregister specific instance (admin action).

**Request**:
```
DELETE /api/v1/sipregistrations/abc123-instance-id
```

**Response**:
```json
{
  "success": true,
  "message": "Registration removed successfully"
}
```

**Note**: This is an admin action that removes the registration from the server, requiring master key authentication.

**Task 5.1: Redis Integration**
- [ ] Migrate from in-memory to Redis
- [ ] Implement Redis connection pool
- [ ] Add Redis key expiration
- [ ] Implement graceful degradation

**Task 5.2: Monitoring & Logging**
- [ ] Add registration success/failure metrics
- [ ] Log all registration attempts
- [ ] Add Prometheus metrics
- [ ] Implement alerting

**Task 5.3: Configuration**
- [ ] Add configurable port binding
- [ ] Add realm configuration
- [ ] Add TTL configuration
- [ ] Add rate limiting configuration

**Task 5.4: Documentation**
- [ ] Write installation guide
- [ ] Write softphone configuration guide
- [ ] Write API documentation
- [ ] Write troubleshooting guide

## Configuration Example

### QuePasa Server Configuration

```yaml
sip_register:
  enabled: true
  listen: "0.0.0.0:5060"
  realm: "quepasa"
  ttl: 3600  # 1 hour
  max_registrations: 100
  allow_no_password: true  # Default mode
  require_password: false  # Set to true to enforce custom passwords
  rate_limit:
    enabled: true
    requests_per_minute: 60
  storage:
    type: "memory"  # or "redis"
    redis:
      addr: "localhost:6379"
      password: ""
      db: 0
```

### Softphone Configuration Example (Zoiper - Default Mode)

```
Account: QuePasa WhatsApp
Username: abc123-instance-id
Password: (leave empty)
Domain: quepasa-server.example.com
Port: 5060
Transport: UDP
```

### Softphone Configuration Example (Zoiper - Custom Password)

```
Account: QuePasa WhatsApp
Username: abc123-instance-id
Password: custom-sip-password
Domain: quepasa-server.example.com
Port: 5060
Transport: UDP
```

## Security Considerations

### Authentication
- Always use Digest Authentication (never plain text)
- Implement proper nonce generation (random, unpredictable)
- Use secure random number generator for nonces
- Limit nonce validity time
- Support MD5 (SIP standard) and consider SHA-256 if clients support

### Authorization
- Verify token exists in QpServer database
- Verify instance is verified
- Prevent cross-instance registration (token must match)
- If custom password set, verify bcrypt hash
- If no custom password, allow registration (token is already secure)
- Implement access control lists for admin APIs if needed

### Rate Limiting
- Limit registration attempts per IP
- Limit registration attempts per phone number
- Implement exponential backoff for failures
- Block abusive IPs temporarily

### Transport Security
- Consider TLS (SIPS) for secure registration
- Support TCP for reliability
- Implement TCP fallback for large messages

## Testing Strategy

### Manual Testing Checklist

1. **Basic Registration (No Password)**
   - [ ] REGISTER without password → 200 OK (default mode)
   - [ ] REGISTER with invalid token → 403 Forbidden
   - [ ] Registration appears in storage
   - [ ] Expires after TTL
   - [ ] Re-registration updates expiration

2. **Custom Password Mode**
   - [ ] Set password via API
   - [ ] REGISTER without password → 401 Unauthorized
   - [ ] REGISTER with wrong password → 403 Forbidden
   - [ ] REGISTER with correct password → 200 OK
   - [ ] Remove password via API
   - [ ] Revert to no-password mode

3. **API Endpoints**
   - [ ] POST /api/v1/sipcredentials - Set password
   - [ ] GET /api/v1/sipcredentials - Check password status
   - [ ] DELETE /api/v1/sipcredentials - Remove password
   - [ ] GET /api/v1/sipregistrations - List registrations
   - [ ] GET /api/v1/sipregistrations/{token} - Get registration
   - [ ] DELETE /api/v1/sipregistrations/{token} - Force unregister

4. **Re-registration**
   - [ ] Refresh before expiration
   - [ ] Contact URI update
   - [ ] Expiration timer reset

5. **De-registration**
   - [ ] REGISTER with Expires: 0
   - [ ] Registration removed
   - [ ] Calls no longer routed

6. **Call Routing**
   - [ ] Incoming WhatsApp call rings softphone
   - [ ] Answer call
   - [ ] Two-way audio
   - [ ] Call termination

7. **Softphone Compatibility**
   - [ ] Zoiper: Register and call (no password)
   - [ ] Zoiper: Register and call (with password)
   - [ ] MicroSIP: Register and call
   - [ ] Linphone: Register and call
   - [ ] Mobile clients: Register and call

### Automated Testing

```go
// Example test structure
func TestSIPRegisterHandler(t *testing.T) {
    tests := []struct {
        name         string
        token        string
        password     string
        expectStatus int
        expectError  bool
    }{
        {
            name:         "Valid registration (no password)",
            token:        "abc123-instance",
            password:     "",
            expectStatus: 200,
            expectError:  false,
        },
        {
            name:         "Invalid token",
            token:        "invalid-instance",
            password:     "",
            expectStatus: 403,
            expectError:  true,
        },
        {
            name:         "Valid registration (with password)",
            token:        "abc123-instance",
            password:     "custom-password",
            expectStatus: 200,
            expectError:  false,
        },
        {
            name:         "Wrong password",
            token:        "abc123-instance",
            password:     "wrong-password",
            expectStatus: 403,
            expectError:  true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Risks & Mitigations

### Risk 1: Breaking Existing SIP Proxy
- **Mitigation**: Implement register server as separate component
- **Mitigation**: Extensive testing before merging
- **Mitigation**: Feature flag for enable/disable

### Risk 2: Security Vulnerabilities
- **Mitigation**: Code review focused on authentication
- **Mitigation**: Security testing before production
- **Mitigation**: Start with in-memory only (no external exposure)

### Risk 3: Softphone Compatibility Issues
- **Mitigation**: Test with multiple softphones
- **Mitigation**: Follow SIP RFC strictly
- **Mitigation**: Comprehensive logging for troubleshooting

### Risk 4: Performance Impact
- **Mitigation**: Benchmark registration handling
- **Mitigation**: Implement rate limiting
- **Mitigation**: Use efficient storage (Redis)

## Success Criteria

1. **Functional**
   - Users can register from any standard SIP softphone
   - Authentication works with token as username
   - Default mode: registration without password
   - Optional mode: registration with custom password
   - Registered softphones receive inbound WhatsApp calls
   - Multiple softphones can register per instance
   - API endpoints for password management
   - API endpoints for registration status

2. **Non-Functional**
   - Registration latency < 100ms
   - Support 100+ concurrent registrations
   - No impact on existing SIP proxy performance
   - Zero security vulnerabilities in authentication

3. **Compatibility**
   - Works with Zoiper, MicroSIP, Linphone, X-Lite
   - Works on Windows, macOS, Linux, iOS, Android
   - Compliant with SIP RFC 3261

## Open Questions

1. **Port Selection**
   - Use standard SIP port (5060) or custom port?
   - Impact on existing services?

2. **Transport Protocol**
   - UDP only or also TCP?
   - TLS (SIPS) support required?

3. **Storage Backend**
   - Start with in-memory and migrate to Redis?
   - Or implement Redis from the start?

4. **Multi-registration**
   - Allow multiple softphones per instance?
   - Ring all or first to answer?

5. **NAT Traversal**
   - STUN/TURN support required?
   - Or rely on existing proxy NAT handling?

6. **Password Policy**
   - Should custom SIP passwords have complexity requirements?
   - Minimum length? Special characters?

7. **Default Security Mode**
   - Allow without password by default (simpler)?
   - Or require custom password setup first (more secure)?

## Next Steps

1. Review this plan with stakeholders
2. Decide on open questions (especially default security mode)
3. Create development branch
4. Begin Phase 1: Research & Setup
5. Set up test environment with softphones
6. Start implementation

## Changes in v2.0

**Major Changes**:
- Changed from phone number to token as SIP username (resolves multi-instance conflicts)
- Added default no-password authentication mode (simpler deployment)
- Added optional custom password support via sip_credentials table
- Added comprehensive API endpoints for SIP credentials and registration management
- Simplified authentication flow (token lookup instead of phone number mapping)

**Authentication Evolution**:
- v1.0: Username=Phone, Password=Token (conflicts with multiple instances)
- v2.0: Username=Token, Password=Optional (no conflicts, token already secure)

**New Features**:
- SIP credentials management API (POST, GET, DELETE)
- Registration status API (list, get by token, force unregister)
- Password rotation support via API
- Bcrypt password hashing
- Foreign key cascade delete for credentials

## References

- RFC 3261: SIP: Session Initiation Protocol
- RFC 8760: The Session Initiation Protocol (SIP) Digest Access Authentication
- [emiago/sipgo GitHub](https://github.com/emiago/sipgo)
- [SIPgo Reddit Discussion](https://www.reddit.com/r/golang/comments/uwwdze/sipgo_for_writing_fast_sip_servers/)
- [Digest Authentication with SIP - Oracle](https://docs.oracle.com/en/industries/communications/session-border-controller/9.1.0/configuration/digest-authentication-sip1.html)
- [SIP REGISTER Method Explained](https://pgkrishna.medium.com/sip-understanding-register-method-and-authentication-process-a008f884ff18)

## Appendix: Example Code Structure

```
quepasa/
├── sip/
│   ├── register/
│   │   ├── server.go           # Register server main logic
│   │   ├── handler.go          # REGISTER method handler
│   │   ├── auth.go             # Digest authentication
│   │   ├── storage.go          # Registration state storage
│   │   ├── storage_redis.go    # Redis implementation
│   │   ├── storage_memory.go   # In-memory implementation
│   │   └── models.go           # Registration data models
│   ├── proxy/
│   │   └── ...                 # Existing proxy code
│   └── common/
│       └── sip.go              # Common SIP utilities
├── config/
│   └── sip_register.yml        # Register server config
└── cmd/
    └── quepasa/
        └── main.go             # Entry point (if needed)
```

---

**Document Version**: 2.0
**Last Updated**: 2026-06-25