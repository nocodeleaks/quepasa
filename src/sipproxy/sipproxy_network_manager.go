package sipproxy

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

// SIPNetworkManager handles UDP networking and STUN operations
type SIPProxyNetworkManager struct {
	SIPProxyNetworkManagerSettings

	logger *logrus.Entry
}

// NewSIPProxyNetworkManager creates a new SIP proxy network manager
func NewSIPProxyNetworkManager(settings SIPProxyNetworkManagerSettings, logentry *logrus.Entry) *SIPProxyNetworkManager {
	internalLogentry := logentry.WithField("component", "network")
	return &SIPProxyNetworkManager{
		SIPProxyNetworkManagerSettings: settings,
		logger:                         internalLogentry,
	}
}

// ConfigureNetwork sets up network discovery using sipgo native methods with STUN fallback
func (snm *SIPProxyNetworkManager) ConfigureNetwork() error {
	// Discover public IP using sipgo native methods with STUN fallback
	publicIP, localIP, err := snm.discoverPublicIPWithSipgoAndSTUN()
	if err != nil {
		snm.logger.Errorf("❌ Failed to discover public IP: %v", err)
		return err
	}

	snm.PublicIP = publicIP
	snm.LocalIP = localIP

	// If local port is 0, discover it now
	if snm.LocalPort == 0 {
		if err := snm.DiscoverLocalPort(); err != nil {
			snm.logger.Errorf("❌ Failed to discover local port: %v", err)
			return err
		}
	}

	snm.IsSTUNConfigured = true

	snm.logger.Infof("🌐 Network configuration complete:")
	snm.logger.Infof("   📍 Public IP: %s", snm.PublicIP)
	snm.logger.Infof("   🏠 Local IP: %s", snm.LocalIP)
	snm.logger.Infof("   🔌 Local Port: %d", snm.LocalPort)

	return nil
}

// GetPublicIP returns the discovered public IP
func (snm *SIPProxyNetworkManager) GetPublicIP() string {
	return snm.PublicIP
}

// GetLocalIP returns the local IP
func (snm *SIPProxyNetworkManager) GetLocalIP() string {
	return snm.LocalIP
}

// GetLocalPort returns the local port
func (snm *SIPProxyNetworkManager) GetLocalPort() int {
	return snm.LocalPort
}

// IsConfigured returns whether network is properly configured
func (snm *SIPProxyNetworkManager) IsConfigured() bool {
	return snm.IsSTUNConfigured
}

// discoverPublicIPWithSipgoAndSTUN uses sipgo native capabilities with STUN fallback for NAT discovery
func (snm *SIPProxyNetworkManager) discoverPublicIPWithSipgoAndSTUN() (publicIP, localIP string, err error) {
	snm.logger.Info("🔍 Discovering public IP using sipgo native methods with STUN fallback")

	// First, get local IP using sipgo-style network discovery
	localIP, err = snm.discoverLocalIPWithSipgo()
	if err != nil {
		snm.logger.WithError(err).Warn("⚠️  Failed to discover local IP, using fallback")
		localIP = "127.0.0.1"
	} else {
		snm.logger.WithField("local_ip", localIP).Info("🏠 Local IP discovered successfully")
	}

	// If public IP is manually configured, use it
	if snm.PublicIP != "" {
		snm.logger.WithFields(logrus.Fields{
			"public_ip": snm.PublicIP,
			"local_ip":  localIP,
		}).Info("✅ Using configured public IP")
		return snm.PublicIP, localIP, nil
	}

	// Try STUN discovery only if needed for NAT traversal
	if snm.StunServer != "" && snm.StunServer != ":0" {
		snm.logger.WithField("stun_server", snm.StunServer).Info("🌐 Using STUN for NAT traversal")
		stunManager := NewSIPProxySTUNManager(snm.logger, snm.StunServer)
		discoveredIP, err := stunManager.DiscoverPublicIPv4WithFallback()
		if err != nil {
			snm.logger.WithError(err).Warn("⚠️  STUN discovery failed, using local IP as public IP")
			return localIP, localIP, nil
		}
		return discoveredIP, localIP, nil
	}

	// If no STUN configured, assume no NAT (local network)
	snm.logger.WithFields(logrus.Fields{
		"local_ip":  localIP,
		"public_ip": localIP,
	}).Info("📡 No STUN server configured, using local IP as public IP (no NAT)")
	return localIP, localIP, nil
}

