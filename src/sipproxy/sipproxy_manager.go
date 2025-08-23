package sipproxy

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	logrus "github.com/sirupsen/logrus"
)

// SIPProxyManager coordena os módulos refatorados usando sipgo
type SIPProxyManager struct {
	mutex     sync.RWMutex
	logger    *log.Entry
	config    SIPProxySettings
	isRunning bool

	// Módulos refatorados
	networkManager     *SIPProxyNetworkManager
	responseHandler    *SIPResponseHandler
	callManagerSipgo   *SIPCallManagerSipgo // sipgo-based call manager
	transactionMonitor *SIPTransactionMonitor

	// Componentes legados (mantidos para compatibilidade)
	upnpManager *UPnPManager
	sipListener *SIPListener

	// Rastreamento de chamadas
	activeCalls  map[string]*SIPProxyCallData
	callAttempts map[string]int
}

var (
	managerInstance *SIPProxyManager
	managerOnce     sync.Once
)

// GetSIPProxyManager retorna a instância singleton do manager refatorado
func GetSIPProxyManager(settings SIPProxySettings) *SIPProxyManager {
	managerOnce.Do(func() {
		logentry := logrus.WithField("package", "sipproxy")

		// Inicializar componentes legados
		upnpManager := NewUPnPManager(logentry)
		sipListener := NewSIPListener(logentry)
		transactionMonitor := NewSIPTransactionMonitor(logentry)

		// Inicializar módulos refatorados
		networkManager := NewSIPProxyNetworkManager(settings.SIPProxyNetworkManagerSettings, logentry)

		responseHandler := NewSIPResponseHandler(
			logentry.WithField("module", "response"),
			transactionMonitor,
		)

		// NEW: Initialize sipgo-based call manager
		callManagerSipgo := NewSIPCallManagerSipgo(
			logentry.WithField("module", "sipgo-call"),
			settings,
			networkManager,
		)

		managerInstance = &SIPProxyManager{
			logger:             logentry,
			config:             settings,
			activeCalls:        make(map[string]*SIPProxyCallData),
			callAttempts:       make(map[string]int),
			networkManager:     networkManager,
			responseHandler:    responseHandler,
			callManagerSipgo:   callManagerSipgo,
			transactionMonitor: transactionMonitor,
			upnpManager:        upnpManager,
			sipListener:        sipListener,
		}

		logentry.Info("🏗️ SIP Proxy Manager inicializado com arquitetura modular usando sipgo")
	})
	return managerInstance
}

// SendSIPInvite inicia uma chamada SIP usando a arquitetura modular
func (m *SIPProxyManager) SendSIPInvite(callID, fromPhone, toPhone string) error {
	m.logger.Infof("🚀 DEBUG: SendSIPInvite recebeu parâmetros:")
	m.logger.Infof("   📞 CallID recebido: %s", callID)
	m.logger.Infof("   🔵 From recebido: %s", fromPhone)
	m.logger.Infof("   🟢 To recebido: %s", toPhone)

	m.logger.Infof("🆕📞 Iniciando chamada SIP modular usando SIPGO: %s → %s (CallID: %s)", fromPhone, toPhone, callID)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Rastrear tentativa de chamada
	m.callAttempts[callID]++
	attemptNumber := m.callAttempts[callID]

	m.logger.Infof("🔄 Tentativa #%d para CallID: %s", attemptNumber, callID)

	m.logger.Infof("🔍 DEBUG: Passando para InitiateCallSipgo (NEW):")
	m.logger.Infof("   📞 CallID que será passado: %s", callID)
	m.logger.Infof("   🔵 From que será passado: %s", fromPhone)
	m.logger.Infof("   🟢 To que será passado: %s", toPhone)

	// Usar o call manager sipgo para iniciar a chamada
	return m.callManagerSipgo.InitiateCallSipgo(callID, fromPhone, toPhone)
}

// SetCallAcceptedHandler define o callback para chamadas aceitas
func (m *SIPProxyManager) SetCallAcceptedHandler(handler SIPCallAcceptedCallback) {
	m.logger.Info("📞 Configurando handler para chamadas aceitas")
	m.transactionMonitor.SetCallbacks(handler, m.transactionMonitor.callRejectedHandler)

	// NOVO: Configurar o callback também no sipgo call manager
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallAcceptedHandler(handler)
		m.logger.Info("✅ Handler de aceitação também configurado no sipgo call manager")
	}
}

// SetCallRejectedHandler define o callback para chamadas rejeitadas
func (m *SIPProxyManager) SetCallRejectedHandler(handler SIPCallRejectedCallback) {
	m.logger.Info("❌ Configurando handler para chamadas rejeitadas")
	m.transactionMonitor.SetCallbacks(m.transactionMonitor.callAcceptedHandler, handler)

	// NOVO: Configurar o callback também no sipgo call manager
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallRejectedHandler(handler)
		m.logger.Info("❌ Handler de rejeição também configurado no sipgo call manager")
	}
}

