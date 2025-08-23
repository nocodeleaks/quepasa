package sipproxy

import (
	environment "github.com/nocodeleaks/quepasa/environment"
	logrus "github.com/sirupsen/logrus"
)

var SIPProxy *SIPProxyManager

func init() {
	env := environment.Settings.SIPProxy

	// Initialize SIP Proxy if configured
	if env.Enabled {

		settings := GetEnvironmentSettings()

		logentry := logrus.WithField("package", "sipproxy")
		logentry.Infof("🔧 SIP Proxy enabled - Host: %s, Port: %d, Protocol: %s",
			settings.ServerHost,
			settings.ServerPort,
			settings.Protocol)

		// Initialize SIP Proxy singleton using environment settings directly
		SIPProxy = GetSIPProxyManager(settings)
		if SIPProxy == nil {
			logentry.Errorf("❌ Failed to get SIP proxy singleton")
		} else {
			// Initialize the SIP proxy
			err := SIPProxy.Initialize()
			if err != nil {
				logentry.Errorf("❌ Failed to initialize SIP proxy: %v", err)
			} else {
				logentry.Info("✅ SIP Proxy initialized successfully")
				logentry.Infof("📡 SIP Proxy listening on port %d, forwarding to %s:%d",
					env.LocalPort,
					env.Host,
					env.Port)
			}
		}
	}
}

// GetEnvironmentSettings retrieves the environment settings for the SIP proxy
func GetEnvironmentSettings() SIPProxySettings {
	env := environment.Settings.SIPProxy

	settings := SIPProxySettings{
		SIPProxyNetworkManagerSettings: SIPProxyNetworkManagerSettings{
			StunServer:       env.STUNServer,
			SIPServer:        env.Host,
			SIPPort:          int(env.Port),
			LocalPort:        int(env.LocalPort),
			PublicIP:         env.PublicIP,
			IsSTUNConfigured: len(env.STUNServer) > 0,
		},

		SDPSessionName: env.SDPSessionName,
		UserAgent:      env.UserAgent,
		ServerHost:     env.Host,
		ServerPort:     int(env.Port),
		ListenerPort:   int(env.LocalPort),
		Protocol:       env.Protocol,
	}

	return settings
}
