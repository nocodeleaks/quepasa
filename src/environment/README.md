# QuePasa Environment Variables Documentation

This document describes all environment variables used by the QuePasa application, organized by category. **Total: 65 variables across 12 categories**.

## 📡 SIP Proxy Configuration

### Core Settings
- **`SIPPROXY_HOST`** - SIP server hostname (e.g., `sip.provider.com`)
  - **Important**: If this is set, SIP Proxy is **ACTIVE**. If empty, SIP Proxy is **INACTIVE**.
  - No default value - must be explicitly configured.

- **`SIPPROXY_PORT`** - SIP server port (default: `5060`)
- **`SIPPROXY_LOCALPORT`** - Local SIP listening port (default: `5060`)

### Network & NAT Settings
- **`SIPPROXY_PUBLICIP`** - Override public IP (leave empty for auto-discovery)
- **`SIPPROXY_STUNSERVER`** - STUN server for NAT discovery (default: `stun.l.google.com:19302`)
- **`SIPPROXY_USEUPNP`** - Enable UPnP port forwarding (default: `true`)

### Media & Protocol Settings
- **`SIPPROXY_MEDIAPORTS`** - RTP media port range (default: `10000-20000`)
- **`SIPPROXY_CODECS`** - Supported audio codecs (default: `PCMU,PCMA,G729`)
- **`SIPPROXY_USERAGENT`** - SIP User-Agent string (default: `QuePasa-SIP-Proxy/1.0`)

### Timing & Retry Settings
- **`SIPPROXY_TIMEOUT`** - SIP transaction timeout in seconds (default: `30`)
- **`SIPPROXY_RETRIES`** - SIP INVITE retry attempts (default: `3`)
- **`SIPPROXY_LOGLEVEL`** - SIP proxy specific log level (default: `info`)
- **`SIPPROXY_PROTOCOL`** - SIP server protocol (default: `UDP`)
- **`SIPPROXY_SDPSESSIONNAME`** - SDP session name

## 🔗 API/Web Server Configuration

- **`WEBAPIHOST`** - Web server bind host *(deprecated, use WEBSERVER_HOST)*
- **`WEBAPIPORT`** - Web server port (default: `31000`) *(deprecated, use WEBSERVER_PORT)*
- **`WEBSERVER_HOST`** - Web server bind host (fallback: `WEBAPIHOST`)
- **`WEBSERVER_PORT`** - Web server port (default: `31000`, fallback: `WEBAPIPORT`)
- **`WEBSOCKETSSL`** - Use SSL for WebSocket QR code (default: `false`)
- **`SIGNING_SECRET`** - Token for hash signing cookies
- **`MASTERKEY`** - Master key for super admin methods
- **`HTTPLOGS`** - Log HTTP requests (default: `false`)
- **`WEBHOOK_TIMEOUT`** - Webhook request timeout in milliseconds (default: `10000` = 10 seconds, minimum: `1`)
- **`API_TIMEOUT`** - API request timeout in milliseconds (default: `30000` = 30 seconds, minimum: `1`)
- **`API_PREFIX`** - API routes prefix

### Database User Seeding (First Startup Only)

**⚠️ Currently only used during initial database seeding** - These variables control the default user created when the database is empty on first startup:

- **`USER`** - Default username/email for initial database seeding
  - **Default behavior**: If not set, uses `"default@quepasa.io"` with **empty password** (INSECURE!)
  - **Recommended**: Always set both `USER` and `PASSWORD` in production environments
  - **Usage**: Only read during first application startup when no users exist in database
  - **Example**: `USER=admin@yourdomain.com`

- **`PASSWORD`** - Password for default user created during seeding
  - **Default behavior**: If `USER` is set but `PASSWORD` is empty, user creation will **FAIL** (security requirement)
  - **Security**: Must be a strong password (recommended: 12+ characters, mixed case, numbers, symbols)
  - **Usage**: Only used during first startup to create the initial admin user
  - **Example**: `PASSWORD=YourSecurePassword123!@#`
  - **Important**: After first startup, password changes must be done through the API or database directly

**Security Notes:**
- These variables are **only read once** during initial database seeding
- If database already has users, these variables are **ignored**
- Empty password is only allowed for legacy `default@quepasa.io` user (NOT recommended)
- For new users via `USER` variable, password validation is **mandatory**

## 💾 Database Configuration

- **`DBDRIVER`** - Database driver (default: `sqlite3`)
- **`DBHOST`** - Database host
- **`DBDATABASE`** - Database name
- **`DBPORT`** - Database port
- **`DBUSER`** - Database user
- **`DBPASSWORD`** - Database password
- **`DBSSLMODE`** - Database SSL mode