// Start inicializa e inicia o SIP proxy manager
func (m *SIPProxyManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("SIP proxy manager já está rodando")
	}

	m.logger.Info("🚀 Iniciando SIP Proxy Manager com arquitetura modular...")

	// Configurar rede (descoberta STUN, etc.)
	if err := m.networkManager.ConfigureNetwork(); err != nil {
		return fmt.Errorf("falha ao configurar rede: %v", err)
	}

	m.isRunning = true
	m.logger.Info("✅ SIP Proxy Manager iniciado com sucesso")

	return nil
}

// Stop para graciosamente o SIP proxy manager
func (m *SIPProxyManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("SIP proxy manager não está rodando")
	}

	m.logger.Info("🛑 Parando SIP Proxy Manager...")

	// Cancelar todas as chamadas ativas
	for _, callID := range m.callManagerSipgo.GetActiveCalls() {
		if err := m.callManagerSipgo.CancelCall(callID); err != nil {
			m.logger.Errorf("Falha ao cancelar chamada %s: %v", callID, err)
		}
	}

	m.isRunning = false
	m.logger.Info("✅ SIP Proxy Manager parado com sucesso")

	return nil
}

// IsRunning retorna se o SIP proxy manager está rodando
func (m *SIPProxyManager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isRunning
}

// GetActiveCallCount retorna o número de chamadas ativas
func (m *SIPProxyManager) GetActiveCallCount() int {
	return len(m.callManagerSipgo.GetActiveCalls())
}

// GetCallState retorna o estado atual de uma chamada
func (m *SIPProxyManager) GetCallState(callID string) (string, bool) {
	// sipgo handles call state internally, for now return unknown
	return "UNKNOWN", false
}

// GetNetworkInfo retorna informações da configuração de rede atual
func (m *SIPProxyManager) GetNetworkInfo() map[string]interface{} {
	return map[string]interface{}{
		"public_ip":     m.networkManager.GetPublicIP(),
		"local_ip":      m.networkManager.GetLocalIP(),
		"local_port":    m.networkManager.GetLocalPort(),
		"sip_server":    m.networkManager.GetSIPServerEndpoint(),
		"is_configured": m.networkManager.IsConfigured(),
	}
}

// GetStats retorna estatísticas do manager
func (m *SIPProxyManager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"is_running":          m.isRunning,
		"active_calls":        m.GetActiveCallCount(),
		"total_call_attempts": len(m.callAttempts),
		"network_configured":  m.networkManager.IsConfigured(),
	}
}

// Métodos de compatibilidade legada
func (m *SIPProxyManager) GetConfig() SIPProxySettings {
	return m.config
}

func (m *SIPProxyManager) GetPublicIP() string {
	return m.networkManager.GetPublicIP()
}

func (m *SIPProxyManager) SetConfig(config SIPProxySettings) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.config = config
	m.logger.Info("📋 Configuração SIP Proxy atualizada")
}

// Métodos avançados de gerenciamento de chamadas
func (m *SIPProxyManager) CancelCall(callID string) error {
	return m.callManagerSipgo.CancelCall(callID)
}

func (m *SIPProxyManager) GetCallInfo(callID string) (map[string]interface{}, bool) {
	// sipgo handles call info internally, for now return basic info
	return map[string]interface{}{
		"call_id": callID,
		"state":   "UNKNOWN",
	}, false
}

// Initialize inicializa o SIP proxy manager (compatibilidade legada)
func (m *SIPProxyManager) Initialize() error {
	m.logger.Info("🔧 Inicializando SIP Proxy Manager...")
	return m.Start()
}

// RemoveCall remove uma chamada ativa (compatibilidade legada)
func (m *SIPProxyManager) RemoveCall(callID string) {
	m.logger.Infof("🗑️ Removendo chamada: %s", callID)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if call exists in active calls map
	if _, exists := m.activeCalls[callID]; exists {
		delete(m.activeCalls, callID)
		m.logger.Infof("✅ Chamada %s removida do mapeamento ativo", callID)
	} else {
		m.logger.Infof("📞ℹ️ Chamada %s não encontrada no mapeamento ativo - pode ter sido removida anteriormente", callID)
	}

	// =========================================================================
	// 🚫 DUPLICATE BYE PREVENTION: Don't send BYE automatically here
	// =========================================================================
	// The CancelCall() method already handles BYE sending and cleanup
	// RemoveCall() should only remove from tracking, not send SIP messages

	// COMMENTED OUT: Automatic CancelCall (causes duplicate BYEs)
	// Cancel the call in call manager (now handles missing calls gracefully)
	// if err := m.callManagerSipgo.CancelCall(callID); err != nil {
	//	m.logger.Errorf("Erro ao cancelar chamada %s: %v", callID, err)
	// } else {
	//	m.logger.Infof("✅ Processo de cancelamento de chamada %s concluído", callID)
	// }

	m.logger.Infof("✅ Chamada %s removida do rastreamento (sem envio de BYE adicional)", callID)
}

// GetActiveCalls retorna todas as chamadas ativas (compatibilidade legada)
func (m *SIPProxyManager) GetActiveCalls() map[string]*SIPProxyCallData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Retorna uma cópia do mapa para evitar modificações concorrentes
	activeCalls := make(map[string]*SIPProxyCallData)
	for callID, callData := range m.activeCalls {
		activeCalls[callID] = callData
	}

	return activeCalls
}
