# INVESTIGATION: SIP Register Server for QuePasa

**Date**: 2026-06-25
**Status**: In Progress
**Related Plan**: `/mnt/quepasa/docs/PLAN-sip-register-server.md`

## Current Architecture Analysis

### Existing SIP Proxy Implementation

**Location**: `/mnt/quepasa/src/sipproxy/`

**Key Files**:
- `sipproxy.go` - Main initialization and environment settings
- `sipproxy_manager.go` - Core manager coordinating SIP proxy modules
- `sipproxy_listener.go` - SIP server and UDP listener using sipgo
- `sipproxy_call_manager_sipgo.go` - SIP call lifecycle management
- `sipproxy_network_manager.go` - Network settings and STUN/UPnP support
- `sipproxy_response_handler.go` - SIP response handling

**Technology Stack**:
- **Library**: `github.com/emiago/sipgo` v0.33.0 (✅ Confirmed)
- **Protocol**: UDP/TCP SIP
- **Supported Methods**: INVITE, BYE, CANCEL, OPTIONS
- **NOT Currently Implemented**: REGISTER method (server-side)

### Instance and Authentication Model

**Server/Instance Model**: `QpServer` struct in `/mnt/quepasa/src/models/qp_server.go`

```go
type QpServer struct {
    library.LogStruct
    whatsapp.WhatsappOptions

    Token string      `db:"token" json:"token"`           // Instance identifier
    Wid   sql.NullString `db:"wid" json:"wid"`            // WhatsApp Session ID
    Verified bool      `db:"verified" json:"verified"`
    Devel    bool      `db:"devel" json:"devel"`
    Metadata QpMetadata `db:"metadata" json:"metadata,omitempty"`

    User      sql.NullString `db:"user" json:"user,omitempty"` // User identifier
    Timestamp time.Time      `db:"timestamp" json:"timestamp,omitempty"`
}
```

**Credential Verification Methods**:
1. `FindByToken(token string) (*QpServer, error)` - Find instance by token
2. `FindForUser(token string, user string) (*QpServer, error)` - Verify token + user pair

**Database Table**: `servers` with columns:
- `token` - Instance identifier (will serve as SIP password)
- `wid` - WhatsApp session ID (contains phone number in JID format)
- `user` - Optional user identifier (may serve as SIP username)
- `verified` - Account verification status

### SIP Credentials Table (New)

**Purpose**: Store optional custom SIP passwords per instance

**Schema**:
```sql
CREATE TABLE sip_credentials (
    token TEXT PRIMARY KEY,  -- Reference to servers.token (PK)
    password_hash TEXT NOT NULL,  -- Hashed SIP password
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (token) REFERENCES servers(token) ON DELETE CASCADE
);
```

**Go Model**:
```go
type QpSipCredentials struct {
    Token        string    `db:"token" json:"token"`
    PasswordHash string    `db:"password_hash" json:"-"`
    CreatedAt    time.Time `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
```

**Security Notes**:
- Passwords hashed using bcrypt (cost 12)
- No password stored in plain text
- Support password rotation (update hash, keep same token)
- Cascade delete when instance is deleted

### Current SIP Proxy Limitations

**What EXISTS**:
- ✅ SIP server using sipgo
- ✅ INVITE handling (outbound calls)
- ✅ BYE/CANCEL handling
- ✅ RTP media streaming
- ✅ Integration with WhatsApp VoIP calls

**What DOES NOT EXIST**:
- ❌ REGISTER method handler (server-side)
- ❌ SIP user registration state management
- ❌ Registration expiration timers
- ❌ Contact URI storage
- ❌ Digest authentication for registration
- ❌ Registration lookup for inbound calls

### Proposed Authentication Scheme (Updated)

**Problem with Phone Number as Username**:
- Multiple instances can have the same WhatsApp phone number
- Would cause registration conflicts
- Token is unique per instance, phone is not

**SIP Username (Recommended)**:
- Use `token` field from QpServer (instance identifier)
- Unique per instance by design
- Already used for API authentication (secure)

**SIP Password**:
- **Default**: No password required (empty string)
- **Optional**: Custom password from separate SIP credentials table
- Token alone provides sufficient security

**Authentication Flow (Default - No Password)**:
```
1. User registers softphone:
   - Username: abc123-instance-id (token)
   - Password: (empty)

2. Server receives REGISTER:
   - Extract username from From header (token)
   - Look up QpServer by token
   - Verify instance exists and is verified
   - Return 200 OK if valid, 403 if invalid
```

**Authentication Flow (Custom Password)**:
```
1. User sets SIP password via API:
   - POST /api/v1/sipcredentials
   - Body: { "password": "custom-sip-password" }

