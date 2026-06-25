package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const connectionDiagnosticMetadataKey = "connection_diagnostic"

type QpConnectionDiagnostic struct {
	Code              string     `json:"code,omitempty"`
	Message           string     `json:"message,omitempty"`
	SuggestedAction   string     `json:"suggested_action,omitempty"`
	OccurredAt        *time.Time `json:"occurred_at,omitempty"`
	RequiresReauth    bool       `json:"requires_reauth,omitempty"`
	DisconnectCause   string     `json:"disconnect_cause,omitempty"`
	DisconnectDetails string     `json:"disconnect_details,omitempty"`
	LogoutReason      string     `json:"logout_reason,omitempty"`
}

func deriveLogoutIssueCode(reason string) string {
	normalized := strings.ToLower(strings.TrimSpace(reason))

	switch {
	case strings.Contains(normalized, "another device"):
		return "logged_out_another_device"
	case strings.Contains(normalized, "banned"):
		return "logged_out_banned"
	case strings.Contains(normalized, "device removed"):
		return "logged_out_device_removed"
	case normalized == "":
		return "logged_out"
	default:
		return "logged_out"
	}
}

func buildLogoutIssueMessage(reason string) string {
	switch deriveLogoutIssueCode(reason) {
	case "logged_out_another_device", "logged_out_device_removed":
		return "WhatsApp removed this session because another device took over the connection."
	case "logged_out_banned":
		return "WhatsApp logged out this session and reported a ban or suspension."
	default:
		if strings.TrimSpace(reason) != "" {
			return fmt.Sprintf("WhatsApp logged out this session: %s", reason)
		}
		return "WhatsApp logged out this session."
	}
}

func deriveDisconnectIssueCode(cause string) string {
	switch strings.ToLower(strings.TrimSpace(cause)) {
	case "temporary_ban":
		return "temporary_ban"
	case "stream_replaced":
		return "session_replaced"
	case "connect_failure":
		return "connect_failure"
	case "stream_error":
		return "stream_error"
	case "network":
		return "network_disconnect"
	default:
		if cause == "" {
			return ""
		}
		return strings.ToLower(strings.TrimSpace(cause))
	}
}

func buildDisconnectIssueMessage(cause, details string) string {
	code := deriveDisconnectIssueCode(cause)
	details = strings.TrimSpace(details)

	switch code {
	case "temporary_ban":
		if details != "" {
			return fmt.Sprintf("WhatsApp temporarily restricted this session: %s", details)
		}
		return "WhatsApp temporarily restricted this session."
	case "session_replaced":
		if details != "" {
			return details
		}
		return "Another client connected with the same session."
	case "connect_failure":
		if details != "" {
			return fmt.Sprintf("QuePasa could not reconnect to WhatsApp: %s", details)
		}
		return "QuePasa could not reconnect to WhatsApp."
	case "network_disconnect":
		if details != "" {
			return fmt.Sprintf("WhatsApp disconnected due to a network issue: %s", details)
		}
		return "WhatsApp disconnected due to a network issue."
	case "stream_error":
		if details != "" {
			return fmt.Sprintf("WhatsApp stream returned an error: %s", details)
		}
		return "WhatsApp stream returned an error."
	default:
		if details != "" {
			return details
		}
		if cause != "" {
			return fmt.Sprintf("WhatsApp connection changed state: %s", cause)
		}
		return ""
	}
}

func suggestedActionForIssue(code string, requiresReauth bool) string {
	switch code {
	case "logged_out", "logged_out_another_device", "logged_out_device_removed", "logged_out_banned":
		return "Scan the QR code or request a new pairing code to connect the WhatsApp account again."
	case "temporary_ban":
		return "Wait for the temporary restriction to expire before sending new messages again."
	case "session_replaced":
		return "Close the other active client or reconnect this session from the QuePasa inbox."
	case "connect_failure", "network_disconnect", "stream_error":
		return "Check QuePasa connectivity and try reconnecting the inbox."
	default:
		if requiresReauth {
			return "Reconnect the WhatsApp account from the QuePasa inbox."
		}
		return ""
	}
}

func decodeConnectionDiagnostic(raw any) *QpConnectionDiagnostic {
	if raw == nil {
		return nil
	}

	switch typed := raw.(type) {
	case *QpConnectionDiagnostic:
		if typed == nil {
			return nil
		}
		copy := *typed
		return &copy
	case QpConnectionDiagnostic:
		copy := typed
		return &copy
	}

	payload, err := json.Marshal(raw)
	if err != nil || len(payload) == 0 || string(payload) == "null" {
		return nil
	}

	diagnostic := &QpConnectionDiagnostic{}
	if err := json.Unmarshal(payload, diagnostic); err != nil {
		return nil
	}

	if diagnostic.Code == "" &&
		diagnostic.Message == "" &&
		diagnostic.DisconnectCause == "" &&
		diagnostic.LogoutReason == "" {
		return nil
	}

	return diagnostic
}

func (server *QpWhatsappServer) ConnectionDiagnostic() *QpConnectionDiagnostic {
	if server == nil || server.QpServer == nil {
		return nil
	}

	diagnostic := decodeConnectionDiagnostic(
		server.GetMetadataValue(connectionDiagnosticMetadataKey),
	)
	if diagnostic == nil {
		return nil
	}

	if diagnostic.Code == "" && diagnostic.LogoutReason != "" {
		diagnostic.Code = deriveLogoutIssueCode(diagnostic.LogoutReason)
	}
	if diagnostic.Code == "" && diagnostic.DisconnectCause != "" {
		diagnostic.Code = deriveDisconnectIssueCode(diagnostic.DisconnectCause)
	}
	if diagnostic.Message == "" && diagnostic.LogoutReason != "" {
		diagnostic.Message = buildLogoutIssueMessage(diagnostic.LogoutReason)
	}
	if diagnostic.Message == "" && diagnostic.DisconnectCause != "" {
		diagnostic.Message = buildDisconnectIssueMessage(
			diagnostic.DisconnectCause,
			diagnostic.DisconnectDetails,
		)
	}

	if diagnostic.OccurredAt != nil {
		occurredAt := diagnostic.OccurredAt.UTC()
		diagnostic.OccurredAt = &occurredAt
	}

	diagnostic.SuggestedAction = suggestedActionForIssue(diagnostic.Code, diagnostic.RequiresReauth)

	return diagnostic
}
