package whatsmeow

import (
	"fmt"

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
		logentry.Info("ℹ️ SIP Proxy desabilitado - não inicializando gerenciador de chamadas")
	}

	return manager
}

// initializeSIPProxy inicializa o singleton SIP proxy
func (scm *WhatsmeowSIPCallManager) initializeSIPProxy() {
	// Obter referência ao singleton (já deve estar inicializado)
	scm.sipProxyManager = sipproxy.SIPProxy

	// Verificar se o singleton foi inicializado corretamente
	if scm.sipProxyManager == nil {
		scm.logger.Error("❌ Singleton SIP proxy não foi inicializado")
		return
	}

	// Verificar se o proxy já está rodando
	if !scm.sipProxyManager.IsRunning() {
		scm.logger.Info("▶️ Iniciando SIP proxy manager...")
		if err := scm.sipProxyManager.Start(); err != nil {
			scm.logger.Errorf("❌ Falha ao iniciar SIP proxy: %v", err)
			return
		}
		scm.logger.Info("✅ SIP proxy manager iniciado com sucesso")
	} else {
		scm.logger.Info("✅ SIP proxy manager já está rodando - reutilizando instância")
	}

	// Configure callbacks
	scm.setupCallbacks()

	scm.logger.Info("✅ SIP proxy manager interno inicializado com sucesso")
}

// setupCallbacks configura os callbacks do SIP proxy
func (scm *WhatsmeowSIPCallManager) setupCallbacks() {
	scm.logger.Info("📞 WhatsmeowSIPCallManager: Configurando callbacks via sip_proxy_integration.go...")

	// Criar e configurar o SIP proxy integration
	scm.logger.Info("🔍 [DEBUG] Criando SIPProxyIntegration...")
	sipIntegration := NewSIPProxyIntegration(scm.logger)
	if sipIntegration != nil {
		scm.logger.Info("✅ [DEBUG] SIPProxyIntegration criado com sucesso")
		sipIntegration.SetupCallbacks(scm.connection)
		scm.logger.Info("✅ Callbacks configurados via sip_proxy_integration.go")

		// ✅ FIX: Configurar a integração SIP no WhatsmeowCallManager
		scm.logger.Info("🔍 [DEBUG] Obtendo CallManager da conexão...")
		if callManager := scm.connection.GetCallManager(); callManager != nil {
			scm.logger.Info("✅ [DEBUG] CallManager obtido com sucesso")
			callManager.SetSIPIntegration(sipIntegration)
			scm.logger.Info("✅ SIP integration configurada no CallManager para call termination")
		} else {
			scm.logger.Warn("⚠️ CallManager não disponível para configurar SIP integration")
		}
	} else {
		scm.logger.Error("❌ Falha ao criar SIPProxyIntegration para configurar callbacks")
	}
}

// IsEnabled verifica se o SIP call manager está habilitado
func (scm *WhatsmeowSIPCallManager) IsEnabled() bool {
	return scm.enabled
}

// ProcessCall processa uma chamada WhatsApp recebida via SIP proxy
func (scm *WhatsmeowSIPCallManager) ProcessCall(callID, fromUser, toUser string) error {
	if !scm.enabled {
		return fmt.Errorf("SIP call manager is disabled")
	}

	if scm.sipProxyManager == nil {
		return fmt.Errorf("SIP proxy manager not initialized")
	}

	scm.logger.Infof("🔄 Processando chamada WhatsApp recebida: %s → %s (CallID: %s)", fromUser, toUser, callID)

	// Enviar SIP INVITE usando o manager interno
	if err := scm.sipProxyManager.SendSIPInvite(callID, fromUser, toUser); err != nil {
		scm.logger.Errorf("❌ Falha ao enviar SIP INVITE: %v", err)
		return err
	}

	scm.logger.Info("✅ Call processed via internal SIP manager")
	return nil
}

// ProcessIncomingCall é um alias para ProcessCall para compatibilidade
func (scm *WhatsmeowSIPCallManager) ProcessIncomingCall(callID, fromUser, toUser string) error {
	return scm.ProcessCall(callID, fromUser, toUser)
}

// Shutdown encerra o call manager
func (scm *WhatsmeowSIPCallManager) Shutdown() error {
	if scm.sipProxyManager != nil && scm.sipProxyManager.IsRunning() {
		scm.logger.Info("🔻 Encerrando SIP proxy manager...")
		return scm.sipProxyManager.Stop()
	}
	return nil
}