2. User registers softphone:
   - Username: abc123-instance-id (token)
   - Password: custom-sip-password

3. Server receives REGISTER:
   - Extract username (token) from From header
   - Extract password from Authorization header
   - Look up QpServer by token
   - Check sip_credentials table for custom password
   - If custom password exists, verify digest
   - If no custom password, allow without password
   - Return 200 OK if valid, 403 if invalid
```

### Integration with Existing SIP Proxy

**Approach 1: Extend Current SIPProxyManager**
- Add REGISTER handler to existing `sipgo.Server`
- Share network manager and settings
- Minimal code duplication
- Risk: May affect existing INVITE flow

**Approach 2: Separate RegisterServer Component**
- Create new `SIPRegisterServer` struct
- Separate listener port or same port (REGISTER on same port)
- Independent lifecycle
- Better isolation and testing

**Recommendation**: Approach 1 (Extend SIPProxyManager)
- Single SIP server listening on same port
- sipgo supports multiple method handlers
- Reduced resource usage
- Simplifies deployment

### Storage Strategy Options

**Option 1: In-Memory (sync.Map)**
```go
type SIPRegistration struct {
    InstanceID   string
    PhoneNumber  string
    ContactURI   string
    ExpiresAt    time.Time
    RegisteredAt time.Time
    UserAgent    string
    CallID       string
}

var registrations sync.Map
```
- ✅ Simple, fast, no external dependencies
- ✅ Sufficient for development/single-instance deployment
- ❌ Lost on restart
- ❌ Not shared across multiple server instances

**Option 2: Redis**
```go
type RedisRegistrationStore struct {
    client *redis.Client
}

// Key format: "sipreg:{instance_id}:{phone_number}"
// TTL set automatically
```
- ✅ Persistent across restarts
- ✅ Shared across instances
- ✅ Native TTL support
- ✅ High performance
- ❌ Requires Redis deployment
- ❌ Additional infrastructure

**Recommendation**: Start with in-memory for PoC, plan Redis migration for production.

### Code Structure for REGISTER Handler

**New File**: `/mnt/quepasa/src/sipproxy/sipproxy_register.go`

```go
package sipproxy

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "strings"
    "sync"
    "time"

    "github.com/emiago/sipgo/sip"
    qplog "github.com/nocodeleaks/quepasa/qplog"
)

type SIPRegisterHandler struct {
    logger       qplog.Logger
    registrations sync.Map  // In-memory storage
    realm        string
    nonceStore   sync.Map  // For digest auth
}

func NewSIPRegisterHandler(logger qplog.Logger, realm string) *SIPRegisterHandler {
    return &SIPRegisterHandler{
        logger: logger,
        realm:  realm,
    }
}

// HandleRegister processes SIP REGISTER requests
func (rh *SIPRegisterHandler) HandleRegister(req *sip.Request, tx sip.ServerTransaction) {
    // 1. Check if request has Authorization header
    // 2. If not, send 401 Unauthorized with WWW-Authenticate
    // 3. If yes, verify digest response
    // 4. Extract username (phone number) and password (token)
    // 5. Look up instance in database
    // 6. Verify phone number matches instance
    // 7. Store or update registration
    // 8. Send 200 OK with Contact and Expires

    // Implementation details...
}

// GenerateNonce creates a unique nonce for digest authentication
func (rh *SIPRegisterHandler) GenerateNonce() string {
    // Implementation...
}

// VerifyDigestResponse verifies the digest authentication response
func (rh *SIPRegisterHandler) VerifyDigestResponse(username, realm, nonce, response, uri, method, password string) bool {
    // Implementation...
}

// GetRegisteredContact returns the Contact URI for a phone number
func (rh *SIPRegisterHandler) GetRegisteredContact(phoneNumber string) (string, bool) {
    // Implementation...
}
```

### Integration with VoIP Call Flow

**Current Flow** (WhatsApp → SIP Proxy → Remote Server):
```
WhatsApp Call → VoIPManager → SIPProxyManager → Remote SIP Server
```

**New Flow** (WhatsApp → SIP Proxy → Registered Softphone):
```
WhatsApp Call → VoIPManager → SIPProxyManager
                          ↓
                   Lookup Registration
                          ↓
                 Registered Softphone Contact URI
                          ↓
                   Forward INVITE to Softphone
