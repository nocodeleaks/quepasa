# SIP Proxy Module

## Objective

Explain the intended role of the `sipproxy` module in QuePasa.

This document is intentionally narrow:

- what the module is for
- what it is **not** for today
- how it may be used in a future architecture

## Current Status

The `sipproxy` module exists in the codebase as **future-facing infrastructure**.

It is **not** the current production path for WhatsApp native VoIP support inside QuePasa.

At the moment, it should be understood as preparation for future telephony integration work, not as a completed end-to-end call feature.

## Intended Future Use

The module is meant for scenarios where QuePasa will interact with an external remote calling system.

The two main target flows are:

1. **Start a call from a remote system**
   - a remote platform requests that QuePasa originate a call flow
   - QuePasa receives the request and coordinates the call entry point

2. **Forward a call to a remote system**
   - QuePasa receives or detects a call-related event
   - QuePasa forwards or bridges that flow to an external SIP-based or telephony-based platform

Typical examples of remote systems:

- SIP servers
- PBX platforms
- call center software
- telecom gateways
- custom voice automation platforms

## Important Non-Goal for Now

The `sipproxy` module should **not** be interpreted as the current solution for:

- native WhatsApp VoIP media handling
- WhatsApp call relay/TURN handling
- SRTP/Opus media processing for WhatsApp calls
- complete inbound/outbound WhatsApp call support in production

In other words:

- `sipproxy` is **not** the current answer to QuePasa WhatsApp VoIP support
- `sipproxy` is **not** the active bridge used by the current WhatsApp call handlers
- `sipproxy` is **not** proof that QuePasa already supports full WhatsApp calling end-to-end

## Relationship to Current WhatsApp Call Events

QuePasa can already observe some WhatsApp call-related events through the WhatsApp integration layer.

However, this does **not** mean the `sipproxy` module is already wired as the production call transport path.

For now, keep these concerns separate:

- **WhatsApp call event observation**
- **future SIP/telephony forwarding or origination**

The existence of both areas in the repository does not imply that the complete call pipeline is already implemented.

## Architectural Intent

The long-term idea is to keep `sipproxy` as a **boundary module** between QuePasa and external voice systems.

That boundary may eventually be used to:

- receive call instructions from remote services
- normalize telephony signaling expectations
- forward QuePasa call flows into external systems
- support controlled integration with enterprise voice platforms

This keeps external telephony concerns isolated from the core WhatsApp session and messaging flows.

## Practical Reading Rule

When reading the codebase, interpret `src/sipproxy/` as:

- planned infrastructure
- exploratory/foundation work
- future integration surface

Do **not** interpret it as:

- finished product behavior
- current official VoIP implementation for WhatsApp
- active production bridge for all call events

## Summary

The `sipproxy` module exists for a **future integration phase**.

Its purpose is to support cases where QuePasa will either:

- initiate a call starting from a remote system
- or forward a call to a remote system

That future direction is valid and intentional, but it is **not the current VoIP solution** for QuePasa today.
