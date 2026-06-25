# n8n ↔ QuePasa API v4 Requests Mapping

Document all HTTP requests made from n8n workflows to the QuePasa API (v4 implicit model).

## Overview

The n8n+Chatwoot workflows use the following QuePasa API endpoints (through the n8n-nodes-quepasa custom node or direct HTTP):

- **Base URL**: `{{$json.qphost}}` (e.g., `http://localhost:2000`)
- **Authentication**: 
  - `token` parameter (n8n custom node)
  - `X-QUEPASA-TOKEN` header (HTTP requests)

## Request Summary by Workflow

### 1. QuepasaChatControl.json

**Purpose**: Handle chat commands like invite link retrieval

#### Request 1.1: Send Text Message
- **Node Name**: `Quepasa`
- **Method**: Custom n8n node (`n8n-nodes-quepasa.quepasa`)
- **Resource**: `method`
- **Operation**: `sendtext`
- **Parameters**:
  - `baseUrl`: `{{$json.qphost}}`
  - `token`: `{{$json.qptoken}}`
  - `chatid`: `{{$json.chatid}}`
  - `text`: `{{$json["response"]}}`
- **Purpose**: Send text message to a chat

#### Request 1.2: Get Invite Link
- **Node Name**: `Quepasa Get Invite Link`
- **Method**: Custom n8n node (`n8n-nodes-quepasa.quepasa`)
- **Resource**: `control`
- **Operation**: `invite`
- **Parameters**:
  - `baseUrl`: `{{$json.qphost}}`
  - `token`: `{{$json.qptoken}}`
  - `chatid`: `{{$json.chatid}}`
- **Purpose**: Get group invite link URL

#### Request 1.3: Contact Search (Chatwoot Integration)
- **Node Name**: HTTP Request to Chatwoot
- **HTTP Method**: `POST`
- **URL**: `{{$json.extra.cwhost}}/api/v1/contacts/search`
- **Headers**:
  - `api_access_token`: `{{$json.extra.atoken}}`
- **Purpose**: Search Chatwoot contacts

#### Request 1.4: Send Message to Chatwoot Conversation
- **Node Name**: HTTP Request
- **HTTP Method**: `POST`
- **URL**: `{{$json.extra.cwhost}}/api/v1/accounts/{{$json.extra.account}}/conversations/{{$json.conversation}}/messages`
- **Headers**:
  - `api_access_token`: `{{$json.extra.atoken}}`
- **Body**: JSON with message content
- **Purpose**: Post message to Chatwoot conversation

### 2. PostToChatwoot.json

**Purpose**: Download media from QuePasa and post to Chatwoot

#### Request 2.1: Download Media from QuePasa
- **Node Name**: `Quepasa Download Incoming`
- **Method**: Custom n8n node (`n8n-nodes-quepasa.quepasa`)
- **Resource**: `download`
- **Operation**: `download`
- **Parameters**:
  - `baseUrl`: `{{$json.extra.qphost}}`
  - `token`: `{{$json.extra.qptoken}}`
  - `messageid`: `{{$json.payload.content_attributes?.items?.quepasa?.msgid ?? $json.payload.echo_id}}`
  - `fileName`: `{{$json.attachment.filename}}`
- **Purpose**: Download media attachment from QuePasa message

### 3. ChatwootProfileUpdate.json

**Purpose**: Update contact profile picture from QuePasa

#### Request 3.1: Get Picture Info from QuePasa
- **Node Name**: `Quepasa Picture Info`
- **HTTP Method**: `GET`
- **URL**: `{{$json.qphost}}/picinfo/{{$json.chatid}}`
- **Headers**:
  - `X-QUEPASA-TOKEN`: `{{$json.qptoken}}`
- **Purpose**: Get profile picture information for a contact

#### Request 3.2: Update Contact Picture on Chatwoot
- **Node Name**: HTTP Request (Chatwoot)
- **HTTP Method**: `PUT`
- **URL**: `{{$json.cwhost}}/api/v1/accounts/{{$json.account}}/contacts/{{$json.contactid}}`
- **Headers**:
  - `api_access_token`: `{{$json.utoken}}`
- **Body**: JSON with avatar_url
- **Purpose**: Update contact avatar in Chatwoot

### 4. QuepasaQrcode.json

**Purpose**: Get QR code for new session pairing

#### Request 4.1: Get QR Code
- **Node Name**: (Implicit - likely similar to n8n-nodes-quepasa pattern)
- **Purpose**: Retrieve QR code for device pairing
- **Result**: Success/failure with QR code data

### 5. PostToWebCallBack.json

**Purpose**: Send responses back to external webhook

#### Request 5.1: Send Text via QuePasa
- **Node Name**: `Quepasa`
- **Method**: Custom n8n node (`n8n-nodes-quepasa.quepasa`)
- **Resource**: `method`
- **Operation**: `sendtext`
- **Parameters**:
  - `baseUrl`: `{{$json.extra.qphost}}`
  - `token`: `{{$json.extra.qptoken ?? $json.extra.identifier}}`
  - `chatid`: `{{$json.chatid}}`
  - `text`: Response text
- **Purpose**: Send callback message back

### 6. QuepasaInboxControl_typebot.json

**Purpose**: Handle TypeBot integration with webhook control

#### Request 6.1: Register Webhook
- **Node Name**: Custom n8n node
- **Resource**: `webhook`
- **Parameters**:
  - `baseUrl`: `{{$('EXTRA').item.json.qphost}}`
  - `token`: `{{$('EXTRA').item.json.qptoken}}`
- **Purpose**: Register webhook endpoint for TypeBot events

#### Request 6.2: Direct HTTP to QuePasa
- **Node Name**: HTTP Request
- **HTTP Method**: `POST`
- **URL**: `{{$json.qphost}}/...`
- **Headers**:
  - `X-QUEPASA-TOKEN`: `{{$('EXTRA').item.json.qptoken}}`
- **Purpose**: Direct HTTP communication with QuePasa

## API Endpoint Summary (v4 Implicit)

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `POST /messages/sendtext` | POST | Token | Send text message |
| `GET /control/invite` | GET | Token | Get group invite link |
| `GET /download/:messageid` | GET | Token | Download media |
| `GET /picinfo/:chatid` | GET | X-QUEPASA-TOKEN | Get contact picture info |
| `POST /webhook/register` | POST | Token | Register webhook |

## Notes

- Most requests use the custom n8n node `n8n-nodes-quepasa.quepasa` which abstracts HTTP calls
- Base URL and token are stored in `$json.extra` or `$json` context
- Some flows use both QuePasa API and Chatwoot API in sequence
- Authentication preferably uses session tokens (`X-QUEPASA-TOKEN`)
- Media downloads are implicit in the custom node

## Version Notes

- Implicit API version: v4
- Some endpoints may be versioned or unversioned (canonical)
- Need to verify exact endpoint paths by examining n8n custom node implementation