```

**Integration Point**:
- In `SIPCallManagerSipgo.SendInvite()`, before sending to remote SIP server:
  - Check if target phone number has active registration
  - If registered, send INVITE to registered Contact URI instead
  - If not registered, fall back to remote SIP server

**File to Modify**: `/mnt/quepasa/src/sipproxy/sipproxy_call_manager_sipgo.go`

### Softphone Configuration Examples

**Zoiper (Default - No Password)**:
```
Account Name: QuePasa WhatsApp
Username: abc123-instance-id
Password: (leave empty)
Domain: <quepasa-server-host>
Port: 5060
Transport: UDP
```

**Zoiper (Custom Password)**:
```
Account Name: QuePasa WhatsApp
Username: abc123-instance-id
Password: custom-sip-password
Domain: <quepasa-server-host>
Port: 5060
Transport: UDP
```

**MicroSIP (Default - No Password)**:
```
SIP Server: <quepasa-server-host>:5060
User: abc123-instance-id
Password: (leave empty)
```

**Linphone (Custom Password)**:
```
SIP Address: sip:abc123-instance-id@<quepasa-server-host>
Username: abc123-instance-id
Password: custom-sip-password
Domain: <quepasa-server-host>
Transport: UDP
```

### Open Questions for Implementation

1. **Multi-registration**:
   - Should we allow multiple softphones to register for the same instance?
   - If yes, ring all or first to answer?
   - Recommendation: Allow multi-registration, ring all (forking)

2. **NAT Traversal**:
   - Do we need STUN/TURN for softphones behind NAT?
   - Or rely on existing Via/rport handling?
   - Recommendation: Start with existing handling, add STUN if needed

3. **Registration TTL**:
   - Default expiration time?
   - Recommendation: 3600 seconds (1 hour), configurable

4. **Port Configuration**:
   - Use existing SIP proxy port or separate port?
   - Recommendation: Same port (5060), different method handlers

5. **Password Policy**:
   - Should custom SIP passwords have complexity requirements?
   - Recommendation: Minimum 8 characters, optional complexity check

6. **Default Auth Mode**:
   - Allow without password by default?
   - Or require custom password setup first?
   - Recommendation: Allow without password (token is already secure)

### Next Steps for Implementation

1. **Phase 1: Core Registration Logic**
   - Create `sipproxy_register.go` with handler
   - Implement digest authentication
   - Implement registration storage (in-memory)
   - Add unit tests

2. **Phase 2: Database Integration**
   - Create credential verification service
   - Implement phone number extraction from JID
   - Add error handling for invalid credentials

3. **Phase 3: SIP Proxy Integration**
   - Modify `SIPCallManagerSipgo` to check registrations
   - Add registration lookup before INVITE
   - Implement fallback to remote SIP server

4. **Phase 4: Testing**
   - Test with Zoiper, MicroSIP, Linphone
   - Test end-to-end call flow
   - Test authentication failures
   - Test registration expiration

5. **Phase 5: Production Readiness**
   - Add metrics and logging
   - Implement rate limiting
   - Add Redis storage support
   - Document configuration

### Dependencies and Requirements

**Existing Dependencies** (Already in go.mod):
- ✅ `github.com/emiago/sipgo v0.33.0` - SIP stack
- ✅ `github.com/icholy/digest v1.1.0` - Digest auth helpers

**New Dependencies** (Potentially needed):
- `github.com/redis/go-redis/v9` - Redis client (for production)

**No Breaking Changes**:
- Register handler is additive
- Existing SIP proxy functionality unchanged
- Can be enabled/disabled via config

### Testing Strategy

**Unit Tests**:
```go
// Test digest authentication
func TestVerifyDigestResponse(t *testing.T)

// Test registration storage
func TestRegistrationStorage(t *testing.T)

// Test expiration handling
func TestRegistrationExpiration(t *testing.T)
```

**Integration Tests**:
```go
// Test full registration flow
func TestRegistrationFlow(t *testing.T)

// Test call routing to registered softphone
func TestCallRouting(t *testing.T)
```

**Manual Testing**:
1. Configure softphone with instance credentials
2. Register successfully
3. Receive incoming WhatsApp call
4. Answer call from softphone
5. Verify two-way audio

## Conclusion

The SIP Register Server implementation is **technically feasible** and **well-aligned** with the existing QuePasa architecture:

- ✅ SIP library (sipgo) already integrated
- ✅ Instance model (`QpServer`) provides authentication data
- ✅ Existing SIP proxy infrastructure can be extended
- ✅ Minimal code changes required
- ✅ No breaking changes to existing functionality

**Estimated Effort**: 2-3 weeks for full implementation including testing.

**Risk Level**: Low (isolated feature, additive changes, well-defined scope).

---

**Next Action**: Review this investigation with stakeholders and proceed to Phase 1 implementation.

**Document Version**: 2.0
**Last Updated**: 2026-06-25
**Changes in v2.0**:
- Changed from phone number to token as SIP username (resolves conflicts)
- Added default no-password authentication mode
- Added sip_credentials table for custom password support
- Simplified authentication flow