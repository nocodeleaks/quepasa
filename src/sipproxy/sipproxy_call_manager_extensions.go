package sipproxy

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/emiago/sipgo/sip"
)

// generateSIPTag generates a random SIP tag following RFC 3261 recommendations
// Tags should be cryptographically random and at least 32 bits of randomness
func generateSIPTag() string {
	// Generate 8 random bytes (64 bits) for strong randomness
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	// Create completely random tag in hex format
	// Format: XXXXXXXXXXXXXXXX (16 hex chars = 64 bits of randomness)
	return fmt.Sprintf("%x", randomBytes)
}

// getOrGenerateCallTag gets existing tag from CallInfo or generates new one
func (source *SIPCallManagerSipgo) getOrGenerateCallTag(callID string) string {
	// Check if call exists and has a tag
	if callInfo, exists := source.activeCalls[callID]; exists && callInfo.SIPTag != "" {
		source.logger.Infof("🏷️  Using existing SIP tag for call %s: %s", callID, callInfo.SIPTag)
		return callInfo.SIPTag
	}

	// Generate new tag and store it
	newTag := generateSIPTag()
	if callInfo, exists := source.activeCalls[callID]; exists {
		callInfo.SIPTag = newTag
		source.logger.Infof("🏷️  Generated and stored new SIP tag for call %s: %s", callID, newTag)
	} else {
		source.logger.Infof("🏷️  Generated SIP tag (call not found in active calls): %s", newTag)
	}

	return newTag
}

func (source *SIPCallManagerSipgo) GetRecipient(toPhone string) (recipient sip.Uri) {

	// Create recipient URI with port optimization (omit :5060 to reduce INVITE size)
	recipient = sip.Uri{
		Scheme: "sip",
		User:   toPhone,
		Host:   source.config.SIPServer,
		Port:   int(source.config.SIPPort),
	}

	// Only omit port if it's the default SIP port (5060) for size optimization
	if recipient.Port == 5060 {
		recipient.Port = 0 // sipgo will omit port when set to 0
	}

	return
}

func (source *SIPCallManagerSipgo) SetFromHeader(headers []sip.Header, fromPhone, callID string) []sip.Header {

	publicIP := source.networkManager.GetPublicIP()
	localPort := source.networkManager.GetLocalPort()

	// Get or generate unique SIP tag for this call
	sipTag := source.getOrGenerateCallTag(callID)

	// Configure From header with port optimization
	// Only include listener_port if not 5060 (default SIP port)
	uri := sip.Uri{
		Scheme: "sip",
		User:   fromPhone, // Use caller ID phone number
		Host:   publicIP,  // Use public/local IP
		Port:   localPort,
	}

	// Only omit port if it's the default SIP port (5060) for size optimization
	if uri.Port == 5060 {
		uri.Port = 0 // sipgo will omit port when set to 0
	}

	// Add custom From header: <sip:phone@ip:port>;tag=XXXXX (no display name, optimized port)
	// The tag parameter must be in the header parameters, not in URI parameters
	header := &sip.FromHeader{
		DisplayName: "", // Empty display name - only shows phone number
		Address:     uri,
		Params:      sip.NewParams().Add("tag", sipTag), // Tag goes in header params, not URI params
	}

	return append(headers, header)
}

// Add custom Via header with correct IP instead of 0.0.0.0
// Format: Via: SIP/2.0/UDP 177.36.191.201:1852;rport;branch=z9hG4bK...
// Port optimization: omit :5060 to reduce INVITE size
func (source *SIPCallManagerSipgo) SetViaHeader(headers []sip.Header) []sip.Header {

	localIP := source.networkManager.GetLocalIP()
	localPort := source.networkManager.GetLocalPort()

	header := &sip.ViaHeader{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       "UDP",
		Host:            localIP,
		Params:          sip.NewParams(),
	}

	// Generate branch in hex format like real SIP servers: z9hG4bKPj7129cc790c0e457c99487c364d48a3cb
	branch := fmt.Sprintf("z9hG4bK%x", time.Now().UnixNano()) // Generate unique branch in hex format
	header.Params.Add("branch", branch)

	// Only include port if not the default SIP port (5060) for size optimization
	if localPort != 5060 {
		header.Port = localPort
		header.Params.Add("rport", "")
	} else {
		header.Port = 0 // sipgo will omit port when set to 0
	}

	return append(headers, header)
}

// Use WhatsApp Call-ID as SIP Call-ID for simplified tracking
// This eliminates the need for Call-ID mapping between WhatsApp and SIP
func SetCallIDHeader(headers []sip.Header, callID string) []sip.Header {

	// Use WhatsApp Call-ID directly
	callIDHeader := sip.CallIDHeader(callID)

	return append(headers, &callIDHeader)
}

// createSDPOffer creates a basic SDP offer for audio call
func (source *SIPCallManagerSipgo) CreateSDPOffer(fromPhone string) string {
	// Generate a robust SDP for audio call
	// Using proper SDP structure with standard codecs
	sessionID := time.Now().Unix()
	sessionVersion := sessionID + 1 // Version should be different from session ID

	// Use dynamic RTP port range (typically 10000-20000 for RTP)
	// Simple port allocation to avoid conflicts
	rtpPort := 10000 + (sessionID % 1000)

	// Get both local and public IPs from network manager
	localIP := source.networkManager.GetLocalIP()
	publicIP := source.networkManager.GetPublicIP()

	// Use public IP for media connection (RTP) to work through NAT
	// Use local IP for session origin info
	return fmt.Sprintf(`v=0
o=%s %d %d IN IP4 %s
s=%s
c=IN IP4 %s
t=0 0
m=audio %d RTP/AVP 0 8 101
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-15
a=ptime:20
a=maxptime:20
a=sendrecv
`,
		fromPhone,
		sessionID,
		sessionVersion,
		localIP, // Origin IP (can be local)
		source.config.SDPSessionName,
		publicIP, // Connection IP (public for NAT traversal)
		rtpPort)
}

/*
	// Add Content-Type header for SDP body
	contentTypeHeader := sip.ContentTypeHeader("application/sdp")
	headers = append(headers, &contentTypeHeader)
	scm.logger.Infof("📄 Content-Type header added: application/sdp")

*/
/*
	// Add Allow header with supported SIP methods
	// Create a custom Allow header by implementing the Header interface
	allowValue := "PRACK, INVITE, ACK, BYE, CANCEL, UPDATE, INFO, SUBSCRIBE, NOTIFY, REFER, MESSAGE, OPTIONS"
	allowHeader := sip.NewHeader("Allow", allowValue)
	if allowHeader != nil {
		headers = append(headers, allowHeader)
		scm.logger.Infof("📋 Allow header added: %s", allowValue)
	} else {
		scm.logger.Warnf("⚠️ Could not create Allow header, sipgo may not support custom headers")
	}
*/

/*
	userAgent := scm.config.UserAgent
	userAgentHeader := sip.NewHeader("User-Agent", userAgent)
	if userAgentHeader != nil {
		headers = append(headers, userAgentHeader)
		scm.logger.Infof("📋 User-Agent header added: %s", userAgent)
	} else {
		scm.logger.Warnf("⚠️ Could not create User-Agent header, sipgo may not support custom headers")
	}
*/
