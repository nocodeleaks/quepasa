package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://github.com/nocodeleaks/quepasa",
        "contact": {
            "name": "QuePasa Support",
            "url": "https://github.com/nocodeleaks/quepasa",
            "email": "support@quepasa.io"
        },
        "license": {
            "name": "MIT",
            "url": "https://github.com/nocodeleaks/quepasa/blob/main/LICENSE.md"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/account": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, update, or manage user accounts (master access required)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Application"
                ],
                "summary": "Manage user accounts",
                "parameters": [
                    {
                        "description": "Account request",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "password": {
                                    "type": "string"
                                },
                                "username": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/chat/presence": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Controls typing indicators and chat presence in WhatsApp conversations",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Chat"
                ],
                "summary": "Control chat presence",
                "parameters": [
                    {
                        "description": "Chat presence request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_main_api.ChatPresenceRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/command": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Execute control commands for the bot server (start, stop, restart, status)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Bot"
                ],
                "summary": "Execute bot commands",
                "parameters": [
                    {
                        "enum": [
                            "start",
                            "stop",
                            "restart",
                            "status"
                        ],
                        "type": "string",
                        "description": "Command action",
                        "name": "action",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/contacts": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves a list of all WhatsApp contacts",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Contacts"
                ],
                "summary": "Get contacts",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpContactsResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/download/{messageid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Downloads media files (images, videos, documents) from WhatsApp messages",
                "produces": [
                    "application/octet-stream"
                ],
                "tags": [
                    "Download"
                ],
                "summary": "Download media",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Message ID (path parameter)",
                        "name": "messageid",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (query parameter)",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (query parameter alternate)",
                        "name": "messageid",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Use cached content",
                        "name": "cache",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (header parameter)",
                        "name": "X-QUEPASA-MESSAGEID",
                        "in": "header"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Media file",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/edit": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Edits the content of an existing message by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Message"
                ],
                "summary": "Edit message",
                "parameters": [
                    {
                        "description": "Message edit request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "content": {
                                    "type": "string"
                                },
                                "messageId": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/create": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Creates a new WhatsApp group with specified title and participants",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Create a new group",
                "parameters": [
                    {
                        "description": "Group creation request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "participants": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                },
                                "title": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSingleGroupResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/description": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Updates the topic/description of a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Set group topic",
                "parameters": [
                    {
                        "description": "Group topic update request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "group_jid": {
                                    "type": "string"
                                },
                                "topic": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSingleGroupResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/get": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves detailed information about a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Get group information",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Group ID",
                        "name": "groupId",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSingleGroupResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/getall": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves a list of all WhatsApp groups that the bot is currently a member of",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Get all groups",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpGroupsResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/leave": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Leave a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Leave group",
                "parameters": [
                    {
                        "description": "Leave group request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "chatId": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/name": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Updates the name of a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Set group name",
                "parameters": [
                    {
                        "description": "Group name update request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "group_jid": {
                                    "type": "string"
                                },
                                "name": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSingleGroupResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/participants": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Add, remove, promote, or demote participants in a WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Update group participants",
                "parameters": [
                    {
                        "description": "Participants update request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "action": {
                                    "type": "string"
                                },
                                "group_jid": {
                                    "type": "string"
                                },
                                "participants": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpParticipantResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/photo": {
            "put": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Updates or removes the photo of a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Set/Remove group photo",
                "parameters": [
                    {
                        "description": "Group photo update request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "group_jid": {
                                    "type": "string"
                                },
                                "remove_img": {
                                    "type": "boolean"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSingleGroupResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/groups/requests": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get, approve, or reject join requests for WhatsApp groups",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Handle group join requests",
                "parameters": [
                    {
                        "description": "Membership request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "action": {
                                    "type": "string"
                                },
                                "group_jid": {
                                    "type": "string"
                                },
                                "participants": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    },
                    {
                        "type": "string",
                        "description": "Group JID (for GET requests)",
                        "name": "group_jid",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpRequestResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get, approve, or reject join requests for WhatsApp groups",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Handle group join requests",
                "parameters": [
                    {
                        "description": "Membership request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "action": {
                                    "type": "string"
                                },
                                "group_jid": {
                                    "type": "string"
                                },
                                "participants": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    },
                    {
                        "type": "string",
                        "description": "Group JID (for GET requests)",
                        "name": "group_jid",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpRequestResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/health": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Basic health check endpoint to verify if the application is running",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Health check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_api_models.HealthResponse"
                        }
                    }
                }
            }
        },
        "/healthapi": {
            "get": {
                "description": "Provides detailed health information for WhatsApp servers with authentication support",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Health"
                ],
                "summary": "Detailed health check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_api_models.HealthResponse"
                        }
                    }
                }
            }
        },
        "/info": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get, update, or delete bot/server information and settings",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Information"
                ],
                "summary": "Manage bot information",
                "parameters": [
                    {
                        "description": "Settings update (for PATCH)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "settings": {
                                    "type": "object"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpInfoResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get, update, or delete bot/server information and settings",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Information"
                ],
                "summary": "Manage bot information",
                "parameters": [
                    {
                        "description": "Settings update (for PATCH)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "settings": {
                                    "type": "object"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpInfoResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get, update, or delete bot/server information and settings",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Information"
                ],
                "summary": "Manage bot information",
                "parameters": [
                    {
                        "description": "Settings update (for PATCH)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "settings": {
                                    "type": "object"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpInfoResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/invite": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates an invite link for a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Generate group invite link",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Chat ID (query parameter)",
                        "name": "chatid",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpInviteResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/invite/{chatid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates an invite link for a specific WhatsApp group",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Generate group invite link",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Chat ID (path parameter)",
                        "name": "chatid",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (query parameter)",
                        "name": "chatid",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpInviteResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/isonwhatsapp": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Checks if provided phone numbers are registered on WhatsApp",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Contacts"
                ],
                "summary": "Check WhatsApp registration",
                "parameters": [
                    {
                        "description": "Phone numbers to check",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "phones": {
                                    "type": "array",
                                    "items": {
                                        "type": "string"
                                    }
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpIsOnWhatsappResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/message/{messageid}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves a specific message by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Message"
                ],
                "summary": "Get message",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Message ID",
                        "name": "messageid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpMessageResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Revokes or deletes a specific message by its ID",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Message"
                ],
                "summary": "Revoke message",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Message ID",
                        "name": "messageid",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpMessageResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/paircode": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates a pairing code for WhatsApp authentication using phone number",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Connection"
                ],
                "summary": "Generate pairing code",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Phone number for pairing",
                        "name": "phone",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/rabbitmq": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, get, or delete RabbitMQ configurations for message queueing",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "RabbitMQ"
                ],
                "summary": "Manage RabbitMQ configurations",
                "parameters": [
                    {
                        "description": "RabbitMQ config (for POST)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "connection_string": {
                                    "type": "string"
                                },
                                "exchange": {
                                    "type": "string"
                                },
                                "routing_key": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    {
                        "type": "string",
                        "description": "Connection string (for DELETE)",
                        "name": "connection_string",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpRabbitMQResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, get, or delete RabbitMQ configurations for message queueing",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "RabbitMQ"
                ],
                "summary": "Manage RabbitMQ configurations",
                "parameters": [
                    {
                        "description": "RabbitMQ config (for POST)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "connection_string": {
                                    "type": "string"
                                },
                                "exchange": {
                                    "type": "string"
                                },
                                "routing_key": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    {
                        "type": "string",
                        "description": "Connection string (for DELETE)",
                        "name": "connection_string",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpRabbitMQResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, get, or delete RabbitMQ configurations for message queueing",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "RabbitMQ"
                ],
                "summary": "Manage RabbitMQ configurations",
                "parameters": [
                    {
                        "description": "RabbitMQ config (for POST)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "connection_string": {
                                    "type": "string"
                                },
                                "exchange": {
                                    "type": "string"
                                },
                                "routing_key": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    {
                        "type": "string",
                        "description": "Connection string (for DELETE)",
                        "name": "connection_string",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpRabbitMQResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/scan": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Generates a QR code for WhatsApp Web authentication scanning",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Connection"
                ],
                "summary": "Generate QR code",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/send": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send text, file, poll, or any type of message to WhatsApp. Supports polls with JSON format in text field. For /sendencoded route, use 'content' field with base64 encoded file data.\n\n**Poll Example:**\n` + "`" + `` + "`" + `` + "`" + `json\n{\n\"chatId\": \"5511999999999@s.whatsapp.net\",\n\"poll\": {\n\"question\": \"What programming languages do you know?\",\n\"options\": [\"JavaScript\", \"Python\", \"Go\", \"Java\", \"C#\", \"Ruby\"],\n\"selections\": 3\n}\n}\n` + "`" + `` + "`" + `` + "`" + `",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Send"
                ],
                "summary": "Send any type of message (text, file, poll, base64 content)",
                "parameters": [
                    {
                        "description": "Send request body (use 'content' field for base64 encoded files, 'url' for file URL, 'poll' for poll JSON)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "chatId": {
                                    "type": "string"
                                },
                                "content": {
                                    "type": "string"
                                },
                                "fileName": {
                                    "type": "string"
                                },
                                "poll": {
                                    "type": "object",
                                    "properties": {
                                        "options": {
                                            "type": "array",
                                            "items": {
                                                "type": "string"
                                            }
                                        },
                                        "question": {
                                            "type": "string"
                                        },
                                        "selections": {
                                            "type": "integer"
                                        }
                                    }
                                },
                                "text": {
                                    "type": "string"
                                },
                                "url": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/sendbinary": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send any binary file (audio, video, image, document) using raw binary data in request body. Supports multiple parameter methods (path, query, headers).",
                "consumes": [
                    "application/octet-stream",
                    "audio/mpeg",
                    "video/mp4",
                    "image/jpeg",
                    "image/png",
                    "application/pdf"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Send"
                ],
                "summary": "Send binary file directly from request body",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Chat ID (query parameter)",
                        "name": "chatId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "File name (query parameter)",
                        "name": "filename",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (query parameter)",
                        "name": "text",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID to reply to",
                        "name": "inreply",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (header parameter)",
                        "name": "X-QUEPASA-CHATID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "File name (header parameter)",
                        "name": "X-QUEPASA-FILENAME",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (header parameter)",
                        "name": "X-QUEPASA-TEXT",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Track ID for message tracking",
                        "name": "X-QUEPASA-TRACKID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "MIME type of the binary file (e.g., audio/mpeg, video/mp4, image/jpeg)",
                        "name": "Content-Type",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/sendbinary/{chatid}": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send any binary file (audio, video, image, document) using raw binary data in request body. Supports multiple parameter methods (path, query, headers).",
                "consumes": [
                    "application/octet-stream",
                    "audio/mpeg",
                    "video/mp4",
                    "image/jpeg",
                    "image/png",
                    "application/pdf"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Send"
                ],
                "summary": "Send binary file directly from request body",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Chat ID (path parameter)",
                        "name": "chatid",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (query parameter)",
                        "name": "chatId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "File name (query parameter)",
                        "name": "filename",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (query parameter)",
                        "name": "text",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID to reply to",
                        "name": "inreply",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (header parameter)",
                        "name": "X-QUEPASA-CHATID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "File name (header parameter)",
                        "name": "X-QUEPASA-FILENAME",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (header parameter)",
                        "name": "X-QUEPASA-TEXT",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Track ID for message tracking",
                        "name": "X-QUEPASA-TRACKID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "MIME type of the binary file (e.g., audio/mpeg, video/mp4, image/jpeg)",
                        "name": "Content-Type",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/sendbinary/{chatid}/{filename}": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send any binary file (audio, video, image, document) using raw binary data in request body. Supports multiple parameter methods (path, query, headers).",
                "consumes": [
                    "application/octet-stream",
                    "audio/mpeg",
                    "video/mp4",
                    "image/jpeg",
                    "image/png",
                    "application/pdf"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Send"
                ],
                "summary": "Send binary file directly from request body",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Chat ID (path parameter)",
                        "name": "chatid",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "File name (path parameter)",
                        "name": "filename",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (query parameter)",
                        "name": "chatId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "File name (query parameter)",
                        "name": "filename",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (query parameter)",
                        "name": "text",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID to reply to",
                        "name": "inreply",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (header parameter)",
                        "name": "X-QUEPASA-CHATID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "File name (header parameter)",
                        "name": "X-QUEPASA-FILENAME",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (header parameter)",
                        "name": "X-QUEPASA-TEXT",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Track ID for message tracking",
                        "name": "X-QUEPASA-TRACKID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "MIME type of the binary file (e.g., audio/mpeg, video/mp4, image/jpeg)",
                        "name": "Content-Type",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/sendbinary/{chatid}/{filename}/{text}": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send any binary file (audio, video, image, document) using raw binary data in request body. Supports multiple parameter methods (path, query, headers).",
                "consumes": [
                    "application/octet-stream",
                    "audio/mpeg",
                    "video/mp4",
                    "image/jpeg",
                    "image/png",
                    "application/pdf"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Send"
                ],
                "summary": "Send binary file directly from request body",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Chat ID (path parameter)",
                        "name": "chatid",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "File name (path parameter)",
                        "name": "filename",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (path parameter)",
                        "name": "text",
                        "in": "path"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (query parameter)",
                        "name": "chatId",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "File name (query parameter)",
                        "name": "filename",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (query parameter)",
                        "name": "text",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID to reply to",
                        "name": "inreply",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Chat ID (header parameter)",
                        "name": "X-QUEPASA-CHATID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "File name (header parameter)",
                        "name": "X-QUEPASA-FILENAME",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Caption text for images (header parameter)",
                        "name": "X-QUEPASA-TEXT",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "Track ID for message tracking",
                        "name": "X-QUEPASA-TRACKID",
                        "in": "header"
                    },
                    {
                        "type": "string",
                        "description": "MIME type of the binary file (e.g., audio/mpeg, video/mp4, image/jpeg)",
                        "name": "Content-Type",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/sendencoded": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send text, file, poll, or any type of message to WhatsApp. Supports polls with JSON format in text field. For /sendencoded route, use 'content' field with base64 encoded file data.\n\n**Poll Example:**\n` + "`" + `` + "`" + `` + "`" + `json\n{\n\"chatId\": \"5511999999999@s.whatsapp.net\",\n\"poll\": {\n\"question\": \"What programming languages do you know?\",\n\"options\": [\"JavaScript\", \"Python\", \"Go\", \"Java\", \"C#\", \"Ruby\"],\n\"selections\": 3\n}\n}\n` + "`" + `` + "`" + `` + "`" + `",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Send"
                ],
                "summary": "Send any type of message (text, file, poll, base64 content)",
                "parameters": [
                    {
                        "description": "Send request body (use 'content' field for base64 encoded files, 'url' for file URL, 'poll' for poll JSON)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "chatId": {
                                    "type": "string"
                                },
                                "content": {
                                    "type": "string"
                                },
                                "fileName": {
                                    "type": "string"
                                },
                                "poll": {
                                    "type": "object",
                                    "properties": {
                                        "options": {
                                            "type": "array",
                                            "items": {
                                                "type": "string"
                                            }
                                        },
                                        "question": {
                                            "type": "string"
                                        },
                                        "selections": {
                                            "type": "integer"
                                        }
                                    }
                                },
                                "text": {
                                    "type": "string"
                                },
                                "url": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/spam": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Send messages using any available server (spam/broadcast functionality)",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Application"
                ],
                "summary": "Send spam messages",
                "parameters": [
                    {
                        "description": "Spam message request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "chatId": {
                                    "type": "string"
                                },
                                "text": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    },
                    "423": {
                        "description": "No server available",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponse"
                        }
                    }
                }
            }
        },
        "/useridentifier": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves the Local Identifier (LID) for a given phone number",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Contacts"
                ],
                "summary": "Get user identifier (LID)",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Phone number",
                        "name": "phone",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Local identifier",
                        "name": "lid",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_main_api.LIDResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/userinfo": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves detailed information for WhatsApp users by their JIDs",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Contacts"
                ],
                "summary": "Get user information",
                "parameters": [
                    {
                        "description": "User info request with JIDs",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_main_api.UserInfoRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_main_api.UserInfoResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/v2/bot/{token}/receive": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves pending messages from WhatsApp with optional timestamp filtering",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Message"
                ],
                "summary": "Receive messages",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Timestamp filter for messages",
                        "name": "timestamp",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpReceiveResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/v3/bot/{token}/download": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Downloads media files (images, videos, documents) from WhatsApp messages",
                "produces": [
                    "application/octet-stream"
                ],
                "tags": [
                    "Download"
                ],
                "summary": "Download media",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Message ID (query parameter)",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (query parameter alternate)",
                        "name": "messageid",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Use cached content",
                        "name": "cache",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (header parameter)",
                        "name": "X-QUEPASA-MESSAGEID",
                        "in": "header"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Media file",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/v3/bot/{token}/download/{messageId}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Downloads media files (images, videos, documents) from WhatsApp messages",
                "produces": [
                    "application/octet-stream"
                ],
                "tags": [
                    "Download"
                ],
                "summary": "Download media",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Message ID (query parameter)",
                        "name": "id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (query parameter alternate)",
                        "name": "messageid",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Use cached content",
                        "name": "cache",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Message ID (header parameter)",
                        "name": "X-QUEPASA-MESSAGEID",
                        "in": "header"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Media file",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/v3/bot/{token}/receive": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Retrieves pending messages from WhatsApp with optional timestamp filtering",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Message"
                ],
                "summary": "Receive messages",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Timestamp filter for messages",
                        "name": "timestamp",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpReceiveResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        },
        "/webhook": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, get, or delete webhook configurations for event notifications",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Webhooks"
                ],
                "summary": "Manage webhook configurations",
                "parameters": [
                    {
                        "description": "Webhook config (for POST/DELETE)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "bearer_token": {
                                    "type": "string"
                                },
                                "failure_bearer_token": {
                                    "type": "string"
                                },
                                "failure_method": {
                                    "type": "string"
                                },
                                "failure_url": {
                                    "type": "string"
                                },
                                "method": {
                                    "type": "string"
                                },
                                "url": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpWebhookResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, get, or delete webhook configurations for event notifications",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Webhooks"
                ],
                "summary": "Manage webhook configurations",
                "parameters": [
                    {
                        "description": "Webhook config (for POST/DELETE)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "bearer_token": {
                                    "type": "string"
                                },
                                "failure_bearer_token": {
                                    "type": "string"
                                },
                                "failure_method": {
                                    "type": "string"
                                },
                                "failure_url": {
                                    "type": "string"
                                },
                                "method": {
                                    "type": "string"
                                },
                                "url": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpWebhookResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create, get, or delete webhook configurations for event notifications",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Webhooks"
                ],
                "summary": "Manage webhook configurations",
                "parameters": [
                    {
                        "description": "Webhook config (for POST/DELETE)",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "bearer_token": {
                                    "type": "string"
                                },
                                "failure_bearer_token": {
                                    "type": "string"
                                },
                                "failure_method": {
                                    "type": "string"
                                },
                                "failure_url": {
                                    "type": "string"
                                },
                                "method": {
                                    "type": "string"
                                },
                                "url": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpWebhookResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "github_com_nocodeleaks_quepasa_api_models.HealthResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "items": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpHealthResponseItem"
                    }
                },
                "stats": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_api_models.HealthStats"
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_api_models.HealthStats": {
            "type": "object",
            "properties": {
                "healthy": {
                    "type": "integer"
                },
                "percentage": {
                    "type": "number"
                },
                "total": {
                    "type": "integer"
                },
                "unhealthy": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_main_api.ChatPresenceRequest": {
            "type": "object",
            "properties": {
                "chatid": {
                    "description": "Required: Chat to show typing in",
                    "type": "string"
                },
                "duration": {
                    "description": "Optional: Auto-stop after duration (ms)",
                    "type": "integer"
                },
                "type": {
                    "description": "Text or audio",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappChatPresenceType"
                        }
                    ]
                }
            }
        },
        "github_com_nocodeleaks_quepasa_main_api.LIDResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "lid": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_main_api.UserInfoRequest": {
            "type": "object",
            "properties": {
                "jids": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "github_com_nocodeleaks_quepasa_main_api.UserInfoResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                },
                "userinfos": {
                    "type": "array",
                    "items": {}
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpContactsResponse": {
            "type": "object",
            "properties": {
                "contacts": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappChat"
                    }
                },
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpDispatching": {
            "type": "object",
            "properties": {
                "broadcasts": {
                    "description": "should handle broadcast messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "calls": {
                    "description": "should handle calls",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "connection_string": {
                    "description": "destination URL (webhook) or connection string (rabbitmq)",
                    "type": "string"
                },
                "extra": {
                    "description": "extra info to append on payload"
                },
                "failure": {
                    "description": "first failure timestamp",
                    "type": "string"
                },
                "forwardinternal": {
                    "description": "forward internal msg from api",
                    "type": "boolean"
                },
                "groups": {
                    "description": "should handle groups messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "readreceipts": {
                    "description": "should emit read receipts",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "success": {
                    "description": "last success timestamp",
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "trackid": {
                    "description": "identifier of remote system to avoid loop",
                    "type": "string"
                },
                "type": {
                    "description": "webhook or rabbitmq",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpGroupsResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "groups": {
                    "type": "array",
                    "items": {}
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpHealthResponseItem": {
            "type": "object",
            "properties": {
                "status": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappConnectionState"
                },
                "token": {
                    "description": "Public token",
                    "type": "string"
                },
                "wid": {
                    "description": "Whatsapp session id",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpInfoResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "server": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpWhatsappServer"
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpInviteResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "url": {
                    "description": "invite public link",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpIsOnWhatsappResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "registered": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpMessageResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "message": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessage"
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpParticipantResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "participants": {
                    "type": "array",
                    "items": {}
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpRabbitMQConfig": {
            "type": "object",
            "properties": {
                "broadcasts": {
                    "description": "should handle broadcast messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "calls": {
                    "description": "should handle calls",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "connection_string": {
                    "description": "RabbitMQ Connection Settings",
                    "type": "string"
                },
                "exchange_name": {
                    "description": "RabbitMQ exchange name for routing",
                    "type": "string"
                },
                "extra": {
                    "description": "extra info to append on payload"
                },
                "failure": {
                    "description": "Status Tracking",
                    "type": "string"
                },
                "forwardinternal": {
                    "description": "Configuration Options",
                    "type": "boolean"
                },
                "groups": {
                    "description": "should handle groups messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "queue_history": {
                    "description": "RabbitMQ history queue name (optional)",
                    "type": "string"
                },
                "readreceipts": {
                    "description": "should emit read receipts",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "routing_key": {
                    "description": "RabbitMQ routing key for exchange routing",
                    "type": "string"
                },
                "success": {
                    "description": "last success timestamp",
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "trackid": {
                    "description": "identifier of remote system to avoid loop",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpRabbitMQResponse": {
            "type": "object",
            "properties": {
                "affected": {
                    "description": "items affected",
                    "type": "integer"
                },
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "rabbitmq": {
                    "description": "current rabbitmq items",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpRabbitMQConfig"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpReceiveResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "messages": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessage"
                    }
                },
                "server": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpServer"
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpRequestResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "requests": {
                    "type": "array",
                    "items": {}
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpSendResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "message": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpSendResponseMessage"
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpSendResponseMessage": {
            "type": "object",
            "properties": {
                "chatId": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "trackId": {
                    "type": "string"
                },
                "wid": {
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpServer": {
            "type": "object",
            "properties": {
                "broadcasts": {
                    "description": "should handle broadcast messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "calls": {
                    "description": "should handle calls",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "devel": {
                    "type": "boolean"
                },
                "groups": {
                    "description": "should handle groups messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "readreceipts": {
                    "description": "should emit read receipts",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "timestamp": {
                    "type": "string"
                },
                "token": {
                    "description": "Public token",
                    "type": "string",
                    "maxLength": 100
                },
                "user": {
                    "type": "string",
                    "maxLength": 36
                },
                "verified": {
                    "type": "boolean"
                },
                "wid": {
                    "description": "Whatsapp session id",
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpSingleGroupResponse": {
            "type": "object",
            "properties": {
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "groupinfo": {},
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpWebhook": {
            "type": "object",
            "properties": {
                "broadcasts": {
                    "description": "should handle broadcast messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "calls": {
                    "description": "should handle calls",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "extra": {
                    "description": "extra info to append on payload"
                },
                "failure": {
                    "description": "first failure timestamp",
                    "type": "string"
                },
                "forwardinternal": {
                    "description": "forward internal msg from api",
                    "type": "boolean"
                },
                "groups": {
                    "description": "should handle groups messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "readreceipts": {
                    "description": "should emit read receipts",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "success": {
                    "description": "last success timestamp",
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "trackid": {
                    "description": "identifier of remote system to avoid loop",
                    "type": "string"
                },
                "url": {
                    "description": "destination",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpWebhookResponse": {
            "type": "object",
            "properties": {
                "affected": {
                    "description": "items affected",
                    "type": "integer"
                },
                "debug": {
                    "description": "Extra interface{} ` + "`" + `json:\"extra,omitempty\"` + "`" + `",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "status": {
                    "type": "string"
                },
                "success": {
                    "type": "boolean"
                },
                "webhooks": {
                    "description": "current items",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpWebhook"
                    }
                }
            }
        },
        "github_com_nocodeleaks_quepasa_models.QpWhatsappServer": {
            "type": "object",
            "properties": {
                "broadcasts": {
                    "description": "should handle broadcast messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "calls": {
                    "description": "should handle calls",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "devel": {
                    "type": "boolean"
                },
                "dispatching": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_nocodeleaks_quepasa_models.QpDispatching"
                    }
                },
                "groups": {
                    "description": "should handle groups messages",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "readreceipts": {
                    "description": "should emit read receipts",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean"
                        }
                    ]
                },
                "reconnect": {
                    "description": "should auto reconnect, false for qrcode scanner",
                    "type": "boolean"
                },
                "starttime": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "token": {
                    "description": "Public token",
                    "type": "string",
                    "maxLength": 100
                },
                "user": {
                    "type": "string",
                    "maxLength": 36
                },
                "verified": {
                    "type": "boolean"
                },
                "wid": {
                    "description": "Whatsapp session id",
                    "type": "string",
                    "maxLength": 255
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappAttachment": {
            "type": "object",
            "properties": {
                "checksum": {
                    "description": "Checksum for the message, used to verify integrity\nand avoid duplicates",
                    "type": "string"
                },
                "filelength": {
                    "description": "important to navigate throw content, declared file length",
                    "type": "integer"
                },
                "filename": {
                    "description": "document",
                    "type": "string"
                },
                "latitude": {
                    "description": "location msgs",
                    "type": "number"
                },
                "longitude": {
                    "type": "number"
                },
                "mime": {
                    "type": "string"
                },
                "seconds": {
                    "description": "audio/video",
                    "type": "integer"
                },
                "sequence": {
                    "description": "live location",
                    "type": "integer"
                },
                "thumbnail": {
                    "description": "small image representing something in this message, MIME: image/jpeg",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageThumbnail"
                        }
                    ]
                },
                "url": {
                    "description": "Public access url helper content",
                    "type": "string"
                },
                "waveform": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappBoolean": {
            "type": "integer",
            "enum": [
                -1,
                0,
                1
            ],
            "x-enum-varnames": [
                "FalseBooleanType",
                "UnSetBooleanType",
                "TrueBooleanType"
            ]
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappChat": {
            "type": "object",
            "properties": {
                "id": {
                    "description": "(Identifier) whatsapp contact id, based on phone number or timestamp",
                    "type": "string"
                },
                "lid": {
                    "description": "(Local Identifier) new whatsapp unique contact id",
                    "type": "string"
                },
                "phone": {
                    "description": "phone number in E164 format",
                    "type": "string"
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappChatPresenceType": {
            "type": "integer",
            "enum": [
                0,
                1,
                2
            ],
            "x-enum-varnames": [
                "WhatsappChatPresenceTypePaused",
                "WhatsappChatPresenceTypeText",
                "WhatsappChatPresenceTypeAudio"
            ]
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappConnectionState": {
            "type": "integer",
            "enum": [
                0,
                1,
                2,
                3,
                4,
                5,
                6,
                7,
                8,
                9,
                10,
                11,
                12,
                13,
                14
            ],
            "x-enum-varnames": [
                "Unknown",
                "UnPrepared",
                "UnVerified",
                "Starting",
                "Connecting",
                "Stopping",
                "Stopped",
                "Restarting",
                "Reconnecting",
                "Connected",
                "Fetching",
                "Ready",
                "Halting",
                "Disconnected",
                "Failed"
            ]
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessage": {
            "type": "object",
            "properties": {
                "ads": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageAds"
                },
                "attachment": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappAttachment"
                },
                "chat": {
                    "description": "Em qual chat (grupo ou direct) essa msg foi postada, para onde devemos responder",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappChat"
                        }
                    ]
                },
                "debug": {
                    "description": "Debug information for debug events",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageDebug"
                        }
                    ]
                },
                "edited": {
                    "description": "Edited message",
                    "type": "boolean"
                },
                "forwardingscore": {
                    "description": "How many times this message was forwarded",
                    "type": "integer"
                },
                "fromhistory": {
                    "description": "Generated from history sync",
                    "type": "boolean"
                },
                "frominternal": {
                    "description": "Sended via api",
                    "type": "boolean"
                },
                "fromme": {
                    "description": "Do i send that ?\nFrom any connected device and api",
                    "type": "boolean"
                },
                "id": {
                    "description": "Upper text msg id",
                    "type": "string"
                },
                "info": {
                    "description": "Extra information for custom messages"
                },
                "inreply": {
                    "description": "Msg in reply of another ? Message ID",
                    "type": "string"
                },
                "participant": {
                    "description": "If this message was posted on a Group, Who posted it !",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappChat"
                        }
                    ]
                },
                "poll": {
                    "description": "Poll if exists",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappPoll"
                        }
                    ]
                },
                "status": {
                    "description": "Delivered, Read, Imported statuses",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageStatus"
                        }
                    ]
                },
                "synopsis": {
                    "description": "Msg in reply preview",
                    "type": "string"
                },
                "text": {
                    "description": "Message text if exists",
                    "type": "string"
                },
                "timestamp": {
                    "type": "string"
                },
                "trackid": {
                    "description": "Optional id of the system that send that message",
                    "type": "string"
                },
                "type": {
                    "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageType"
                },
                "url": {
                    "description": "Url if exists",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageUrl"
                        }
                    ]
                },
                "wid": {
                    "description": "WhatsApp ID of the sender",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageAds": {
            "type": "object",
            "properties": {
                "app": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "sourceid": {
                    "type": "string"
                },
                "sourceurl": {
                    "type": "string"
                },
                "thumbnail": {
                    "description": "small image representing something in this message, MIME: image/jpeg",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageThumbnail"
                        }
                    ]
                },
                "title": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageDebug": {
            "type": "object",
            "properties": {
                "event": {
                    "type": "string"
                },
                "info": {
                    "description": "Additional information about the event"
                },
                "reason": {
                    "description": "Reason for the debug event",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageStatus": {
            "type": "string",
            "enum": [
                "",
                "error",
                "imported",
                "delivered",
                "read"
            ],
            "x-enum-varnames": [
                "WhatsappMessageStatusUnknown",
                "WhatsappMessageStatusError",
                "WhatsappMessageStatusImported",
                "WhatsappMessageStatusDelivered",
                "WhatsappMessageStatusRead"
            ]
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageThumbnail": {
            "type": "object",
            "properties": {
                "data": {
                    "description": "base64 data",
                    "type": "string"
                },
                "mime": {
                    "description": "content mime type",
                    "type": "string"
                },
                "urlprefix": {
                    "description": "trick for '\u003cimg src=' urls prefix",
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageType": {
            "type": "integer",
            "enum": [
                0,
                1,
                2,
                3,
                4,
                5,
                6,
                7,
                8,
                9,
                10,
                11,
                12
            ],
            "x-enum-varnames": [
                "UnhandledMessageType",
                "ImageMessageType",
                "DocumentMessageType",
                "AudioMessageType",
                "VideoMessageType",
                "TextMessageType",
                "LocationMessageType",
                "ContactMessageType",
                "CallMessageType",
                "SystemMessageType",
                "GroupMessageType",
                "RevokeMessageType",
                "PollMessageType"
            ]
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageUrl": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "reference": {
                    "type": "string"
                },
                "thumbnail": {
                    "description": "small image representing something in this message, MIME: image/jpeg",
                    "allOf": [
                        {
                            "$ref": "#/definitions/github_com_nocodeleaks_quepasa_whatsapp.WhatsappMessageThumbnail"
                        }
                    ]
                },
                "title": {
                    "type": "string"
                }
            }
        },
        "github_com_nocodeleaks_quepasa_whatsapp.WhatsappPoll": {
            "type": "object",
            "properties": {
                "options": {
                    "description": "Required: Array of poll options",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "question": {
                    "description": "Required: Poll question/title",
                    "type": "string"
                },
                "selections": {
                    "description": "Optional: Maximum number of options a user can select (default: 1)",
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "X-QUEPASA-TOKEN",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "4.0.0",
	Host:             "localhost:31000",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "QuePasa WhatsApp API",
	Description:      "QuePasa is a Go-based WhatsApp bot platform that exposes HTTP APIs for WhatsApp messaging integration",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
