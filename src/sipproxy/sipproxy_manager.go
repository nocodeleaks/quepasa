package sipproxy

import (
	"fmt"
	"sync"

	qplog "github.com/nocodeleaks/quepasa/qplog"
)

// SIPProxyManager coordena os módulos refatorados usando sipgo
type SIPProxyManager struct {
	mutex     sync.RWMutex
	logger    qplog.Logger
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
		logentry := qplog.New().WithField("package", "sipproxy")

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

		logentry.Infof("🏗️ SIP Proxy Manager inicializado com arquitetura modular usando sipgo")
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

// SetLocalRTPPort registers the local UDP RTP port to advertise in the SDP
// offer for callID. The audio bridge calls this with the port of the socket it
// listens on (and sends from), so the SIP server's RTP reaches the bridge.
// Must be called before SendSIPInvite for that call.
func (m *SIPProxyManager) SetLocalRTPPort(callID string, port int) {
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetLocalRTPPort(callID, port)
	}
}

// GetRemoteRTPAddr returns the SIP server's RTP address ("ip:port") parsed from
// the 200 OK SDP answer for callID, once the call has been accepted.
func (m *SIPProxyManager) GetRemoteRTPAddr(callID string) (string, bool) {
	if m.callManagerSipgo != nil {
		return m.callManagerSipgo.GetRemoteRTPAddr(callID)
	}
	return "", false
}

// HangupCall sends a SIP BYE to the server to tear down the call leg. It is
// called when the WhatsApp side ends so the SIP server (asterisk) hangs up too.
// Safe to call for an already-removed call (no-op).
func (m *SIPProxyManager) HangupCall(callID string) {
	if m.callManagerSipgo == nil {
		return
	}
	if err := m.callManagerSipgo.CancelCall(callID); err != nil {
		m.logger.Errorf("HangupCall: failed to send SIP BYE for %s: %v", callID, err)
	}
}

// SetCallAcceptedHandler define o callback para chamadas aceitas
func (m *SIPProxyManager) SetCallAcceptedHandler(handler SIPCallAcceptedCallback) {
	m.logger.Infof("📞 Configurando handler para chamadas aceitas")
	m.transactionMonitor.SetCallbacks(handler, m.transactionMonitor.callRejectedHandler)

	// NOVO: Configurar o callback também no sipgo call manager
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallAcceptedHandler(handler)
		m.logger.Infof("✅ Handler de aceitação também configurado no sipgo call manager")
	}
}

// SetCallRejectedHandler define o callback para chamadas rejeitadas
func (m *SIPProxyManager) SetCallRejectedHandler(handler SIPCallRejectedCallback) {
	m.logger.Infof("❌ Configurando handler para chamadas rejeitadas")
	m.transactionMonitor.SetCallbacks(m.transactionMonitor.callAcceptedHandler, handler)

	// NOVO: Configurar o callback também no sipgo call manager
	if m.callManagerSipgo != nil {
		m.callManagerSipgo.SetCallRejectedHandler(handler)
		m.logger.Infof("❌ Handler de rejeição também configurado no sipgo call manager")
	}
}

// Start inicializa e inicia o SIP proxy manager
func (m *SIPProxyManager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("SIP proxy manager já está rodando")
	}

	m.logger.Infof("🚀 Iniciando SIP Proxy Manager com arquitetura modular...")

	// Configurar rede (descoberta STUN, etc.)
	if err := m.networkManager.ConfigureNetwork(); err != nil {
		return fmt.Errorf("falha ao configurar rede: %v", err)
	}

	m.isRunning = true
	m.logger.Infof("✅ SIP Proxy Manager iniciado com sucesso")

	return nil
}

// Stop para graciosamente o SIP proxy manager
func (m *SIPProxyManager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("SIP proxy manager não está rodando")
	}

	m.logger.Infof("🛑 Parando SIP Proxy Manager...")

	// Cancelar todas as chamadas ativas
	for _, callID := range m.callManagerSipgo.GetActiveCalls() {
		if err := m.callManagerSipgo.CancelCall(callID); err != nil {
			m.logger.Errorf("Falha ao cancelar chamada %s: %v", callID, err)
		}
	}

	m.isRunning = false
	m.logger.Infof("✅ SIP Proxy Manager parado com sucesso")

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
	m.logger.Infof("📋 Configuração SIP Proxy atualizada")
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
	m.logger.Infof("🔧 Inicializando SIP Proxy Manager...")
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

// GetSipgoCallManager returns the underlying SIPGO call manager for advanced operations
// like per-call handler registration
func (m *SIPProxyManager) GetSipgoCallManager() *SIPCallManagerSipgo {
	return m.callManagerSipgo
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

// BridgeInboundWhatsAppCall handles an incoming WhatsApp VoIP call by:
//  1. Creating a raw RTP stream (two UDP sockets, no automatic forwarders)
//  2. Storing call state for tracking
//
// The SIP INVITE to the configured SIP server is sent separately by the
// SIPCallManagerSipgo. This method only prepares the RTP media path so the
// VoIPBridge can read/write μ-law RTP packets directly.
//
// The caller receives the *RTPStream handle and owns all I/O on its sockets.
func (m *SIPProxyManager) BridgeInboundWhatsAppCall(
	callID string,
	fromPhone string,
	toPhone string,
) (*RTPStream, error) {

	m.logger.Infof("🌉 BridgeInboundWhatsAppCall: CallID=%s, From=%s, To=%s", callID, fromPhone, toPhone)

	// Get local and public IPs from the network manager
	localIP := m.networkManager.GetLocalIP()
	if localIP == "" {
		localIP = "0.0.0.0"
	}
	publicIP := m.networkManager.GetPublicIP()

	// Create a dedicated RTPProxy instance for this bridge call.
	// We use a fresh instance because the bridge owns the sockets and does
	// not share the legacy forwarding model.
	baseLogger := qplog.New()
	rtpProxy := NewRTPProxy(baseLogger, localIP, publicIP)

	// The SIP server host/port come from SIPProxyManager config.
	sipHost := m.config.ServerHost
	sipPort := m.config.ServerPort
	if sipHost == "" {
		return nil, fmt.Errorf("BridgeInboundWhatsAppCall: SIP server host not configured")
	}

	// Create raw RTP stream (no forwarding goroutines).
	stream, err := rtpProxy.CreateRTPStreamRaw(callID, sipHost, sipPort)
	if err != nil {
		return nil, fmt.Errorf("BridgeInboundWhatsAppCall: failed to create RTP stream: %v", err)
	}

	// Store call data for tracking.
	callData := &SIPProxyCallData{
		CallID:    callID,
		From:      fromPhone,
		To:        toPhone,
		Status:    "initiated",
	}

	m.mutex.Lock()
	m.activeCalls[callID] = callData
	m.mutex.Unlock()

	m.logger.Infof("🌉✅ Bridge RTP stream ready: CallID=%s, WAPort=%d, SIPPort=%d → %s:%d",
		callID, stream.WhatsAppPort, stream.SIPPort, sipHost, sipPort)

	return stream, nil
}
