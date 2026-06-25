package sipproxy

import (
	"strings"

	"github.com/emiago/sipgo/sip"
	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// SIPResponseHandler handles SIP response processing
type SIPResponseHandler struct {
	logger             qplog.Logger
	transactionMonitor *SIPTransactionMonitor
}

// NewSIPResponseHandler creates a new SIP response handler
func NewSIPResponseHandler(logger qplog.Logger, transactionMonitor *SIPTransactionMonitor) *SIPResponseHandler {
	return &SIPResponseHandler{
		logger:             logger,
		transactionMonitor: transactionMonitor,
	}
}

// ProcessImmediateResponse processes immediate SIP responses
func (srh *SIPResponseHandler) ProcessImmediateResponse(statusCode, callID, fromPhone, toPhone, response string) {
	srh.logger.Infof("🔍 Processing immediate SIP response: %s for CallID: %s", statusCode, callID)

	// Parse status code for proper handling
	switch statusCode {
	case "200":
		srh.logger.Infof("✅ Call %s immediately ACCEPTED (200 OK)", callID)
		if srh.transactionMonitor.callAcceptedHandler != nil {
			mockResponse := &sip.Response{
				StatusCode: 200,
				Reason:     "OK",
			}
			srh.logger.Infof("🎉 Calling onCallAccepted handler for immediate 200 OK")
			srh.transactionMonitor.callAcceptedHandler(callID, fromPhone, toPhone, mockResponse)
		}
	case "603":
		srh.logger.Infof("❌ Call %s immediately REJECTED (603 Decline)", callID)
		if srh.transactionMonitor.callRejectedHandler != nil {
			mockResponse := &sip.Response{
				StatusCode: 603,
				Reason:     "Decline",
			}
			srh.logger.Infof("💔 Calling onCallRejected handler for immediate 603")
			srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
		}
	default:
		if statusCode[0] >= '4' { // 4xx, 5xx, 6xx
			srh.logger.Infof("❌ Call %s immediately REJECTED (%s)", callID, statusCode)
			if srh.transactionMonitor.callRejectedHandler != nil {
				mockResponse := srh.createMockErrorResponse(statusCode)
				srh.logger.Infof("💔 Calling onCallRejected handler for %s", statusCode)
				srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
			}
		} else {
			srh.logger.Infof("📞 Call %s PROGRESS (%s)", callID, statusCode)
		}
	}
}

// ProcessContinuousResponse processes responses in continuous monitoring
func (srh *SIPResponseHandler) ProcessContinuousResponse(response string, callID, fromPhone, toPhone string) (shouldStop bool) {
	srh.logger.Infof("🔄 ProcessContinuousResponse called for CallID: %s", callID)
	srh.logger.Infof("🌟 SIP SERVER RESPONSE RECEIVED! Processing response for call %s → %s", fromPhone, toPhone)

	// 🆕 CHECK FOR SIP REQUESTS (BYE, CANCEL, etc.) - not just responses
	if strings.Contains(response, "BYE ") || strings.Contains(response, "CANCEL ") {
		srh.logger.Infof("🚨🚨🚨 SIP REQUEST RECEIVED FROM SERVER! 🚨🚨🚨")
		srh.logger.Infof("📡 THIS IS A SIP REQUEST (not response) - server is asking us to do something!")
		srh.handleSIPRequest(response, callID, fromPhone, toPhone)
		// Continue monitoring for more messages
		return false
	}

	// Parse the status code and handle the response
	if strings.Contains(response, "SIP/2.0") {
		lines := strings.Split(response, "\n")
		if len(lines) > 0 {
			statusLine := strings.TrimSpace(lines[0])
			if strings.HasPrefix(statusLine, "SIP/2.0 ") {
				parts := strings.Fields(statusLine)
				if len(parts) >= 3 {
					statusCode := parts[1]
					srh.logger.Infof("🎯 Processing SIP response: %s for CallID: %s", statusCode, callID)

					// Handle different response types
					switch statusCode {
					case "100":
						srh.logger.Infof("📞 100 Trying received - call is being processed")
						srh.logger.Infof("⏳ SIP server is processing the call...")
						// Continue monitoring for final response
						return false
					case "180":
						srh.logger.Infof("📞 180 Ringing received - destination is ringing")
						srh.logger.Infof("🔔 Phone is ringing on SIP server side...")
						// Continue monitoring for final response
						return false
					case "183":
						srh.logger.Infof("📞 183 Session Progress received - early media")
						srh.logger.Infof("🎵 Session progress with early media...")
						// Continue monitoring for final response
						return false
					case "200":
						srh.logger.Infof("✅ 200 OK received - CALL ACCEPTED!")
						srh.logger.Infof("🎉🎉🎉 ACCEPT RESPONSE DETECTED! WHATSAPP CALL WILL CONTINUE! 🎉🎉🎉")
						srh.logger.Infof("📞 Call %s has been ACCEPTED by SIP server", callID)
						if srh.transactionMonitor.callAcceptedHandler != nil {
							mockResponse := &sip.Response{
								StatusCode: 200,
								Reason:     "OK",
							}
							srh.logger.Infof("🎉 Calling onCallAccepted handler for 200 OK response")
							srh.transactionMonitor.callAcceptedHandler(callID, fromPhone, toPhone, mockResponse)
						}

						// 🆕 CONTINUE monitoring after 200 OK to capture post-accept messages
						srh.logger.Infof("👂 CONTINUING to monitor SIP server for post-accept messages (BYE, CANCEL, etc.)")
						srh.logger.Infof("🔍 Will log ANY message from SIP server after this 200 OK")
						return false // Changed from true to false - KEEP MONITORING!
					case "603":
						srh.logger.Infof("❌ 603 Decline received - CALL REJECTED!")
						srh.logger.Infof("🔥🔥🔥 REJECT RESPONSE DETECTED! WILL TERMINATE WHATSAPP CALL! 🔥🔥🔥")
						if srh.transactionMonitor.callRejectedHandler != nil {
							mockResponse := &sip.Response{
								StatusCode: 603,
								Reason:     "Decline",
							}
							srh.logger.Infof("💔 Calling onCallRejected handler for 603 - this should terminate WhatsApp call")
							srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
						} else {
							srh.logger.Errorf("❌ callRejectedHandler is nil! Cannot process rejection!")
						}
						// Final response - stop monitoring
						return true
					default:
						if statusCode[0] >= '4' { // 4xx, 5xx, 6xx
							srh.logger.Infof("❌ %s received - CALL REJECTED!", statusCode)
							srh.logger.Infof("💔 SIP server rejected the call with status: %s", statusCode)
							if srh.transactionMonitor.callRejectedHandler != nil {
								mockResponse := srh.createMockErrorResponse(statusCode)
								srh.logger.Infof("💔 Calling onCallRejected handler for status: %s", statusCode)
								srh.transactionMonitor.callRejectedHandler(callID, fromPhone, toPhone, mockResponse)
							}
							// Final response - stop monitoring
							return true
						} else {
							srh.logger.Infof("📞 %s received - call progress", statusCode)
							srh.logger.Infof("📡 SIP server sent progress response: %s - continuing to monitor...", statusCode)
							// Continue monitoring for final response
							return false
						}
					}
				}
			}
		}
	} else {
		srh.logger.Warnf("⚠️ Non-SIP response received for CallID: %s", callID)
		srh.logger.Warnf("🤔 Response doesn't contain 'SIP/2.0' - might be malformed or non-SIP")
		srh.logger.Warnf("📄 Raw response content: %s", response)
	}

	// Continue monitoring by default
	srh.logger.Infof("🔄 Continuing to monitor for more responses...")
	return false
}

// LogResponse logs a SIP response with proper formatting
func (srh *SIPResponseHandler) LogResponse(response string, callID string, responseNumber int, bytesReceived int) {
	srh.logger.Infof("📨 SIP Response #%d received (%d bytes) for CallID: %s:", responseNumber, bytesReceived, callID)
	srh.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Parse and log the response
	for _, line := range strings.Split(response, "\n") {
		if strings.TrimSpace(line) != "" {
			srh.logger.Infof("%s", line)
		}
	}
	srh.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// createMockErrorResponse creates a mock SIP error response based on status code
func (srh *SIPResponseHandler) createMockErrorResponse(statusCode string) *sip.Response {
	mockResponse := &sip.Response{
		StatusCode: 400, // Default
		Reason:     "Client Error",
	}

	// Parse actual status code if possible
	if len(statusCode) >= 3 {
		switch statusCode {
		case "404":
			mockResponse.StatusCode = 404
			mockResponse.Reason = "Not Found"
		case "486":
			mockResponse.StatusCode = 486
			mockResponse.Reason = "Busy Here"
		case "480":
			mockResponse.StatusCode = 480
			mockResponse.Reason = "Temporarily Unavailable"
		default:
			if statusCode[0] == '5' {
				mockResponse.StatusCode = 500
				mockResponse.Reason = "Server Error"
			} else if statusCode[0] == '6' {
				mockResponse.StatusCode = 600
				mockResponse.Reason = "Global Failure"
			}
		}
	}

	return mockResponse
}

// handleSIPRequest processes SIP requests received from the server (BYE, CANCEL, etc.)
func (srh *SIPResponseHandler) handleSIPRequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("🔍 Analyzing SIP request from server for CallID: %s", callID)

	// Log the complete request with detailed formatting
	srh.logSIPRequestDetails(request, callID)

	// Parse the request line to determine the method
	lines := strings.Split(request, "\n")
	if len(lines) > 0 {
		requestLine := strings.TrimSpace(lines[0])
		srh.logger.Infof("📋 SIP Request Line: %s", requestLine)

		// Check for specific SIP methods
		if strings.HasPrefix(requestLine, "BYE ") {
			srh.handleBYERequest(request, callID, fromPhone, toPhone)
		} else if strings.HasPrefix(requestLine, "CANCEL ") {
			srh.handleCANCELRequest(request, callID, fromPhone, toPhone)
		} else if strings.HasPrefix(requestLine, "INVITE ") {
			srh.handleINVITERequest(request, callID, fromPhone, toPhone)
		} else if strings.HasPrefix(requestLine, "ACK ") {
			srh.handleACKRequest(request, callID, fromPhone, toPhone)
		} else {
			srh.handleOtherSIPRequest(requestLine, request, callID, fromPhone, toPhone)
		}
	}
}

// logSIPRequestDetails logs the complete SIP request with formatting
func (srh *SIPResponseHandler) logSIPRequestDetails(request string, callID string) {
	srh.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	srh.logger.Infof("📨 COMPLETE SIP REQUEST FROM SERVER (CallID: %s):", callID)
	srh.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	lines := strings.Split(request, "\n")
	for i, line := range lines {
		cleanLine := strings.TrimSpace(line)
		if cleanLine != "" {
			srh.logger.Infof("📄 Line %d: %s", i+1, cleanLine)
		}
	}

	srh.logger.Infof("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// handleBYERequest processes BYE requests from the server
func (srh *SIPResponseHandler) handleBYERequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("🛑🛑🛑 BYE REQUEST RECEIVED FROM SIP SERVER! 🛑🛑🛑")
	srh.logger.Infof("📞 SIP server is requesting to TERMINATE the call!")
	srh.logger.Infof("💡 CallID: %s | Call: %s → %s", callID, fromPhone, toPhone)
	srh.logger.Infof("🎯 ACTION NEEDED: Should terminate WhatsApp call and respond with 200 OK")

	// TODO: Implement call termination logic here
	// For now, just log what should happen
	srh.logger.Infof("📋 Next steps (TODO):")
	srh.logger.Infof("   1. 📞❌ Terminate WhatsApp call for CallID: %s", callID)
	srh.logger.Infof("   2. 📤 Send 200 OK response to SIP server")
	srh.logger.Infof("   3. 🧹 Clean up call resources")
}

// handleCANCELRequest processes CANCEL requests from the server
func (srh *SIPResponseHandler) handleCANCELRequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("❌❌❌ CANCEL REQUEST RECEIVED FROM SIP SERVER! ❌❌❌")
	srh.logger.Infof("📞 SIP server is requesting to CANCEL the ongoing invitation!")
	srh.logger.Infof("💡 CallID: %s | Call: %s → %s", callID, fromPhone, toPhone)
	srh.logger.Infof("🎯 ACTION NEEDED: Should cancel WhatsApp call and respond with 200 OK")

	// TODO: Implement call cancellation logic here
	srh.logger.Infof("📋 Next steps (TODO):")
	srh.logger.Infof("   1. 📞❌ Cancel WhatsApp call invitation for CallID: %s", callID)
	srh.logger.Infof("   2. 📤 Send 200 OK response to SIP server")
	srh.logger.Infof("   3. 🧹 Clean up invitation resources")
}

// handleINVITERequest processes INVITE requests from the server
func (srh *SIPResponseHandler) handleINVITERequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("📞📞📞 INVITE REQUEST RECEIVED FROM SIP SERVER! 📞📞📞")
	srh.logger.Infof("🔄 SIP server is sending a new or re-INVITE!")
	srh.logger.Infof("💡 CallID: %s | Call: %s → %s", callID, fromPhone, toPhone)
	srh.logger.Infof("🎯 This might be a call modification or re-negotiation")
}

// handleACKRequest processes ACK requests from the server
func (srh *SIPResponseHandler) handleACKRequest(request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("✅✅✅ ACK REQUEST RECEIVED FROM SIP SERVER! ✅✅✅")
	srh.logger.Infof("🤝 SIP server is acknowledging our response!")
	srh.logger.Infof("💡 CallID: %s | Call: %s → %s", callID, fromPhone, toPhone)
	srh.logger.Infof("🎯 This confirms the SIP transaction is complete")
}

// handleOtherSIPRequest processes other SIP requests
func (srh *SIPResponseHandler) handleOtherSIPRequest(requestLine, request string, callID, fromPhone, toPhone string) {
	srh.logger.Infof("❓❓❓ UNKNOWN SIP REQUEST RECEIVED FROM SERVER! ❓❓❓")
	srh.logger.Infof("📋 Request Line: %s", requestLine)
	srh.logger.Infof("💡 CallID: %s | Call: %s → %s", callID, fromPhone, toPhone)
	srh.logger.Infof("🎯 This is an unrecognized SIP method - logging for analysis")
}
