# QuePasa Environment Variables Documentation

This document describes all environment variables used by the QuePasa application, organized by category.

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

## 🔗 API/Web Server Configuration

- **`WEBAPIHOST`** - Web server bind host
- **`WEBAPIPORT`** - Web server port (default: `31000`)
- **`WEBSOCKETSSL`** - Use SSL for WebSocket QR code (default: `false`)
- **`SIGNING_SECRET`** - Token for hash signing cookies
- **`MASTERKEY`** - Master key for super admin methods
- **`HTTPLOGS`** - Log HTTP requests (default: `false`)

## 💾 Database Configuration

- **`DBDRIVER`** - Database driver (default: `sqlite3`)
- **`DBHOST`** - Database host
- **`DBDATABASE`** - Database name
- **`DBPORT`** - Database port
- **`DBUSER`** - Database user
- **`DBPASSWORD`** - Database password
- **`DBSSLMODE`** - Database SSL mode

## 📱 WhatsApp Configuration

- **`READUPDATE`** - Mark chat read when sending messages (default: `false`)
- **`READRECEIPTS`** - Handle read receipts (default: `false`)
- **`CALLS`** - Handle calls (default: `false`)
- **`GROUPS`** - Handle group messages (default: `false`)
- **`BROADCASTS`** - Handle broadcast messages (default: `false`)
- **`HISTORYSYNCDAYS`** - History sync days
- **`PRESENCE`** - Presence state (default: `unavailable`)

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
- **`ACCOUNTSETUP`** - Enable account creation (default: `true`)
- **`TESTING`** - Testing mode (default: `false`)
- **`DISPATCHUNHANDLED`** - Dispatch unhandled messages (default: `false`)

## 🐰 RabbitMQ Configuration

- **`RABBITMQ_QUEUE`** - RabbitMQ queue name
- **`RABBITMQ_CONNECTIONSTRING`** - RabbitMQ connection string
- **`RABBITMQ_CACHELENGTH`** - RabbitMQ cache length (default: `0`)

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
```
