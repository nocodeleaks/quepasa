<!-- VERSION: 3.26.0220.0025 -->
[![Go Build](https://github.com/nocodeleaks/quepasa/actions/workflows/go.yml/badge.svg)](https://github.com/nocodeleaks/quepasa/actions/workflows/go.yml)

<p align="center">
	<img src="https://github.com/nocodeleaks/quepasa/raw/main/src/assets/favicon.png" alt="Quepasa-logo" width="100" />	
	<p align="center">QuePasa is an open-source, free license software to exchange messages with WhatsApp Platform</p>
</p>
<hr />
<p align="left">
	<img src="https://telegram.org/favicon.ico" alt="Telegram-logo" width="32" />
	<span>Chat with us on Telegram: </span>
	<a href="https://t.me/quepasa_api" target="_blank">Group</a>
	<span> || </span>
	<a href="https://t.me/quepasa_channel" target="_blank">Channel</a>
</p>
<p align="left">
	<span>Special thanks to <a target="_blank" href="https://agenciaoctos.com.br">Lukas Prais</a>, who developed this logo.</span>
</p>
<hr />

# QuePasa

> A micro web-application to make web-based WhatsApp bots easy to write.

**Current Version:** `3.26.0220.0025`

[![Run in Postman](https://run.pstmn.io/button.svg)](https://god.gw.postman.com/run-collection/5047984-405506cf-59f5-479e-b512-4ba5b935411b?action=collection%2Ffork&source=rip_markdown&collection-url=entityId%3D5047984-405506cf-59f5-479e-b512-4ba5b935411b%26entityType%3Dcollection%26workspaceId%3Dbd72aaba-0c31-40ad-801c-d5ba19184aff#?env%5BQuepasa%5D=W3sia2V5IjoiYmFzZVVybCIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6MH0seyJrZXkiOiJ0b2tlbiIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6MX0seyJrZXkiOiJjaGF0SWQiLCJ2YWx1ZSI6IiIsImVuYWJsZWQiOnRydWUsInR5cGUiOiJkZWZhdWx0Iiwic2Vzc2lvblZhbHVlIjoiIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiIiLCJzZXNzaW9uSW5kZXgiOjJ9LHsia2V5IjoiZmlsZU5hbWUiLCJ2YWx1ZSI6IiIsImVuYWJsZWQiOnRydWUsInR5cGUiOiJkZWZhdWx0Iiwic2Vzc2lvblZhbHVlIjoiIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiIiLCJzZXNzaW9uSW5kZXgiOjN9LHsia2V5IjoidGV4dCIsInZhbHVlIjoiIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiIiLCJjb21wbGV0ZVNlc3Npb25WYWx1ZSI6IiIsInNlc3Npb25JbmRleCI6NH0seyJrZXkiOiJ0cmFja0lkIiwidmFsdWUiOiJwb3N0bWFuIiwiZW5hYmxlZCI6dHJ1ZSwidHlwZSI6ImRlZmF1bHQiLCJzZXNzaW9uVmFsdWUiOiJwb3N0bWFuIiwiY29tcGxldGVTZXNzaW9uVmFsdWUiOiJwb3N0bWFuIiwic2Vzc2lvbkluZGV4Ijo1fV0=)

## 🚀 Quick Start

### Docker Installation (Recommended)
The fastest way to get QuePasa running:

```bash
# Clone the repository
git clone https://github.com/nocodeleaks/quepasa.git
cd quepasa/docker

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start with Docker Compose
docker-compose up -d --build
```

📖 **[Complete Docker Setup Guide](docker/docker.md)**

### Alternative Installation Methods
- **[Local Development Setup](#local-development)**
- **[Manual Installation](#manual-installation)**

## 📋 Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Docker Setup](#docker-installation-recommended)
  - [Local Development](#local-development)
- [Integration Examples](#integration-examples)
- [API Documentation](#api-documentation)
- [Configuration](#configuration)
- [Community & Support](#community--support)
- [Contributing](#contributing)

## ✨ Features

QuePasa provides a simple HTTP API to integrate WhatsApp messaging into your applications:

- 📱 **QR Code Authentication** - Easy WhatsApp Web connection setup
- 💾 **Persistent Sessions** - Account data and keys stored securely  
- 🔗 **HTTP API Endpoints** for:
  - Sending messages (text, media, documents)
  - Receiving messages via webhooks
  - Downloading attachments
  - Managing contacts and groups
  - Group administration
- 🔄 **Webhook Support** - Real-time message notifications
- 📊 **Message History Sync** - Configurable history retrieval
- 🎯 **Advanced Features**:
  - Read receipts
  - Message reactions
  - Broadcast messages
  - Call handling
  - Presence management

## 🐳 Installation

### Docker Installation (Recommended)

The easiest way to deploy QuePasa is using Docker with our pre-configured setup:

1. **Quick Setup**
   ```bash
   git clone https://github.com/nocodeleaks/quepasa.git
   cd quepasa/docker
   cp .env.example .env
   # Edit .env with your configurations
   docker-compose up -d --build
   ```

2. **Access QuePasa**
   - Web Interface: `http://localhost:31000`

📖 **[Detailed Docker Setup Guide](docker/docker.md)** - Complete installation instructions, configuration options, and troubleshooting.

### Local Development

For development or custom installations:

#### Prerequisites
- **Go 1.20+** - [Download here](https://golang.org/dl/)
- **PostgreSQL** - Database for persistent storage
- **Git** - For cloning the repository

#### Build from Source
```bash
# Clone repository
git clone https://github.com/nocodeleaks/quepasa.git
cd quepasa/src

# Install dependencies
go mod download

# Build application
go build -o quepasa main.go

# Run
./quepasa
```

#### API Documentation (Swagger)
QuePasa uses Swagger/OpenAPI for API documentation:

```bash
# Install swag CLI tool (one-time setup)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate/update API documentation
cd src
swag init --output ./swagger

# Or use the provided script
# Windows: double-click generate-swagger.bat
# Or run: .\generate-swagger.bat

# Or use VS Code task: Ctrl+Shift+P → "Tasks: Run Task" → "Generate Swagger Docs"
```

The documentation will be available at `http://localhost:PORT/swagger` (with or without trailing slash) when the application is running.

## 🔗 Integration Examples

### N8N Automation Workflows
Pre-built N8N workflows for common automation scenarios:

- 📁 **[N8N + Chatwoot Integration](extra/n8n+chatwoot/README.md)**
  - Customer service automation
  - Ticket management integration
  - Contact synchronization

- 🤖 **[TypeBot Integration](extra/n8n+chatwoot/README.md)**
  - Chatbot workflows
  - Interactive conversations
  - AI-powered responses

### Chatwoot Help Desk
Complete setup for customer service integration:

- 📁 **[Chatwoot Configuration](extra/chatwoot/README.md)**
  - Help desk setup
  - Nginx configuration
  - Multi-agent support

### API Integration Examples
```bash
# Connect and get QR code
# token could be empty, if empty a new token will be generated
# user is the user that will be manage this connection

curl --location 'localhost:31000/scan' \
  --header 'Accept: application/json' \
  --header 'X-QUEPASA-USER: :user' \
  --header 'X-QUEPASA-TOKEN: :token' \
  --data ''


# Send a message
curl --location 'localhost:31000/send' \
  --header 'Accept: application/json' \
  --header 'X-QUEPASA-TRACKID: :trackid' \
  --header 'X-QUEPASA-CHATID: :chatid' \
  --header 'Content-Type: application/json' \
  --header 'X-QUEPASA-TOKEN: :token' \
  --data '{
      
      "text": "Hello World ! \nHello World !"
  }'

# Set webhook
curl --location 'localhost:31000/webhook' \
  --header 'Accept: application/json' \
  --header 'Content-Type: application/json' \
  --header 'X-QUEPASA-TOKEN: :token' \
  --data '{
      "url": "https://webhook.example.com/webhook/5465465241654",
      "forwardinternal": true,
      "trackid": "custom-track",
      "extra": {
        "clientId": "12345",
        "company": "myCompany",
        "enviroment": "production",
        "version": "1.0"
      }
  }'
```

## 📚 API Documentation

### Core Endpoints
- **Messages**: `/send`
- **Media**: `/send`
- **Groups**: `/groups/`
- **Webhooks**: `/webhook`
- **RabbitMQ**: `/rabbitmq`

### API Versions
- **v4** (Latest) - Recommended for new integrations
- **v3** - Legacy support
- **v2** - Legacy support
- **v1** - Deprecated

📖 **[Complete API Documentation](docs/)** - Detailed endpoint documentation with examples.

## ⚙️ Configuration

### Environment Variables Overview

Key configuration options (see [docker/.env.example](docker/.env.example) for complete list):

```bash
# Basic Configuration
DOMAIN=your-domain.com
MASTERKEY=your-secret-key
ACCOUNTSETUP=true  # Enable for first setup

# Database
DBDRIVER=postgres
DBHOST=postgres
DBDATABASE=quepasa_whatsmeow

# Features
GROUPS=true
READRECEIPTS=true
CALLS=true
WEBSOCKETSSL=false

# Performance
CACHELENGTH=800
HISTORYSYNCDAYS=30
```

📖 **[Environment Variables Reference](src/environment/README.md)** - Complete configuration documentation.

## 🏗️ Architecture

QuePasa is built with:
- **Backend**: Go with [Whatsmeow](https://github.com/tulir/whatsmeow) library
- **Database**: PostgreSQL for data persistence
- **API**: RESTful HTTP endpoints
- **Real-time**: WebSocket support for live updates

## 🤝 Community & Support

### Get Help
- 💬 **Telegram Group**: [QuePasa API](https://t.me/quepasa_api)
- 📢 **Telegram Channel**: [QuePasa Channel](https://t.me/quepasa_channel)
- 🐛 **Issues**: [GitHub Issues](https://github.com/nocodeleaks/quepasa/issues)

### Alternative Projects
Looking for Node.js? Check out [whatsapp-web.js](https://github.com/pedroslopez/whatsapp-web.js) - A more complete Node.js WhatsApp API.

## ⚠️ Important Notices

- **Security**: This application has not been security audited. Use at your own risk.
- **Unofficial**: This is a third-party project, not affiliated with WhatsApp.
- **Terms**: Ensure compliance with WhatsApp's Terms of Service.
- **Rate Limits**: Respect WhatsApp's rate limiting to avoid account suspension.

## 🔄 Development & Contributing

### Project Structure
```
├── src/                    # Go source code
├── docker/                 # Docker configuration
├── extra/                  # Integration examples
│   ├── chatwoot/          # Chatwoot integration
│   ├── n8n+chatwoot/      # N8N workflow examples
│   └── typebot/           # TypeBot integration
├── docs/                   # Documentation
└── helpers/               # Installation helpers
```

### Building
```bash
# Development build
go build -o .dist/quepasa-dev src/main.go

# Production build
go build -ldflags="-s -w" -o .dist/quepasa-prod src/main.go
```

### Environment Variables Reference

For detailed configuration options, see [docker/.env.example](docker/.env.example) and [src/environment/README.md](src/environment/README.md).

#### Core Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `DOMAIN` | Your domain name for the service | `localhost` |
| `WEBAPIPORT` | HTTP server port | `31000` |
| `WEBSOCKETSSL` | Use SSL for WebSocket connections | `false` |
| `MASTERKEY` | Master key for administration | *required* |
| `ACCOUNTSETUP` | Enable account creation setup | `true` |

#### WhatsApp Features
| Variable | Description | Default |
|----------|-------------|---------|
| `GROUPS` | Enable group messaging | `true` |
| `BROADCASTS` | Enable broadcast messages | `false` |
| `READRECEIPTS` | Trigger webhooks for read receipts | `false` |
| `CALLS` | Accept incoming calls | `true` |
| `READUPDATE` | Mark chats as read when sending | `true` |

#### Performance & Caching
| Variable | Description | Default |
|----------|-------------|---------|
| `CACHELENGTH` | Number of messages in cache | `800` |
| `CACHEDAYS` | Days to keep messages in cache | `7` |
| `HISTORYSYNCDAYS` | Days of history to sync on QR scan | `30` |
| `SYNOPSISLENGTH` | Length for message synopsis | `50` |

#### Database Configuration
| Variable | Description | Default |
|----------|-------------|---------|
| `DBDRIVER` | Database driver | `postgres` |
| `DBHOST` | Database host | `localhost` |
| `DBPORT` | Database port | `5432` |
| `DBDATABASE` | Database name | `quepasa_whatsmeow` |
| `DBUSER` | Database user | `quepasa` |
| `DBPASSWORD` | Database password | *required* |

#### Logging & Debug
| Variable | Description | Options |
|----------|-------------|---------|
| `LOGLEVEL` | Application log level | `ERROR`, `WARN`, `INFO`, `DEBUG`, `TRACE` |
| `WHATSMEOW_LOGLEVEL` | WhatsApp library log level | `error`, `warn`, `info`, `debug` |
| `HTTPLOGS` | Log HTTP requests | `true`, `false` |
| `DEBUGREQUESTS` | Debug API requests | `true`, `false` |

#### Media & Conversion
| Variable | Description | Default |
|----------|-------------|---------|
| `CONVERT_PNG_TO_JPG` | Convert PNG to JPG format | `false` |
| `COMPATIBLE_MIME_AS_AUDIO` | Convert audio to OGG/PTT | `true` |
| `REMOVEDIGIT9` | Remove digit 9 from BR numbers | `false` |

#### Regional Settings
| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone | `America/Sao_Paulo` |
| `APP_TITLE` | App title suffix | `QuePasa` |
| `PRESENCE` | Default presence state | `unavailable` |

## 📄 License

[![License GNU AGPL v3.0](https://img.shields.io/badge/License-AGPL%203.0-lightgrey.svg)](https://github.com/nocodeleaks/quepasa/blob/main/LICENSE.md)

QuePasa is free software licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

### What this means:
- ✅ **Free to use** for personal and commercial purposes
- ✅ **Modify and distribute** freely
- ✅ **No warranty** - use at your own risk
- ⚠️ **Copyleft license** - derivative works must also be AGPL-3.0
- ⚠️ **Network use** - if you run a modified version as a service, you must provide source code

## 🔗 References

- [WhatsApp Official](https://whatsapp.com) - Official WhatsApp platform
- [Whatsmeow Library](https://github.com/tulir/whatsmeow) - Go library for WhatsApp Web API
- [Docker Documentation](https://docs.docker.com/) - Container platform documentation
- [PostgreSQL](https://postgresql.org/) - Database system documentation

---

<p align="center">
	<strong>Made with ❤️ by the QuePasa Community</strong><br>
</p>
