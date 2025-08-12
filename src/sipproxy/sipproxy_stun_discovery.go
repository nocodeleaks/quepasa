package sipproxy

import (
	"fmt"
	"net"

	"github.com/pion/stun"
	log "github.com/sirupsen/logrus"
)

// STUNDiscovery handles public IP discovery via STUN servers (IPv4 only)
type STUNDiscovery struct {
	logger      *log.Entry
	stunServers []string
}

// NewSTUNDiscovery creates a new STUN discovery instance
func NewSTUNDiscovery(logger *log.Entry) *STUNDiscovery {
	return &STUNDiscovery{
		logger: logger,
		stunServers: []string{
			"stun1.l.google.com:19302",
			"stun2.l.google.com:19302",
			"stun3.l.google.com:19302",
			"stun4.l.google.com:19302",
		},
	}
}

// DiscoverPublicIPv4 discovers the public IPv4 address using STUN servers
func (sd *STUNDiscovery) DiscoverPublicIPv4() (string, error) {
	for _, stunServer := range sd.stunServers {
		sd.logger.Infof("🔍 Trying STUN server: %s (IPv4 only)", stunServer)

		// Verificar se o servidor STUN tem endereço IPv4
		host, port, err := net.SplitHostPort(stunServer)
		if err != nil {
			sd.logger.Warnf("❌ Invalid STUN server format %s: %v", stunServer, err)
			continue
		}

		// Resolver apenas endereços IPv4
		ips, err := net.LookupIP(host)
		if err != nil {
			sd.logger.Warnf("❌ Failed to resolve STUN server %s: %v", stunServer, err)
			continue
		}

		var ipv4Addr net.IP
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				ipv4Addr = ipv4
				break
			}
		}

		if ipv4Addr == nil {
			sd.logger.Warnf("❌ STUN server %s has no IPv4 address, skipping", stunServer)
			continue
		}

		stunServerIPv4 := net.JoinHostPort(ipv4Addr.String(), port)
		sd.logger.Infof("🌐 Using IPv4 address for STUN server: %s", stunServerIPv4)

		// Forçar conexão IPv4 usando "udp4" com endereço IPv4 resolvido
		conn, err := net.Dial("udp4", stunServerIPv4)
		if err != nil {
			sd.logger.Warnf("❌ Failed to connect to STUN server %s: %v", stunServerIPv4, err)
			continue
		}
		defer conn.Close()

		// Verificar se a conexão local também é IPv4
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		if localAddr.IP.To4() == nil {
			sd.logger.Warnf("❌ Local connection is not IPv4, skipping STUN server %s", stunServer)
			conn.Close()
			continue
		}
		sd.logger.Infof("📍 Local IPv4 address: %s", localAddr.IP.String())

		// Criar cliente STUN
		client, err := stun.NewClient(conn)
		if err != nil {
			sd.logger.Warnf("❌ Failed to create STUN client for %s: %v", stunServerIPv4, err)
			continue
		}
		defer client.Close()

		// Fazer request STUN
		var publicAddr *stun.XORMappedAddress
		err = client.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
			if res.Error != nil {
				sd.logger.Warnf("❌ STUN response error from %s: %v", stunServerIPv4, res.Error)
				return
			}

			var xorAddr stun.XORMappedAddress
			if err := xorAddr.GetFrom(res.Message); err == nil {
				sd.logger.Infof("📡 STUN returned address: %s", xorAddr.IP.String())
				publicAddr = &xorAddr
			}
		})

		if err != nil {
			sd.logger.Warnf("❌ STUN request failed for %s: %v", stunServerIPv4, err)
			continue
		}

		if publicAddr != nil {
			// Verificar rigorosamente se é IPv4
			if ipv4 := publicAddr.IP.To4(); ipv4 != nil && len(ipv4) == 4 {
				// Validar que não é um endereço IPv6 mapeado
				if publicAddr.IP.To16() != nil && publicAddr.IP.String() != ipv4.String() {
					sd.logger.Warnf("❌ STUN server %s returned IPv6-mapped address %s, skipping", stunServer, publicAddr.IP.String())
					continue
				}

				publicIP := ipv4.String()
				sd.logger.Infof("✅ Successfully discovered public IPv4: %s via STUN server %s", publicIP, stunServer)
				return publicIP, nil
			} else {
				sd.logger.Warnf("❌ STUN server %s returned non-IPv4 address %s (type: %T), skipping", stunServer, publicAddr.IP.String(), publicAddr.IP)
				continue
			}
		} else {
			sd.logger.Warnf("❌ STUN server %s returned no address", stunServer)
		}
	}

	return "", fmt.Errorf("failed to discover public IPv4 address via STUN")
}
