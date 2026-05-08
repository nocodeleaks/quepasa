# GET /api/system/environment - Environment Configuration Endpoint

## Overview

The `GET /api/system/environment` endpoint provides access to QuePasa environment configuration. The response varies based on authentication:

- **Without authentication (anonymous)**: Returns a public preview of non-sensitive settings
- **With master key**: Returns full environment configuration

## Authentication Modes

### Anonymous Request (No Master Key)
```bash
curl http://localhost:31000/api/system/environment
```

**Response** (200 OK):
```json
{
  "result": "success",
  "message": "successfully retrieved public environment settings preview",
  "preview": {
    "groups": "true",
    "broadcasts": "false",
    "read_receipts": "false",
    "calls": "true",
    "history_sync": "30 days",
    "log_level": "INFO",
    "db_log_level": "WARN",
    "retry_message_store": "true",
    "presence": "available",
    "read_update": "true",
    "default_api_version": "v4",
    "wakeup_hour": "00:00",
    "wakeup_duration": "3600 seconds"
  }
}
```

### With Master Key
```bash
curl -H "X-QUEPASA-MASTERKEY: your-master-key" \
  http://localhost:31000/api/system/environment
```

**Response** (200 OK):
```json
{
  "result": "success",
  "message": "successfully retrieved full environment settings",
  "settings": {
    "api": {
      "domain": "localhost:31000",
      "port": 31000,
      "enabled": true,
      "relaxedSessions": true,
      "defaultVersion": "v4",
      "masterKey": ""
    },
    "database": {
      "driver": "postgres",
      "host": "db.example.com",
      "port": 5432,
      "database": "quepasa",
      "user": "quepasa",
      "password": "***masked***",
      "maxOpenConnections": 10,
      "maxIdleConnections": 5,
      "connectionMaxLifetime": 3600
    },
    "webserver": {
      "environment": "production",
      "corsOrigin": "*",
      "corsAllowCredentials": false,
      "readTimeout": 30,
      "writeTimeout": 30,
      "idleTimeout": 30
    },
    "whatsapp": {
      "groups": "true",
      "broadcasts": "false",
      "readReceipts": "false",
      "calls": "true",
      "historySyncDays": 30,
      "synopsisLength": 50,
      "readUpdate": "true",
      "presence": "available",
      "wakeUpHour": "00:00",
      "wakeUpDuration": 3600
    },
    "whatsmeow": {
      "logLevel": "INFO",
      "dbLogLevel": "WARN",
      "useRetryMessageStore": true
    },
    "logging": {
      "level": "INFO",
      "format": "json",
      "output": "stdout"
    },
    "cache": {
      "backend": "memory",
      "maxMessages": 800,
      "retentionDays": 7,
      "initFallback": true
    },
    "rabbitmq": {
      "connectionString": "",
      "enabled": false
    },
    "redis": {
      "host": "localhost",
      "port": 6379,
      "poolSize": 10,
      "maxRetries": 3
    }
  },
  "masterKeyConfigured": true
}
```

## Public Settings Preview (Anonymous)

The following settings are exposed in the **public preview** (no authentication required):

| Field | Description | Format | Example |
|-------|-------------|--------|---------|
| `groups` | Group messaging enabled | Boolean state | `"true"`, `"false"`, `"forced true"`, `"unset"` |
| `broadcasts` | Broadcast messaging enabled | Boolean state | `"false"` |
| `read_receipts` | Read receipt notifications | Boolean state | `"false"` |
| `calls` | Incoming calls handling | Boolean state | `"true"` |
| `history_sync` | Message history sync depth | Duration | `"30 days"` or `"disabled"` |
| `log_level` | Application log level | String | `"INFO"`, `"DEBUG"`, `"WARN"`, `"ERROR"` |
| `db_log_level` | Database log level | String | `"WARN"`, `"ERROR"` |
| `retry_message_store` | Retry queue backend enabled | Boolean | `"true"`, `"false"` |
| `presence` | Presence/availability mode | String | `"available"`, `"unavailable"`, `"composing"` |
| `read_update` | Mark messages as read on send | Boolean state | `"true"`, `"false"` |
| `default_api_version` | Default version used by the unversioned `/api/...` alias | String | `"v4"`, `"v5"` |
| `wakeup_hour` | Service wake-up time (optional) | Time | `"00:00"` |
| `wakeup_duration` | Wake-up duration (optional) | Duration | `"3600 seconds"` |

## Hidden Settings (Master Key Only)

The following sensitive settings are **hidden** from anonymous requests:

### Database Configuration
- Connection string
- Host, port, database name
- Username, password
- Connection pool settings

### API Configuration
- Domain binding
- CORS settings
- MasterKey status (whether configured)
- Session settings

### Cache & Storage
- Backend type and configuration
- Redis credentials
- RabbitMQ connection details

### Security & Performance
- TLS/SSL settings
- Rate limiting configuration
- Timeout values
- Connection limits

## Use Cases

### Public Discovery
```bash
# Client discovers server capabilities without credentials
curl http://localhost:31000/api/system/environment

# Result: Can see what features are enabled (groups, calls, broadcasts)
# but cannot see infrastructure details
```

### Full Configuration Access
```bash
# Administrator retrieves complete configuration with credentials
curl -H "X-QUEPASA-MASTERKEY: $(cat .env | grep MASTERKEY | cut -d= -f2)" \
  http://localhost:31000/api/system/environment

# Result: Full environment settings for backup, debugging, or migration
```

### Integration Discovery
```bash
# Third-party app discovers server capabilities
curl http://localhost:31000/api/system/environment | jq '.preview'

# Result: Can determine which features are available (e.g., call handling enabled)
```

## Response Status Codes

| Code | Scenario |
|------|----------|
| 200 | Successfully retrieved settings |
| 400 | Invalid request parameters |

## Security Notes

- **No authentication required** for public preview
- **Master key recommended** for full settings access
- Sensitive data (passwords, keys) is **never exposed** in responses
- Preview contains **feature availability only**, not infrastructure details
- Use in admin panels to inform feature availability to users
