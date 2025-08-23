package sipproxy

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/stun"
	log "github.com/sirupsen/logrus"
)

// SIPProxySTUNManager handles public IP discovery via STUN servers with fallback support
type SIPProxySTUNManager struct {
	logger           *log.Entry
	configuredServer string
	fallbackServers  []string
}

// NewSIPProxySTUNManager creates a new STUN manager with configured server and fallback list
func NewSIPProxySTUNManager(logger *log.Entry, configuredServer string) *SIPProxySTUNManager {
	return &SIPProxySTUNManager{
		logger:           logger,
		configuredServer: configuredServer,
		fallbackServers: []string{
			"stun1.l.google.com:19302",
			"stun2.l.google.com:19302",
			"stun3.l.google.com:19302",
			"stun4.l.google.com:19302",
			"stun.stunprotocol.org:3478",
			"stun.sipgate.net:3478",
			"stun.ekiga.net:3478",
		},
	}
}

// DiscoverPublicIPv4WithFallback discovers public IPv4 address using configured server with fallback
func (sm *SIPProxySTUNManager) DiscoverPublicIPv4WithFallback() (string, error) {
	// First try configured server if provided
	if sm.configuredServer != "" && sm.configuredServer != ":0" {
		sm.logger.WithField("stun_server", sm.configuredServer).Debug("Trying configured STUN server")
		if ip, err := sm.discoverIPFromServer(sm.configuredServer); err == nil {
			sm.logger.WithFields(log.Fields{
				"public_ip":   ip,
				"stun_server": sm.configuredServer,
			}).Infof("Public IP discovered using configured STUN server, IP: %s", ip)
			return ip, nil
		} else {
			sm.logger.WithError(err).WithField("stun_server", sm.configuredServer).Warn("Failed to discover IP using configured server, trying fallback")
		}
	} else {
		sm.logger.Debug("No configured STUN server or empty value, using fallback servers")
	}

	// Try fallback servers
	for _, server := range sm.fallbackServers {
		sm.logger.WithField("stun_server", server).Debug("Trying fallback STUN server")
		if ip, err := sm.discoverIPFromServer(server); err == nil {
			sm.logger.WithFields(log.Fields{
				"public_ip":   ip,
				"stun_server": server,
			}).Infof("Public IP discovered using fallback STUN server, IP: %s", ip)
			return ip, nil
		} else {
			sm.logger.WithError(err).WithField("stun_server", server).Debug("Failed with this fallback server")
		}
	}

	return "", fmt.Errorf("failed to discover public IP using all available STUN servers")
}

// discoverIPFromServer performs STUN discovery from a specific server (IPv4 only)
func (sm *SIPProxySTUNManager) discoverIPFromServer(stunServer string) (string, error) {
	sm.logger.WithField("stun_server", stunServer).Debug("Attempting IPv4 STUN discovery")

	// Create UDP4 connection to STUN server to force IPv4
	conn, err := net.Dial("udp4", stunServer)
	if err != nil {
		return "", fmt.Errorf("failed to connect to STUN server %s via IPv4: %w", stunServer, err)
	}
	defer conn.Close()

	// Set connection timeout
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Create STUN binding request
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// Send request
	_, err = conn.Write(message.Raw)
	if err != nil {
		return "", fmt.Errorf("failed to send STUN request to %s: %w", stunServer, err)
	}

	// Read response
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read STUN response from %s: %w", stunServer, err)
	}

	// Parse STUN response
	response := &stun.Message{Raw: buffer[:n]}
	if err := response.Decode(); err != nil {
		return "", fmt.Errorf("failed to decode STUN response from %s: %w", stunServer, err)
	}

	// Extract mapped address
	var mappedAddr stun.XORMappedAddress
	if err := mappedAddr.GetFrom(response); err != nil {
		// Try regular mapped address if XOR mapped address fails
		var regularAddr stun.MappedAddress
		if err := regularAddr.GetFrom(response); err != nil {
			return "", fmt.Errorf("failed to get mapped address from STUN response: %w", err)
		}

		// Validate that we got an IPv4 address
		if ipv4 := regularAddr.IP.To4(); ipv4 != nil {
			publicIP := regularAddr.IP.String()
			sm.logger.WithFields(log.Fields{
				"public_ipv4": publicIP,
				"stun_server": stunServer,
			}).Debug("Successfully discovered public IPv4 via STUN (regular address)")
			return publicIP, nil
		} else {
			return "", fmt.Errorf("STUN server %s returned IPv6 address, but IPv4 is required", stunServer)
		}
	}

	// Validate that we got an IPv4 address from XOR mapped address
	if ipv4 := mappedAddr.IP.To4(); ipv4 != nil {
		publicIP := mappedAddr.IP.String()
		sm.logger.WithFields(log.Fields{
			"public_ipv4": publicIP,
			"stun_server": stunServer,
		}).Debug("Successfully discovered public IPv4 via STUN (XOR mapped address)")
		return publicIP, nil
	} else {
		return "", fmt.Errorf("STUN server %s returned IPv6 address, but IPv4 is required", stunServer)
	}
}

// GetConfiguredServer returns the configured STUN server
func (sm *SIPProxySTUNManager) GetConfiguredServer() string {
	return sm.configuredServer
}

// GetFallbackServers returns the list of fallback STUN servers
func (sm *SIPProxySTUNManager) GetFallbackServers() []string {
	return sm.fallbackServers
}

// HasValidConfiguredServer checks if a valid configured server is available
func (sm *SIPProxySTUNManager) HasValidConfiguredServer() bool {
	return sm.configuredServer != "" && sm.configuredServer != ":0"
}