## 📱 WhatsApp Configuration

- **`READUPDATE`** - Global: Mark chat as read when receiving messages (default: `false`). Can be overridden per server.
- **`READRECEIPTS`** - Handle read receipts (default: `false`)
- **`CALLS`** - Handle calls (default: `false`)
- **`GROUPS`** - Handle group messages (default: `false`)
- **`BROADCASTS`** - Handle broadcast messages (default: `false`)
- **`HISTORYSYNCDAYS`** - History sync days
- **`PRESENCE`** - Presence state (default: `unavailable`)
- **`WAKEUP_HOUR`** - Single hour (0-23) to activate presence daily (e.g., `9` for 9 AM)
- **`WAKEUP_DURATION`** - Duration in seconds to keep presence online during scheduled wake-up (default: `10`)

### Individual Server Configuration

Each server can override global settings for `READUPDATE`, `GROUPS`, `CALLS`, `READRECEIPTS`, and `BROADCASTS`:
- Set to `true` or `1` to enable for this specific server
- Set to `false` or `-1` to disable for this specific server  
- Leave unset to use global environment variable value

## 📋 Logging Configuration

- **`LOGLEVEL`** - General log level
- **`WHATSMEOW_LOGLEVEL`** - Whatsmeow library log level
- **`WHATSMEOW_DBLOGLEVEL`** - Whatsmeow database log level

## ⚙️ General Application Settings

- **`MIGRATIONS`** - Enable database migrations (default: `true`)
- **`APP_TITLE`** - Application title for WhatsApp device list
- **`REMOVEDIGIT9`** - Remove digit 9 from phone numbers (default: `false`)
- **`SYNOPSISLENGTH`** - Synopsis length for messages (default: `50`)
- **`CACHELENGTH`** - Cache max items (default: `0` = unlimited)
- **`CACHEDAYS`** - Cache max days (default: `0` = unlimited)
- **`CONVERT_WAVE_TO_OGG`** - Convert wave to OGG (default: `true`)
- **`COMPATIBLE_MIME_AS_AUDIO`** - Treat compatible MIME as audio (default: `true`)
- **`CONVERT_PNG_TO_JPG`** - Convert PNG images to JPG using FFmpeg (default: `false`)
- **`ACCOUNTSETUP`** - Enable account creation (default: `true`)
- **`TESTING`** - Testing mode (default: `false`)

## 📋 Form/Web Interface Configuration

- **`FORM`** - Enable/disable web form interface (default: `true`)
- **`FORM_PREFIX`** - Form endpoint path prefix (default: `form`)

## 📊 Metrics Configuration

- **`METRICS`** - Enable/disable metrics endpoint (default: `true`)
- **`METRICS_PREFIX`** - Metrics endpoint path prefix (default: `metrics`)
- **`METRICS_DASHBOARD`** - Enable/disable metrics dashboard endpoint (default: `true`)
- **`METRICS_DASHBOARD_PREFIX`** - Metrics dashboard endpoint path prefix (default: `dashboard`)

## 🐰 RabbitMQ Configuration

- **`RABBITMQ_QUEUE`** - RabbitMQ queue name
- **`RABBITMQ_CONNECTIONSTRING`** - RabbitMQ connection string
- **`RABBITMQ_CACHELENGTH`** - RabbitMQ cache length (default: `0`)

## 📖 Swagger Configuration

- **`SWAGGER`** - Enable/disable Swagger UI (default: `true`)
- **`SWAGGER_PREFIX`** - Swagger UI path prefix (default: `swagger`)

## 🌐 WebServer Configuration

- **`WEBSERVER_HOST`** - Web server bind host (fallback: `WEBAPIHOST`)
- **`WEBSERVER_PORT`** - Web server port (default: `31000`, fallback: `WEBAPIPORT`)
- **`WEBSERVER_LOGS`** - Enable web server HTTP logs (default: `false`, fallback: `HTTPLOGS`)

## 📱 Whatsmeow Configuration

- **`DISPATCHUNHANDLED`** - Dispatch unhandled messages (default: `false`)
- **`WHATSMEOW_LOGLEVEL`** - Whatsmeow library log level
- **`WHATSMEOW_DBLOGLEVEL`** - Whatsmeow database log level
- **`WHATSMEOW_USE_RETRY_MESSAGE_STORE`** - Persist outgoing messages in database for retry receipts after process restarts (default: `false`)

