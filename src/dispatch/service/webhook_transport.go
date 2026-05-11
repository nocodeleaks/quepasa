package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
	log "github.com/sirupsen/logrus"
)

// WebhookRequest is the outbound HTTP contract used by the dispatch module
// to deliver events/messages to external webhook endpoints.
type WebhookRequest struct {
	ConnectionString string
	Wid              string
	Extra            interface{}
	Timeout          time.Duration
}

type WebhookResponse struct {
	StatusCode int
	Duration   time.Duration
	TimedOut   bool
}

type webhookPayload struct {
	*whatsapp.WhatsappMessage
	Extra interface{} `json:"extra,omitempty"`
}

// SendWebhook performs external HTTP delivery for one message.
// The caller owns domain-level metrics/state updates.
func SendWebhook(message *whatsapp.WhatsappMessage, request *WebhookRequest, logger *log.Entry) (*WebhookResponse, error) {
	if request == nil {
		return &WebhookResponse{}, nil
	}

	startTime := time.Now()

	payload := &webhookPayload{
		WhatsappMessage: message,
		Extra:           request.Extra,
	}

	payloadJSON, err := json.Marshal(&payload)
	if err != nil {
		return &WebhookResponse{}, err
	}

	if logger != nil {
		logger.Debugf("posting webhook payload: %s", payloadJSON)
	}

	req, err := http.NewRequest("POST", request.ConnectionString, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return &WebhookResponse{}, err
	}

	req.Header.Set("User-Agent", "Quepasa")
	req.Header.Set("X-QUEPASA-WID", request.Wid)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if request.Timeout > 0 {
		client.Timeout = request.Timeout
	}

	resp, err := client.Do(req)
	result := &WebhookResponse{Duration: time.Since(startTime)}

	if err != nil {
		if netErr, ok := err.(interface{ Timeout() bool }); ok {
			result.TimedOut = netErr.Timeout()
		}
		return result, err
	}

	defer resp.Body.Close()
	result.StatusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("invalid webhook response status: %d", resp.StatusCode)
	}

	return result, nil
}
