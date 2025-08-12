package sipproxy

import (
	"fmt"
	"net"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
	log "github.com/sirupsen/logrus"
)

// UPnPManager handles automatic port forwarding via UPnP
type UPnPManager struct {
	logger         *log.Entry
	client         *internetgateway2.WANIPConnection1
	portMapped     bool
	externalIP     string
	mappedPort     int
	mappedProtocol string
}

// NewUPnPManager creates a new UPnP manager instance
func NewUPnPManager(logger *log.Entry) *UPnPManager {
	return &UPnPManager{
		logger: logger,
	}
}

// Setup configures UPnP for automatic port forwarding
func (um *UPnPManager) Setup() error {
	um.logger.Infof("🔌 Setting up UPnP for automatic port forwarding")

	// Descobrir dispositivos UPnP na rede
	clients, _, err := internetgateway2.NewWANIPConnection1Clients()
	if err != nil {
		// Tentar IGDv1 se IGDv2 falhar
		clients1, _, err1 := internetgateway1.NewWANIPConnection1Clients()
		if err1 != nil {
			return fmt.Errorf("no UPnP IGD devices found (IGDv2: %v, IGDv1: %v)", err, err1)
		}

		// Converter cliente IGDv1 para IGDv2 (compatibilidade)
		if len(clients1) > 0 {
			um.logger.Infof("🔌 Found IGDv1 device, using compatibility mode")
			// Para simplicidade, vamos focar no IGDv2 por enquanto
			return fmt.Errorf("IGDv1 not supported yet, please upgrade router firmware")
		}
	}

	if len(clients) == 0 {
		return fmt.Errorf("no UPnP IGDv2 devices found")
	}

	um.client = clients[0]
	um.logger.Infof("✅ UPnP IGDv2 device found and configured")

	// Descobrir IP externo via UPnP
	externalIP, err := um.client.GetExternalIPAddress()
	if err != nil {
		um.logger.Warnf("⚠️ Failed to get external IP via UPnP: %v", err)
	} else {
		um.externalIP = externalIP
		um.logger.Infof("🌐 UPnP discovered external IP: %s", um.externalIP)
	}

	return nil
}

// OpenPort opens a port via UPnP port forwarding
func (um *UPnPManager) OpenPort(port int, protocol string) error {
	if um.client == nil {
		return fmt.Errorf("UPnP client not initialized")
	}

	localIP := um.getLocalIP()
	um.logger.Infof("🔌 Opening UPnP port %d/%s for local IP %s", port, protocol, localIP)

	// Mapear porta (externa = interna)
	externalPort := uint16(port)
	internalPort := uint16(port)

	err := um.client.AddPortMapping(
		"", // RemoteHost (vazio = qualquer)
		externalPort,
		protocol,
		internalPort,
		localIP,
		true,                // Enabled
		"QuePasa SIP Proxy", // Description
		0,                   // LeaseDuration (0 = permanente)
	)

	if err != nil {
		return fmt.Errorf("failed to add UPnP port mapping: %v", err)
	}

	um.portMapped = true
	um.mappedPort = port
	um.mappedProtocol = protocol
	um.logger.Infof("✅ UPnP port %d/%s opened successfully", port, protocol)

	return nil
}

// ClosePort closes the UPnP port forwarding
func (um *UPnPManager) ClosePort() error {
	if um.client == nil || !um.portMapped {
		return nil
	}

	um.logger.Infof("🔌 Closing UPnP port %d/%s", um.mappedPort, um.mappedProtocol)

	err := um.client.DeletePortMapping("", uint16(um.mappedPort), um.mappedProtocol)
	if err != nil {
		um.logger.Warnf("⚠️ Failed to close UPnP port: %v", err)
		return err
	}

	um.portMapped = false
	um.logger.Infof("✅ UPnP port %d/%s closed successfully", um.mappedPort, um.mappedProtocol)

	return nil
}

// IsPortMapped returns true if a port is currently mapped
func (um *UPnPManager) IsPortMapped() bool {
	return um.portMapped
}

// GetExternalIP returns the external IP discovered via UPnP
func (um *UPnPManager) GetExternalIP() string {
	return um.externalIP
}

// getLocalIP obtém o IP local da máquina
func (um *UPnPManager) getLocalIP() string {
	// Tentar conectar ao Google DNS para descobrir IP local
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		um.logger.Warnf("Failed to get local IP: %v", err)
		return "192.168.1.100" // fallback
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