## 📋 Current Working Configuration

Based on our successful NAT traversal tests with `sip.provider.com:5060`:

```env
SIPPROXY_HOST=sip.provider.com
SIPPROXY_PORT=5060
SIPPROXY_LOCALPORT=5060
SIPPROXY_STUNSERVER=stun.l.google.com:19302
SIPPROXY_USEUPNP=true
SIPPROXY_MEDIAPORTS=10000-20000
SIPPROXY_CODECS=PCMU,PCMA,G729
SIPPROXY_USERAGENT=QuePasa-SIP-Proxy/1.0
SIPPROXY_TIMEOUT=30
SIPPROXY_RETRIES=3
```

## 🧪 Testing Instructions

### Official Go Testing Convention

This package follows **official Go testing conventions**. We **NO longer use** separate `tests/` folders.

#### ✅ Correct Approach (Go Standard)
```bash
# Run tests from project root where environment variables are available
cd /path/to/quepasa/src
go test -v github.com/nocodeleaks/quepasa/environment

# Or run tests from environment directory
cd environment
go test -v
```

#### 📁 Test File Naming Convention
- **`*_test.go`** - Standard Go test files
- **`TestFunctionName`** - Test function names must start with `Test`
- **Example:** `environment_test.go`, `sipproxy_test.go`

#### 🔧 VS Code Integration
The environment package automatically loads `.env` files when running via:
- **F5 Debug** - Uses `launch.json` configuration with `envFile: "${workspaceFolder}/.env"`
- **Build Tasks** - Automatically copies `.env` to `.dist/` folder

#### 🚀 Build Tasks Available
1. **`Build and run`** - Standard Go build (default task)
2. **`Build and copy env`** - Build + automatically copy `.env` to `.dist/`
3. **`Copy env to dist`** - Just copy `.env` to distribution folder

#### 📅 Environment File Versioning
The `.env` file includes automatic versioning headers:
```env
# ================================================================
# QUEPASA ENVIRONMENT CONFIGURATION
# ================================================================
# Version: 20251005120000 (YYYYMMDDHHMMSS)
# Last Updated: 5 de outubro de 2025 - 12:00:00
# Build Target: Development/Production Environment
# Source: Updated documentation for develop branch
# Environment Package: 63 variables across 12 categories
# ================================================================
```

#### 📁 File Structure
```
project/
├── .env                    # ← Root .env (VS Code loads this)
├── .dist/
│   ├── .env               # ← Copied during build
│   └── quepasa.exe        # ← Compiled executable
└── environment/
    ├── environment.go     # ← Main environment package
    ├── *_test.go         # ← Test files (Go standard)
    └── README.md         # ← This documentation
```

#### 📝 Test Categories Available
1. **`TestEnvironmentPackageStructure`** - Verifies all environment files exist
2. **`TestEnvironmentVariablesDefault`** - Tests default values
3. **`TestEnvironmentVariablesFromSystem`** - Tests real environment loading
4. **`TestSIPProxyActivationLogic`** - Tests SIP proxy HOST-based activation
5. **`TestEnvironmentSettingsSingleton`** - Tests Settings initialization
6. **`TestEnvironmentVariablesCoverage`** - Tests all 47 environment variables

#### 🎯 Running Specific Tests
```bash
# Run specific test
go test -v -run TestSIPProxyActivationLogic

# Run with timeout
go test -v -timeout=30s

# Run tests and show coverage
go test -v -cover
```

#### ⚠️ Important Notes
- Environment variables are loaded from VS Code's `.env` injection when debugging
- When running via `go test` in terminal, default values are used
- SIP Proxy activation depends on `SIPPROXY_HOST` being set
- All 47 environment variables are tested for accessibility

## 💡 Usage Examples

```go
// Check if SIP Proxy is active
if environment.Settings.SIPProxy.Enabled() {
    host := environment.Settings.SIPProxy.Host()
    port := environment.Settings.SIPProxy.Port()
    // SIP Proxy is active
}

// Get database parameters
dbParams := environment.Settings.Database.GetDBParameters()

// Check WhatsApp call handling
if environment.Settings.WhatsApp.Calls().ToBoolean(false) {
    // Handle calls
}

// Check wake-up timer settings
if len(environment.Settings.WhatsApp.WakeUpHour) > 0 {
    wakeUpHour := environment.Settings.WhatsApp.WakeUpHour
    duration := environment.Settings.WhatsApp.WakeUpDuration
    // Wake-up timer is configured (single hour)
}
```
