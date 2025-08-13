package whatsmeow

import (
	"fmt"

	"github.com/emiago/sipgo/sip"
	environment "github.com/nocodeleaks/quepasa/environment"
	sipproxy "github.com/nocodeleaks/quepasa/sipproxy"
	logrus "github.com/sirupsen/logrus"
)

// WhatsmeowSIPCallManager gerencia chamadas SIP integradas ao Whatsmeow
type WhatsmeowSIPCallManager struct {
	logger          *logrus.Entry
	connection      *WhatsmeowConnection
	sipProxyManager *sipproxy.SIPProxyManager
	enabled         bool
}

// NewWhatsmeowSIPCallManager cria um novo gerenciador de chamadas SIP interno
func NewWhatsmeowSIPCallManager(conn *WhatsmeowConnection) *WhatsmeowSIPCallManager {
	logentry := conn.GetLogger().WithField("component", "sip_call_manager")

	manager := &WhatsmeowSIPCallManager{
		logger:     logentry,
		connection: conn,
		enabled:    environment.Settings.SIPProxy.Enabled,
	}

	// Só inicializa o SIP proxy se estiver habilitado
	if manager.enabled {
		logentry.Info("🔧 SIP Proxy habilitado - inicializando gerenciador interno de chamadas")
		manager.initializeSIPProxy()
	} else {
		logentry.Debug("🔇 SIP Proxy desabilitado - gerenciador de chamadas inativo")
	}

	return manager
}

// initializeSIPProxy initializes the internal SIP proxy manager
func (scm *WhatsmeowSIPCallManager) initializeSIPProxy() {

	// Get SIP proxy manager instance
	scm.sipProxyManager = sipproxy.SIPProxy

	if scm.sipProxyManager == nil {
		scm.logger.Error("❌ Falha ao obter instância do SIP proxy manager")
		scm.enabled = false
		return
	}

	// ✅ SIP proxy já foi inicializado no init() - apenas verificar se está rodando
	if !scm.sipProxyManager.IsRunning() {
		scm.logger.Warn("⚠️ SIP proxy manager não está rodando - tentando inicializar")
		if err := scm.sipProxyManager.Initialize(); err != nil {
			scm.logger.Errorf("❌ Falha ao inicializar SIP proxy: %v", err)
			scm.enabled = false
			return
		}
	} else {
		scm.logger.Info("✅ SIP proxy manager já está rodando - reutilizando instância")
	}

	// Configurar callbacks
	scm.setupCallbacks()

	scm.logger.Info("✅ SIP proxy manager interno inicializado com sucesso")
}

// setupCallbacks configura os callbacks do SIP proxy
func (scm *WhatsmeowSIPCallManager) setupCallbacks() {
	// Callback para chamadas aceitas
	scm.sipProxyManager.SetCallAcceptedHandler(func(callID, fromPhone, toPhone string, response *sip.Response) {
		scm.logger.Infof("📞✅ Chamada SIP aceita: %s → %s (CallID: %s)", fromPhone, toPhone, callID)
		// Aqui pode adicionar lógica adicional para chamadas aceitas
	})

	// Callback para chamadas rejeitadas
	scm.sipProxyManager.SetCallRejectedHandler(func(callID, fromPhone, toPhone string, response *sip.Response) {
		scm.logger.Infof("📞❌ Chamada SIP rejeitada: %s → %s (CallID: %s)", fromPhone, toPhone, callID)
		// Aqui pode adicionar lógica adicional para chamadas rejeitadas
	})
}

// IsEnabled verifica se o SIP call manager está habilitado
func (scm *WhatsmeowSIPCallManager) IsEnabled() bool {
	return scm.enabled
}

// ProcessIncomingCall processa uma chamada WhatsApp recebida
func (scm *WhatsmeowSIPCallManager) ProcessIncomingCall(callID, fromPhone, toPhone string) error {
	if !scm.enabled {
		return fmt.Errorf("SIP call manager não está habilitado")
	}

	if scm.sipProxyManager == nil {
		return fmt.Errorf("SIP proxy manager não foi inicializado")
	}

	scm.logger.Infof("🔄 Processando chamada WhatsApp recebida: %s → %s (CallID: %s)", fromPhone, toPhone, callID)

	// Iniciar chamada via SIP proxy usando o método correto
	return scm.sipProxyManager.SendSIPInvite(callID, fromPhone, toPhone)
}

// AcceptCall aceita uma chamada
func (scm *WhatsmeowSIPCallManager) AcceptCall(callID string) error {
	if !scm.enabled {
		return fmt.Errorf("SIP call manager não está habilitado")
	}

	scm.logger.Infof("📞✅ Aceitando chamada: %s", callID)
	// Lógica para aceitar a chamada pode ser implementada aqui
	return nil
}

// RejectCall rejeita uma chamada
func (scm *WhatsmeowSIPCallManager) RejectCall(callID, reason string) error {
	if !scm.enabled {
		return fmt.Errorf("SIP call manager não está habilitado")
	}

	scm.logger.Infof("📞❌ Rejeitando chamada: %s (Razão: %s)", callID, reason)

	if scm.sipProxyManager != nil {
		return scm.sipProxyManager.CancelCall(callID)
	}

	return nil
}

// CancelCall cancela uma chamada
func (scm *WhatsmeowSIPCallManager) CancelCall(callID string) error {
	if !scm.enabled {
		return fmt.Errorf("SIP call manager não está habilitado")
	}

	scm.logger.Infof("📞🚫 Cancelando chamada: %s", callID)

	if scm.sipProxyManager != nil {
		return scm.sipProxyManager.CancelCall(callID)
	}

	return nil
}

// Shutdown finaliza o SIP call manager
func (scm *WhatsmeowSIPCallManager) Shutdown() {
	if scm.enabled && scm.sipProxyManager != nil {
		scm.logger.Info("🛑 Finalizando SIP call manager interno")
		scm.sipProxyManager.Stop()
	}
}