// discoverLocalIPWithSipgo discovers local IP using sipgo-style network detection (IPv4 only)
func (snm *SIPProxyNetworkManager) discoverLocalIPWithSipgo() (string, error) {
	// Method 1: Connect to a public IPv4 address to determine best local IPv4 interface
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		localIP := localAddr.IP.String()

		// Double-check it's IPv4
		if ipv4 := localAddr.IP.To4(); ipv4 != nil {
			snm.logger.WithField("local_ipv4", localIP).Info("🏠 Local IPv4 discovered via UDP4 routing to Google DNS")
			return localIP, nil
		} else {
			snm.logger.WithField("unexpected_ip", localIP).Warn("⚠️  UDP4 connection returned non-IPv4 address, trying interface detection")
		}
	}

	snm.logger.WithError(err).Info("UDP4 routing method failed, trying network interface detection for IPv4")

	// Method 2: Find first non-loopback IPv4 interface (sipgo style)
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	snm.logger.Debug("🔍 Scanning network interfaces for IPv4 addresses")
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				// Explicitly check for IPv4
				if ipv4 := ipNet.IP.To4(); ipv4 != nil {
					localIP := ipv4.String()
					snm.logger.WithFields(logrus.Fields{
						"local_ipv4": localIP,
						"interface":  iface.Name,
					}).Info("🏠 Local IPv4 discovered via interface detection")
					return localIP, nil
				} else {
					// Log IPv6 addresses found but skip them
					snm.logger.WithFields(logrus.Fields{
						"ipv6_address": ipNet.IP.String(),
						"interface":    iface.Name,
					}).Debug("🚫 Skipping IPv6 address (IPv4 required)")
				}
			}
		}
	}

	return "", fmt.Errorf("no suitable IPv4 address found on any network interface")
}

// GetSIPServerEndpoint returns the configured SIP server endpoint
func (snm *SIPProxyNetworkManager) GetSIPServerEndpoint() string {
	return fmt.Sprintf("%s:%d", snm.SIPServer, snm.SIPPort)
}

// DiscoverLocalPort discovers the local port by creating a test UDP connection
func (snm *SIPProxyNetworkManager) DiscoverLocalPort() error {
	if snm.LocalPort != 0 {
		snm.logger.Infof("🔌 Local port already set: %d", snm.LocalPort)
		return nil
	}

	snm.logger.Infof("🔍 Discovering local port via test connection to %s:%d...", snm.SIPServer, snm.SIPPort)

	// Resolve SIP server address
	snm.logger.Infof("🌐 Resolving SIP server address...")
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", snm.SIPServer, snm.SIPPort))
	if err != nil {
		snm.logger.Errorf("❌ Failed to resolve SIP server: %v", err)
		return fmt.Errorf("failed to resolve SIP server: %v", err)
	}
	snm.logger.Infof("✅ SIP server resolved: %s", serverAddr)

	// Create temporary UDP connection to discover local port
	snm.logger.Infof("🔌 Creating test UDP connection...")
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		snm.logger.Errorf("❌ Failed to create test UDP connection: %v", err)
		return fmt.Errorf("failed to create test UDP connection: %v", err)
	}
	defer conn.Close()
	snm.logger.Infof("✅ Test UDP connection established")

	// Get the local address and port
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	snm.LocalPort = localAddr.Port

	snm.logger.Infof("🔌 Local port discovered: %d", snm.LocalPort)
	return nil
}
