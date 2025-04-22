package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/nocodeleaks/quepasa/models"
	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// TypingRequest defines the parameters for controlling typing indicators
type TypingRequest struct {
	ChatId    string `json:"chat_id"`              // Required: Chat to show typing in
	IsTyping  bool   `json:"is_typing"`            // True to start typing, false to stop
	Duration  int    `json:"duration,omitempty"`   // Optional: Auto-stop after duration (ms)
	MediaType string `json:"media_type,omitempty"` // Optional: "audio" for voice recording indicator
}

// TypingIndicatorController handles API requests for typing indicators
func TypingIndicatorController(w http.ResponseWriter, r *http.Request) {
	// Setting default response type as json
	w.Header().Set("Content-Type", "application/json")

	response := &models.QpResponse{}

	// Get server
	server, err := GetServer(r)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Check if global typing indicators are enabled
	if !whatsapp.Options.ShowTyping {
		response.ParseError(fmt.Errorf("typing indicators are disabled (SHOWTYPING=false)"))
		RespondInterface(w, response)
		return
	}

	// Parse request body
	var request TypingRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.ParseError(fmt.Errorf("invalid request: %v", err))
		RespondInterface(w, response)
		return
	}

	// Validate request
	if request.ChatId == "" {
		response.ParseError(fmt.Errorf("chat_id is required"))
		RespondInterface(w, response)
		return
	}

	// Format and validate the chat ID
	formattedChatId, err := whatsapp.FormatEndpoint(request.ChatId)
	if err != nil {
		response.ParseError(fmt.Errorf("invalid chat_id: %v", err))
		RespondInterface(w, response)
		return
	}

	request.ChatId = formattedChatId

	// Get logger
	logentry := server.GetLogger()

	// Checking for ready state
	status := server.GetStatus()
	if status != whatsapp.Ready {
		err = &ApiServerNotReadyException{Wid: server.GetWId(), Status: status}
		response.ParseError(err)
		RespondInterfaceCode(w, response, http.StatusServiceUnavailable)
		return
	}

	// Send the typing indicator
	err = server.SendChatPresence(request.ChatId, request.IsTyping, request.MediaType)
	if err != nil {
		response.ParseError(err)
		RespondInterface(w, response)
		return
	}

	// Log the action
	if request.IsTyping {
		logentry.Infof("Started typing indicator in chat %s (media: %s)",
			request.ChatId, request.MediaType)
	} else {
		logentry.Infof("Stopped typing indicator in chat %s", request.ChatId)
	}

	// Handle auto-stop if duration is provided
	if request.IsTyping && request.Duration > 0 {
		go func() {
			time.Sleep(time.Duration(request.Duration) * time.Millisecond)
			// Stop typing after duration
			err := server.SendChatPresence(request.ChatId, false, "")
			if err != nil {
				logentry.Warnf("Failed to auto-stop typing: %v", err)
			} else {
				logentry.Infof("Auto-stopped typing in chat %s after %dms",
					request.ChatId, request.Duration)
			}
		}()

		logentry.Infof("Set auto-stop typing after %dms for chat %s",
			request.Duration, request.ChatId)
	}

	// Create successful response
	response.Success = true
	response.ParseSuccess(fmt.Sprintf("Typing indicator %s for chat %s",
		statusText(request.IsTyping), request.ChatId))

	RespondInterface(w, response)
}

// Helper function for status text
func statusText(isTyping bool) string {
	if isTyping {
		return "started"
	}
	return "stopped"
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
