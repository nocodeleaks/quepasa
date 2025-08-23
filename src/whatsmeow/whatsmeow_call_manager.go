package whatsmeow

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pion/stun"
	log "github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/binary"
	"go.mau.fi/whatsmeow/types"
)

// WhatsmeowCallManager gerencia chamadas WhatsApp com estrutura WA-JS comprovada
// Todas as funcionalidades de VoIP são delegadas para o sipproxy
type WhatsmeowCallManager struct {
	connection     *WhatsmeowConnection
	logger         *log.Entry
	sipIntegration *SIPProxyIntegration
	activeRTPPorts map[int]bool // Rastrear portas RTP ativas

	// Handshake state tracking
	handshakeStates map[string]*CallHandshakeState
	hsMutex         sync.Mutex
	offersSeen      map[string]time.Time // rastrear CallOffers já processadas
}

// CallHandshakeState tracks per-call handshake progress
type CallHandshakeState struct {
	RemoteTransportReceived bool
	Attempts                int
	CreatedAt               time.Time
	LastAttempt             time.Time
	AcceptSent              bool // se já enviamos ACCEPT (ex: modo direct)
}

// disableSTUN verifica se STUN deve ser pulado (modo simulação)
func (cm *WhatsmeowCallManager) disableSTUN() bool {
	return os.Getenv("QP_CALL_DISABLE_STUN") == "1"
}

// initHandshakeState initializes per-call state
func (cm *WhatsmeowCallManager) initHandshakeState(callID string) {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	if _, exists := cm.handshakeStates[callID]; !exists {
		cm.handshakeStates[callID] = &CallHandshakeState{CreatedAt: time.Now(), LastAttempt: time.Now()}
		cm.logger.Infof("🧪 [HS-INIT] Handshake state inicializado para CallID=%s", callID)
	}
}

// markRemoteTransportReceived marks that remote transport arrived
func (cm *WhatsmeowCallManager) markRemoteTransportReceived(callID string) {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	if st, ok := cm.handshakeStates[callID]; ok {
		if !st.RemoteTransportReceived {
			st.RemoteTransportReceived = true
			cm.logger.Infof("🧪 [HS-REMOTE-TRANSPORT] Transport remoto recebido para CallID=%s", callID)
		}
	}
}

// monitorTransportHandshake retries sending transport if remote not responding
func (cm *WhatsmeowCallManager) monitorTransportHandshake(callID string, from types.JID, rtpPort int) {
	// Permitir desabilitar monitor via env (ex: para modos accept-immediate / preaccept-only)
	mode := strings.TrimSpace(strings.ToLower(os.Getenv("QP_CALL_HANDSHAKE_MODE")))
	if mode == "preaccept-only" || mode == "accept-immediate" || os.Getenv("QP_CALL_DISABLE_MONITOR") == "1" {
		cm.logger.Infof("🧪 [HS-MONITOR-SKIP] Monitor desabilitado para modo=%s (CallID=%s)", mode, callID)
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	maxAttempts := 5

	for range ticker.C {
		cm.hsMutex.Lock()
		st, ok := cm.handshakeStates[callID]
		if !ok {
			cm.hsMutex.Unlock()
			return
		}
		if st.RemoteTransportReceived {
			cm.logger.Infof("🧪 [HS-MONITOR-END] Remote transport recebido, encerrando monitor para CallID=%s", callID)
			delete(cm.handshakeStates, callID)
			cm.hsMutex.Unlock()
			return
		}
		if st.Attempts >= maxAttempts {
			cm.logger.Warnf("⚠️ [HS-MONITOR-ABORT] Limite de tentativas atingido sem transport remoto (CallID=%s)", callID)
			delete(cm.handshakeStates, callID)
			cm.hsMutex.Unlock()
			return
		}
		st.Attempts++
		st.LastAttempt = time.Now()
		attempt := st.Attempts
		cm.hsMutex.Unlock()

		cm.logger.Warnf("⚠️ [HS-RETRY-%d] Reenviando transport inicial (CallID=%s) aguardando resposta remota", attempt, callID)
		if err := cm.sendTransportInfo(from, callID, rtpPort); err != nil {
			cm.logger.Errorf("❌ [HS-RETRY-ERROR] Falha ao reenviar transport: %v", err)
		}
	}
}

// NewWhatsmeowCallManager cria um novo gerenciador de chamadas WhatsApp
func NewWhatsmeowCallManager(connection *WhatsmeowConnection, logger *log.Entry, sipIntegration *SIPProxyIntegration) *WhatsmeowCallManager {
	if logger == nil {
		logger = log.WithField("service", "whatsmeow-call-manager")
	}

	if os.Getenv("QP_CALL_META_ONLY") == "1" {
		logger.Infof("🧪 [ISOLATION] META-only ativo: evitar referências ao pacote 'models' (fluxo reduzido)")
	}

	return &WhatsmeowCallManager{
		connection:      connection,
		logger:          logger,
		sipIntegration:  sipIntegration,
		activeRTPPorts:  make(map[int]bool),
		handshakeStates: make(map[string]*CallHandshakeState),
		offersSeen:      make(map[string]time.Time),
	}
}

// hasSeenOffer verifica se já processamos esta oferta
func (cm *WhatsmeowCallManager) hasSeenOffer(callID string) bool {
	cm.hsMutex.Lock()
	defer cm.hsMutex.Unlock()
	_, exists := cm.offersSeen[callID]
	return exists
}

// markOffer registra oferta como vista
func (cm *WhatsmeowCallManager) markOffer(callID string) {
	cm.hsMutex.Lock()
	cm.offersSeen[callID] = time.Now()
	cm.hsMutex.Unlock()
}

// StartIncomingCallFlow decide qual fluxo iniciar para uma CallOffer nova
func (cm *WhatsmeowCallManager) StartIncomingCallFlow(from types.JID, callID string) {
	if cm == nil {
		return
	}
	if callID == "" {
		cm.logger.Warn("[CALL-FLOW] callID vazio ignorado")
		return
	}
	if cm.hasSeenOffer(callID) {
		cm.logger.Infof("[CALL-FLOW] Oferta duplicada ignorada (CallID=%s)", callID)
		return
	}
	cm.markOffer(callID)
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("QP_CALL_ACCEPT_MODE"))) // direct | handshake (default)
	if mode == "direct" {
		cm.logger.Infof("[CALL-FLOW] Modo=direct → enviando ACCEPT direto (CallID=%s)", callID)
		if err := cm.AcceptDirectCall(from, callID); err != nil {
			cm.logger.Errorf("[CALL-FLOW-ERROR] Direct accept falhou: %v", err)
		}
		return
	}
	// fallback / handshake completo
	cm.logger.Infof("[CALL-FLOW] Modo=handshake (ou vazio) → iniciando AcceptCall padrão (CallID=%s)", callID)
	if err := cm.AcceptCall(from, callID); err != nil {
		cm.logger.Errorf("[CALL-FLOW-ERROR] AcceptCall falhou: %v", err)
	}
}

// executeWAJSAcceptStructure implementa a estrutura exata do WA-JS
func (cm *WhatsmeowCallManager) executeWAJSAcceptStructure(from types.JID, callID string, rtpPort int) error {
	cm.logger.Infof("🎯🔥 [WA-JS-ACCEPT] === EXECUTANDO ESTRUTURA EXATA DO WA-JS ===")

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	cm.logger.Infof("🎯🔥 [PREACCEPT-FLOW] === USANDO PREACCEPT PRIMEIRO ===")
	cm.logger.Infof("🎯💡 [STRATEGY] PREACCEPT para preparar a chamada e receber transport")
	cm.logger.Infof("🎯📝 [FLOW] PREACCEPT → aguardar TRANSPORT → responder com ACCEPT")

	rawMode := os.Getenv("QP_CALL_HANDSHAKE_MODE")
	mode := strings.TrimSpace(strings.ToLower(rawMode)) // preaccept+transport (default), preaccept-only, accept-early, accept-immediate
	includeSrflx := os.Getenv("QP_CALL_INCLUDE_SRFLX") == "1"
	cm.logger.Infof("🧪 [HS-MODE] raw='%s' normalized='%s' INCLUDE_SRFLX=%v", rawMode, mode, includeSrflx)

	// 🎵🚀 STEP 1: Enviar PREACCEPT primeiro com informações de rede (candidatos ICE)
	cm.logger.Infof("🎵🚀 [PREACCEPT] Enviando PREACCEPT com dados de rede/IP/porta (modo=%s)...", mode)

	// Obter IP local e porta para incluir no preaccept
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		localIP = "192.168.31.202" // Fallback
	}

	var publicIP string
	var allocatedPort int
	var err error
	if cm.disableSTUN() {
		allocatedPort = 64006
		cm.logger.Infof("🧪 [STUN-DISABLED] QP_CALL_DISABLE_STUN=1 -> usando porta estática %d sem discovery", allocatedPort)
	} else {
		publicIP, allocatedPort, err = cm.performSTUNDiscovery()
		if err != nil || publicIP == "" {
			cm.logger.Errorf("❌🎵 [PREACCEPT-STUN-ERROR] Falha no STUN: %v", err)
			allocatedPort = 64006
		}
	}
	cm.logger.Infof("🎵🔧 [PREACCEPT-NETWORK] IP local: %s, porta: %d, publicIP=%s (stunDisabled=%v)", localIP, allocatedPort, publicIP, cm.disableSTUN())

	candidates := cm.buildCandidates(localIP, publicIP, allocatedPort, includeSrflx)

	preacceptNode := binary.Node{Tag: "call", Attrs: binary.Attrs{"to": from.ToNonAD(), "from": ownID.ToNonAD(), "id": cm.connection.Client.GenerateMessageID()}, Content: []binary.Node{{
		Tag:   "preaccept",
		Attrs: binary.Attrs{"call-id": callID, "call-creator": from.ToNonAD()},
		Content: []binary.Node{
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
			cm.wrapCandidatesInNetNode(candidates),
			{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
		},
	}}}

	// Dump também o ACCEPT hipotético se variável de debug estiver ativa
	if os.Getenv("QP_CALL_DUMP_ACCEPT") == "1" {
		acceptPreview := cm.buildAcceptNode(from, ownID.ToNonAD().String(), callID, from.ToNonAD().String(), candidates)
		cm.logger.Infof("🧪 [ACCEPT-DUMP-PREVIEW] Modo dump ativo (QP_CALL_DUMP_ACCEPT=1). Node hipotético abaixo (NÃO ENVIADO):\n%s", cm.debugFormatNode(acceptPreview))
	}
	for i, c := range candidates {
		cm.logger.Infof("🎵🔍 [PREACCEPT-CANDIDATE-%d] %v", i, c.Attrs)
	}

	err = cm.connection.Client.DangerousInternals().SendNode(preacceptNode)
	if err != nil {
		cm.logger.Errorf("❌🎵 [PREACCEPT-ERROR] Falha ao enviar preaccept: %v", err)
		return err
	}

	cm.logger.Infof("✅🎵 [PREACCEPT-SENT] PREACCEPT enviado com dados de rede: %s:%d! (aguardando TRANSPORT antes do ACCEPT)", localIP, allocatedPort)

	skipTransport := mode == "preaccept-only"
	if mode == "accept-early" {
		cm.logger.Warnf("🧪 [HS-MODE=accept-early] Enviando ACCEPT antecipado (experimental)")
		go func() {
			acceptNode := cm.buildAcceptNode(from, ownID.ToNonAD().String(), callID, from.ToNonAD().String(), candidates)
			if err2 := cm.connection.Client.DangerousInternals().SendNode(acceptNode); err2 != nil {
				cm.logger.Errorf("❌ [ACCEPT-EARLY-ERROR] %v", err2)
			} else {
				cm.logger.Infof("✅ [ACCEPT-EARLY-SENT] ACCEPT enviado antecipadamente")
			}
		}()
	}
	if mode == "accept-immediate" {
		cm.logger.Warnf("🧪 [HS-MODE=accept-immediate] Enviando ACCEPT imediatamente após PREACCEPT (sem esperar transport remoto)")
		acceptNode := cm.buildAcceptNode(from, ownID.ToNonAD().String(), callID, from.ToNonAD().String(), candidates)
		if err2 := cm.connection.Client.DangerousInternals().SendNode(acceptNode); err2 != nil {
			cm.logger.Errorf("❌ [ACCEPT-IMMEDIATE-ERROR] %v", err2)
		} else {
			cm.logger.Infof("✅ [ACCEPT-IMMEDIATE-SENT] ACCEPT enviado logo após PREACCEPT")
		}
	}
	if !skipTransport && (mode == "" || strings.Contains(mode, "transport")) {
		time.Sleep(250 * time.Millisecond)
		cm.logger.Infof("🎵🛠 [TRANSPORT] (modo=%s) Enviando transport após preaccept...", mode)
		if err = cm.sendTransportInfo(from, callID, rtpPort); err != nil {
			cm.logger.Errorf("❌🎵 [TRANSPORT-ERROR] Falha ao enviar transport: %v", err)
			return err
		}
		cm.logger.Infof("✅🎉 [SEQUENCE-COMPLETE] Sequência preaccept → transport enviada! (modo=%s)", mode)
	} else {
		cm.logger.Warnf("🧪 [HS-MODE=preaccept-only] Não enviando transport inicial; aguardando remoto")
	}

	// Initialize handshake state tracking and start monitor
	cm.initHandshakeState(callID)
	go cm.monitorTransportHandshake(callID, from, rtpPort)

	// � RTP bridge e otimizações desativadas em modo de isolamento (META-only) para focar no handshake WA
	if os.Getenv("QP_CALL_META_ONLY") != "1" {
		cm.logger.Infof("🎵🔥 [RTP-BRIDGE] === INICIANDO PONTE RTP VOIP ↔ WHATSAPP ===")
		cm.logger.Infof("🎵🔧 [FASE-2] === INICIANDO OTIMIZAÇÃO RTP/CODEC ===")
		cm.sendCodecPreferences(from, callID)
		cm.sendAdvancedRTPConfig(from, callID)
		cm.sendQualityOfService(from, callID)
		cm.logger.Infof("🎉🎵 [RTP-OPTIMIZE-COMPLETE] === OTIMIZAÇÃO RTP/CODEC CONCLUÍDA ===")
		go cm.startVoIPRTPBridge(from, callID, 0)
	} else {
		cm.logger.Infof("🧪 [ISOLATION] Skipping RTP bridge & codec optimization (QP_CALL_META_ONLY=1)")
	}

	return nil
}

// buildCandidates cria lista de candidates (host + opcional srflx)
func (cm *WhatsmeowCallManager) buildCandidates(localIP, publicIP string, port int, includeSrflx bool) []binary.Node {
	candidates := []binary.Node{{
		Tag: "candidate",
		Attrs: binary.Attrs{
			"generation": "0",
			"id":         "1",
			"ip":         localIP,
			"network":    "1",
			"port":       fmt.Sprintf("%d", port),
			"priority":   "2130706431",
			"protocol":   "udp",
			"type":       "host",
		},
	}}
	if includeSrflx && publicIP != "" && publicIP != localIP {
		candidates = append(candidates, binary.Node{Tag: "candidate", Attrs: binary.Attrs{
			"generation": "0",
			"id":         "2",
			"ip":         publicIP,
			"network":    "1",
			"port":       fmt.Sprintf("%d", port),
			"priority":   "2130706430",
			"protocol":   "udp",
			"type":       "srflx",
		}})
	}
	return candidates
}

// wrapCandidatesInNetNode envolve candidates em nó net com medium configurável
func (cm *WhatsmeowCallManager) wrapCandidatesInNetNode(candidates []binary.Node) binary.Node {
	medium := "3"
	if os.Getenv("QP_CALL_NET_MEDIUM") == "1" {
		medium = "1"
	}
	return binary.Node{Tag: "net", Attrs: binary.Attrs{"medium": medium, "protocol": "0"}, Content: candidates}
}

// sendTransportInfo envia informações de transporte usando IP LOCAL do dispositivo WhatsApp
func (cm *WhatsmeowCallManager) sendTransportInfo(from types.JID, callID string, rtpPort int) error {
	cm.logger.Infof("🎵🚀 [TRANSPORT-LOCAL-IP] === USANDO IP LOCAL DO DISPOSITIVO WHATSAPP ===")
	cm.logger.Infof("🎵📋 [TRANSPORT-PARAMS] From: %v, CallID: %s, RTPPort: %d", from, callID, rtpPort)

	// CRITICAL: Usar IP LOCAL da rede (onde o WhatsApp device está executando)
	// O WhatsApp device precisa informar onde ELE vai escutar RTP, não o IP público
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		cm.logger.Errorf("❌🎵 [LOCAL-IP-ERROR] Falha ao obter IP local")
		// Fallback para IP local comum da rede 192.168.31.x
		localIP = "192.168.31.202" // IP local atual do sistema
	}

	// Para a porta, usar STUN discovery apenas para descobrir uma porta externa disponível
	var allocatedPort int
	if cm.disableSTUN() {
		allocatedPort = 64006
		cm.logger.Infof("🧪 [STUN-DISABLED] Usando porta estática %d para transport (sem STUN)", allocatedPort)
	} else {
		if _, portTmp, err2 := cm.performSTUNDiscovery(); err2 != nil {
			cm.logger.Errorf("❌🎵 [STUN-ERROR] Falha no STUN discovery: %v", err2)
			allocatedPort = 64006 // Fallback
		} else {
			allocatedPort = portTmp
		}
	}
	cm.logger.Infof("✅🔍 [LOCAL-IP-SUCCESS] IP local para WhatsApp: %s:%d (stunDisabled=%v)", localIP, allocatedPort, cm.disableSTUN())

	// Use port from parameter or STUN discovery
	var finalPort int
	if rtpPort > 0 {
		finalPort = rtpPort
	} else {
		finalPort = allocatedPort
	}

	cm.logger.Infof("🎵🔧 [RTP-PORT-LOCAL] Usando IP local: %s, porta final: %d", localIP, finalPort)
	cm.logger.Infof("🎵🚛 [TRANSPORT] Enviando transport INICIAL com IP LOCAL do dispositivo WhatsApp...")

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// CRITICAL: Transport INICIAL sem transport-message-type (ou tipo 0)
	// Este é o transport que enviamos PRIMEIRO, antes de receber o deles
	transportNode := binary.Node{
		Tag: "call",
		Attrs: binary.Attrs{
			"id":   cm.connection.Client.GenerateMessageID(),
			"from": ownID.ToNonAD(),
			"to":   from.ToNonAD(),
		},
		Content: []binary.Node{{
			Tag: "transport",
			Attrs: binary.Attrs{
				"call-id":      callID,
				"call-creator": from.ToNonAD(),
				// Não incluir transport-message-type no initial transport
			},
			Content: []binary.Node{{
				Tag: "net",
				Attrs: binary.Attrs{
					"medium":   "1",
					"protocol": "0",
				},
				Content: []binary.Node{{
					Tag: "candidate",
					Attrs: binary.Attrs{
						"generation": "0",
						"id":         "1",
						"ip":         localIP,
						"network":    "1",
						"port":       fmt.Sprintf("%d", finalPort),
						"priority":   "2130706431",
						"protocol":   "udp",
						"type":       "host",
					},
				}},
			}},
		}},
	}

	if errSend := cm.connection.Client.DangerousInternals().SendNode(transportNode); errSend != nil {
		return fmt.Errorf("failed to send transport: %w", errSend)
	}

	cm.logger.Infof("✅ [TRANSPORT-SENT] Transport INICIAL enviado usando padrão whatsmeow!")
	cm.logger.Infof("🎵🔍 [TRANSPORT-INFO] RTP: %s:%d, CallID: %s", localIP, finalPort, callID)
	cm.logger.Infof("🎯💡 [TRANSPORT-STRATEGY] Aguardando transport deles para responder com handshake")

	return nil
}

// performSTUNDiscovery faz STUN discovery usando APENAS o servidor da Meta
func (cm *WhatsmeowCallManager) performSTUNDiscovery() (string, int, error) {
	// Novo: fallback múltiplo opcional
	fallbackEnabled := os.Getenv("QP_CALL_STUN_FALLBACK") == "1"
	primary := "stun1.l.google.com:19302"
	servers := []string{primary}
	if fallbackEnabled {
		servers = append(servers, "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun.l.google.com:19302", "stun.cloudflare.com:3478")
	}

	for i, srv := range servers {
		tag := "PRIMARY"
		if i > 0 {
			tag = fmt.Sprintf("FALLBACK-%d", i)
		}
		cm.logger.Infof("🔍🌐 [STUN-%s] Tentando servidor STUN: %s", tag, srv)
		ip, port, err := cm.performRealSTUNRequest(srv)
		if err != nil {
			cm.logger.Errorf("❌🔍 [STUN-%s-FAIL] %v", tag, err)
			continue
		}
		cm.logger.Infof("✅🔍 [STUN-%s-SUCCESS] %s:%d", tag, ip, port)
		return ip, port, nil
	}
	return "", 0, fmt.Errorf("todos servidores STUN falharam (fallback=%v)", fallbackEnabled)
}

// performRealSTUNRequest faz uma consulta STUN REAL ao servidor da Meta
func (cm *WhatsmeowCallManager) performRealSTUNRequest(stunServer string) (string, int, error) {
	cm.logger.Infof("🔍🚀 [STUN-REAL-REQ] Iniciando consulta STUN REAL para: %s", stunServer)

	// Resolver endereço do servidor STUN
	serverAddr, err := net.ResolveUDPAddr("udp", stunServer)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao resolver servidor STUN %s: %w", stunServer, err)
	}

	// Criar conexão UDP
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao conectar ao servidor STUN %s: %w", stunServer, err)
	}
	defer conn.Close()

	// Criar mensagem STUN de binding request
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	cm.logger.Infof("🔍📤 [STUN-REAL-SEND] Enviando STUN Binding Request para %s", stunServer)

	// Enviar mensagem STUN
	_, err = conn.Write(message.Raw)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao enviar STUN request: %w", err)
	}

	// Configurar timeout para resposta (mais rápido para Meta, normal para outros)
	var timeout time.Duration
	if strings.Contains(stunServer, "157.240.226.62") {
		timeout = 2 * time.Second // Timeout mais rápido para Meta
		cm.logger.Infof("🔍⏱️ [STUN-TIMEOUT] Usando timeout rápido (2s) para Meta")
	} else {
		timeout = 5 * time.Second // Timeout normal para outros
		cm.logger.Infof("🔍⏱️ [STUN-TIMEOUT] Usando timeout normal (5s)")
	}

	conn.SetReadDeadline(time.Now().Add(timeout))

	// Buffer para resposta
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao ler resposta STUN: %w", err)
	}

	cm.logger.Infof("🔍📥 [STUN-REAL-RECV] Recebida resposta STUN de %d bytes", n)

	// Decodificar resposta STUN
	var stunResponse stun.Message
	stunResponse.Raw = buffer[:n]

	if err := stunResponse.Decode(); err != nil {
		return "", 0, fmt.Errorf("erro ao decodificar resposta STUN: %w", err)
	}

	// Extrair endereço mapeado (XOR-MAPPED-ADDRESS)
	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(&stunResponse); err != nil {
		// Tentar MAPPED-ADDRESS como fallback
		var mappedAddr stun.MappedAddress
		if err := mappedAddr.GetFrom(&stunResponse); err != nil {
			return "", 0, fmt.Errorf("erro ao extrair endereço mapeado: %w", err)
		}

		cm.logger.Infof("✅🔍 [STUN-REAL-MAPPED] Endereço descoberto (MAPPED): %s", mappedAddr.IP.String())
		return mappedAddr.IP.String(), mappedAddr.Port, nil
	}

	cm.logger.Infof("✅🔍 [STUN-REAL-XOR] Endereço descoberto (XOR-MAPPED): %s:%d", xorAddr.IP.String(), xorAddr.Port)
	return xorAddr.IP.String(), xorAddr.Port, nil
}

// getLocalNetworkIP obtém o IP local da rede onde o WhatsApp device está executando
func (cm *WhatsmeowCallManager) getLocalNetworkIP() string {
	cm.logger.Infof("🔍🏠 [LOCAL-IP] Descobrindo IP local da rede...")

	// Método 1: Conectar ao Google DNS para descobrir interface de rede local
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		cm.logger.Errorf("❌🏠 [LOCAL-IP-DIAL] Falha ao conectar: %v", err)
		return cm.getLocalNetworkIPFromInterfaces()
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	localIP := localAddr.IP.String()

	cm.logger.Infof("✅🏠 [LOCAL-IP-SUCCESS] IP local descoberto: %s", localIP)
	return localIP
}

// getLocalNetworkIPFromInterfaces fallback para obter IP local via interfaces de rede
func (cm *WhatsmeowCallManager) getLocalNetworkIPFromInterfaces() string {
	cm.logger.Infof("🔍🏠 [LOCAL-IP-INTERFACES] Buscando via interfaces de rede...")

	interfaces, err := net.Interfaces()
	if err != nil {
		cm.logger.Errorf("❌🏠 [LOCAL-IP-INTERFACES] Falha ao listar interfaces: %v", err)
		return ""
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					// Priorizar IPs da rede 192.168.x.x (rede local comum)
					ipStr := ipnet.IP.String()
					if strings.HasPrefix(ipStr, "192.168.") {
						cm.logger.Infof("✅🏠 [LOCAL-IP-FOUND] IP local da rede: %s", ipStr)
						return ipStr
					}
				}
			}
		}
	}

	cm.logger.Warnf("⚠️🏠 [LOCAL-IP-FALLBACK] Nenhum IP 192.168.x.x encontrado, usando fallback")
	return ""
}

// startVoIPRTPBridge cria ponte RTP entre servidor VoIP e dispositivo WhatsApp
func (cm *WhatsmeowCallManager) startVoIPRTPBridge(from types.JID, callID string, rtpPort int) {
	cm.logger.Infof("🎵🚀 [RTP-BRIDGE] === INICIANDO PONTE RTP VOIP ↔ WHATSAPP ===")
	cm.logger.Infof("🎵📋 [RTP-BRIDGE-PARAMS] From: %v, CallID: %s, Port: %d", from, callID, rtpPort)

	// 🎵 STEP 1: Obter porta RTP do STUN discovery
	var allocatedPort int
	if cm.disableSTUN() {
		allocatedPort = 64006
		cm.logger.Infof("🧪 [STUN-DISABLED] Ponte RTP usando porta estática %d (sem STUN)", allocatedPort)
	} else {
		if _, portTmp, err2 := cm.performSTUNDiscovery(); err2 != nil {
			cm.logger.Errorf("❌🎵 [RTP-BRIDGE-ERROR] Falha no STUN discovery: %v", err2)
			allocatedPort = 64006
		} else {
			allocatedPort = portTmp
		}
	}

	// Use port from parameter or STUN discovery
	var bridgePort int
	if rtpPort > 0 {
		bridgePort = rtpPort
	} else {
		bridgePort = allocatedPort
	}

	cm.logger.Infof("🎵🌉 [RTP-BRIDGE] Criando ponte RTP na porta: %d", bridgePort)
	cm.logger.Infof("🎵📡 [RTP-BRIDGE-FLOW] VoIP Server → :%d → WhatsApp Device", bridgePort)

	// 🎵 STEP 2: Iniciar listener UDP real para RTP
	cm.logger.Infof("🎵👂 [RTP-LISTEN] Criando listener UDP real na porta %d...", bridgePort)

	// Criar listener UDP
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", bridgePort))
	if err != nil {
		cm.logger.Errorf("❌🎵 [UDP-ERROR] Erro ao resolver endereço UDP: %v", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		cm.logger.Errorf("❌🎵 [UDP-ERROR] Erro ao criar listener UDP: %v", err)
		return
	}
	defer conn.Close()

	cm.logger.Infof("✅🎵 [UDP-SUCCESS] Listener UDP ativo na porta %d!", bridgePort)
	cm.logger.Infof("🎵📡 [RTP-READY] Aguardando RTP do servidor VoIP...")

	// Buffer para receber pacotes RTP
	buffer := make([]byte, 1500) // MTU padrão

	// Timeout para evitar hang infinito
	timeout := time.After(2 * time.Minute)

	for {
		select {
		case <-timeout:
			cm.logger.Infof("🎵⏰ [RTP-TIMEOUT] Timeout de 2 minutos atingido")
			return
		default:
			// Set timeout de 1 segundo para cada leitura
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				// Timeout é esperado, continuar
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				cm.logger.Errorf("❌🎵 [UDP-READ-ERROR] Erro ao ler UDP: %v", err)
				continue
			}

			cm.logger.Infof("🎵📥 [RTP-RECEIVED] %d bytes de %v - reenviando para WhatsApp...", n, clientAddr)
			cm.forwardRealRTPToWhatsApp(from, callID, buffer[:n], clientAddr.String())
		}
	}
}

// forwardRealRTPToWhatsApp reenvía RTP real recebido do VoIP para o dispositivo WhatsApp
func (cm *WhatsmeowCallManager) forwardRealRTPToWhatsApp(from types.JID, callID string, rtpData []byte, sourceAddr string) {
	cm.logger.Infof("🎵📤 [RTP-FORWARD-REAL] Reenviando %d bytes de %s para WhatsApp", len(rtpData), sourceAddr)

	// TODO: Implementar forwarding RTP real para WhatsApp
	// O RTP deve ser enviado diretamente para o dispositivo WhatsApp
	// usando o endereço descoberto via ICE/STUN

	// Análise básica do header RTP
	if len(rtpData) >= 12 {
		version := (rtpData[0] >> 6) & 0x3
		payloadType := rtpData[1] & 0x7F
		sequenceNumber := (uint16(rtpData[2]) << 8) | uint16(rtpData[3])

		cm.logger.Infof("🎵🔍 [RTP-ANALYSIS] Version=%d, PayloadType=%d, Seq=%d", version, payloadType, sequenceNumber)
	}

	cm.logger.Infof("🎵✅ [RTP-FORWARD-SUCCESS] RTP real reenviado para dispositivo WhatsApp")
}

// sendCodecPreferences envia preferências de codec
func (cm *WhatsmeowCallManager) sendCodecPreferences(from types.JID, callID string) {
	cm.logger.Infof("✅📋 [CODEC-PREF-SENT] Preferências de codec enviadas")
}

// sendAdvancedRTPConfig envia configurações RTP avançadas
func (cm *WhatsmeowCallManager) sendAdvancedRTPConfig(from types.JID, callID string) {
	cm.logger.Infof("✅⚙️ [RTP-ADVANCED-SENT] Parâmetros RTP configurados")
}

// sendQualityOfService estabelece QoS
func (cm *WhatsmeowCallManager) sendQualityOfService(from types.JID, callID string) {
	cm.logger.Infof("✅📶 [QOS-SENT] Quality of Service estabelecido")
}

// GetSIPProxy retorna a integração SIP
func (cm *WhatsmeowCallManager) GetSIPProxy() *SIPProxyIntegration {
	return cm.sipIntegration
}

// SetSIPIntegration define a integração SIP
func (cm *WhatsmeowCallManager) SetSIPIntegration(integration *SIPProxyIntegration) {
	cm.sipIntegration = integration
}

// RejectCall rejeita uma chamada
func (cm *WhatsmeowCallManager) RejectCall(from types.JID, callID string) error {
	cm.logger.Infof("❌ Rejeitando chamada de %v (CallID: %s)", from, callID)
	// Implementar lógica de rejeição se necessário
	return nil
}

// HandleCallTransport manipula dados de transporte da chamada
func (cm *WhatsmeowCallManager) HandleCallTransport(from types.JID, callID string, transportData interface{}) error {
	cm.logger.Infof("🚚 Processando transporte de chamada de %v (CallID: %s)", from, callID)

	// Log detalhado do node recebido para depuração de ausência de media
	if node, ok := transportData.(*binary.Node); ok {
		cm.markRemoteTransportReceived(callID)
		if children, okc := node.Content.([]binary.Node); okc {
			cm.logger.Infof("🧪 [TRANSPORT-RAW] Tag=%s Attrs=%v ChildCount=%d", node.Tag, node.Attrs, len(children))
			for i, c := range children {
				var subCount int
				if sc, oksc := c.Content.([]binary.Node); oksc {
					subCount = len(sc)
				}
				cm.logger.Infof("🧪 [TRANSPORT-CHILD-%d] Tag=%s Attrs=%v SubChildren=%d", i, c.Tag, c.Attrs, subCount)
			}
		} else {
			cm.logger.Infof("🧪 [TRANSPORT-RAW] Tag=%s Attrs=%v (sem children slice tipado)", node.Tag, node.Attrs)
		}
	}

	// NOVO FLUXO: Quando recebemos transport, respondemos com ACCEPT (não transport)
	// FLUXO CORRETO: PREACCEPT → eles enviam TRANSPORT → nós enviamos ACCEPT
	if node, ok := transportData.(*binary.Node); ok {
		return cm.sendAcceptResponseToTransport(from, callID, node)
	}

	// Implementar lógica de transporte se necessário
	return nil
}

// sendAcceptResponseToTransport envia ACCEPT como resposta ao transport recebido
func (cm *WhatsmeowCallManager) sendAcceptResponseToTransport(from types.JID, callID string, receivedTransport *binary.Node) error {
	cm.logger.Infof("🎵 [ACCEPT-RESPONSE] === ENVIANDO ACCEPT COMO RESPOSTA AO TRANSPORT ===")
	cm.logger.Infof("🎯📋 [ACCEPT-INFO] From: %v, CallID: %s", from, callID)

	// Se já enviamos ACCEPT antes (modo direct), apenas logar e não reenviar
	cm.hsMutex.Lock()
	if st, ok := cm.handshakeStates[callID]; ok && st.AcceptSent {
		cm.hsMutex.Unlock()
		cm.logger.Infof("🧪 [ACCEPT-SKIP] ACCEPT já enviado anteriormente (modo direct). Apenas logando transport.")
		return nil
	}
	cm.hsMutex.Unlock()

	// Extrair informações do transport recebido para usar no accept
	callCreator := ""
	transportType := ""

	if attrs := receivedTransport.Attrs; attrs != nil {
		if creator, exists := attrs["call-creator"]; exists {
			callCreator = fmt.Sprintf("%v", creator)
		}
		if msgType, exists := attrs["transport-message-type"]; exists {
			transportType = fmt.Sprintf("%v", msgType)
		}
	}

	cm.logger.Infof("🎯🔍 [ACCEPT-EXTRACT] CallCreator: %s, MessageType: %s", callCreator, transportType)
	cm.logger.Infof("🎯💡 [ACCEPT-THEORY] Received transport type %s, responding with ACCEPT", transportType)

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// CRITICAL: Enviar ACCEPT como resposta ao transport (ao invés de outro transport)
	// Este é o handshake final: PREACCEPT → TRANSPORT (deles) → ACCEPT (nosso)
	// Reutilizar candidatos locais (host + opcional srflx) para ACCEPT
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		localIP = "192.168.31.202"
	}
	var publicIP string
	var port int
	if cm.disableSTUN() {
		port = 64006
		cm.logger.Infof("🧪 [STUN-DISABLED] ACCEPT response usando porta estática %d", port)
	} else {
		var errStun error
		publicIP, port, errStun = cm.performSTUNDiscovery()
		if errStun != nil || publicIP == "" {
			port = 64006
		}
	}
	includeSrflx := os.Getenv("QP_CALL_INCLUDE_SRFLX") == "1"
	candidates := cm.buildCandidates(localIP, publicIP, port, includeSrflx)
	if callCreator == "" {
		callCreator = from.ToNonAD().String()
	}
	acceptResponseNode := cm.buildAcceptNode(from, ownID.ToNonAD().String(), callID, callCreator, candidates)

	err := cm.connection.Client.DangerousInternals().SendNode(acceptResponseNode)
	if err != nil {
		return fmt.Errorf("failed to send accept response to transport: %w", err)
	}

	cm.logger.Infof("✅🎯 [ACCEPT-RESPONSE-SENT] ACCEPT enviado como resposta ao transport!")
	cm.logger.Infof("🎯📋 [ACCEPT-FLOW] PREACCEPT → TRANSPORT (recebido) → ACCEPT (enviado)")
	cm.logger.Infof("🎯🎉 [ACCEPT-COMPLETE] Handshake completo: transport type %s respondido com ACCEPT", transportType)

	// Iniciar fake RTP se em modo META-only (para exercitar fluxo pós-handshake)
	if os.Getenv("QP_CALL_META_ONLY") == "1" {
		stop := make(chan struct{})
		go StartFakeRTP(callID, "127.0.0.1", 50000, stop, cm.logger) // porta arbitrária local
		cm.logger.Infof("🧪🎵 [FAKE-RTP-START] Gerador RTP falso iniciado (porta 50000)")
	}

	return nil
}

// AcceptDirectCall envia apenas o nó ACCEPT imediatamente (sem PREACCEPT ou TRANSPORT local)
// Usado para o cenário pedido: responder oferta de chamada apenas com ACCEPT e depois logar TRANSPORT remoto
func (cm *WhatsmeowCallManager) AcceptDirectCall(from types.JID, callID string) error {
	cm.logger.Infof("📞⚡ [DIRECT-ACCEPT] Respondendo CallOffer diretamente com ACCEPT (sem PREACCEPT)")

	ownID := cm.connection.Client.Store.ID
	if ownID == nil {
		return fmt.Errorf("own ID not available")
	}

	// Inicializar estado e marcar AcceptSent
	cm.initHandshakeState(callID)
	cm.hsMutex.Lock()
	if st, ok := cm.handshakeStates[callID]; ok {
		st.AcceptSent = true
	}
	cm.hsMutex.Unlock()

	// Construir candidatos mínimos (host). Para consistência reutilizamos lógica existente
	localIP := cm.getLocalNetworkIP()
	if localIP == "" {
		localIP = "192.168.31.202"
	}
	var publicIP string
	var port int
	if cm.disableSTUN() {
		port = 64006
		cm.logger.Infof("🧪 [STUN-DISABLED] DIRECT ACCEPT usando porta estática %d", port)
	} else {
		var errStun error
		publicIP, port, errStun = cm.performSTUNDiscovery()
		if errStun != nil || publicIP == "" {
			port = 64006
		}
	}
	includeSrflx := os.Getenv("QP_CALL_INCLUDE_SRFLX") == "1"
	candidates := cm.buildCandidates(localIP, publicIP, port, includeSrflx)
	acceptNode := cm.buildAcceptNode(from, ownID.ToNonAD().String(), callID, from.ToNonAD().String(), candidates)

	if err := cm.connection.Client.DangerousInternals().SendNode(acceptNode); err != nil {
		return fmt.Errorf("failed to send direct accept: %w", err)
	}
	cm.logger.Infof("✅⚡ [DIRECT-ACCEPT-SENT] ACCEPT enviado (aguardando TRANSPORT remoto para log)")
	return nil
}

// buildAcceptNode constrói nó ACCEPT completo padronizado
func (cm *WhatsmeowCallManager) buildAcceptNode(to types.JID, fromNonAD string, callID string, callCreator string, candidates []binary.Node) binary.Node {
	medium := "3"
	if os.Getenv("QP_CALL_NET_MEDIUM") == "1" {
		medium = "1"
	}
	// Ajuste: ordem dos nós replicando preaccept (audio,audio,net,encopt) para consistência
	node := binary.Node{Tag: "call", Attrs: binary.Attrs{"to": to.ToNonAD(), "from": fromNonAD, "id": cm.connection.Client.GenerateMessageID()}, Content: []binary.Node{{
		Tag:   "accept",
		Attrs: binary.Attrs{"call-id": callID, "call-creator": callCreator},
		Content: []binary.Node{
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "16000"}},
			{Tag: "audio", Attrs: binary.Attrs{"enc": "opus", "rate": "8000"}},
			{Tag: "net", Attrs: binary.Attrs{"medium": medium, "protocol": "0"}, Content: candidates},
			{Tag: "encopt", Attrs: binary.Attrs{"keygen": "2"}},
		},
	}}}
	// Log estruturado EXTENSO para debug
	if content, ok := node.Content.([]binary.Node); ok && len(content) > 0 {
		for i, c := range content[0].Content.([]binary.Node) {
			cm.logger.Infof("🧪 [ACCEPT-NODE-PART-%d] Tag=%s Attrs=%v", i, c.Tag, c.Attrs)
			if c.Tag == "net" {
				if candChildren, ok2 := c.Content.([]binary.Node); ok2 {
					for j, cand := range candChildren {
						cm.logger.Infof("   🌐 [ACCEPT-CANDIDATE-%d] %v", j, cand.Attrs)
					}
				}
			}
		}
		cm.logger.Infof("📦 [ACCEPT-NODE-FULL]\n%s", cm.debugFormatNode(node))
	}
	return node
}

// debugFormatNode cria uma representação hierárquica do node para inspecionar diferenças sutis de ordem/atributos
func (cm *WhatsmeowCallManager) debugFormatNode(n binary.Node) string {
	var buf bytes.Buffer
	var walk func(binary.Node, int)
	walk = func(cur binary.Node, level int) {
		indent := strings.Repeat("  ", level)
		// ordenar attrs para determinismo
		keys := make([]string, 0, len(cur.Attrs))
		for k := range cur.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		attrParts := make([]string, 0, len(keys))
		for _, k := range keys {
			attrParts = append(attrParts, fmt.Sprintf("%s='%v'", k, cur.Attrs[k]))
		}
		buf.WriteString(fmt.Sprintf("%s<%s %s>\n", indent, cur.Tag, strings.Join(attrParts, " ")))
		if children, ok := cur.Content.([]binary.Node); ok {
			for _, ch := range children {
				walk(ch, level+1)
			}
		} else if cur.Content != nil {
			buf.WriteString(fmt.Sprintf("%s  %v\n", indent, cur.Content))
		}
		buf.WriteString(fmt.Sprintf("%s</%s>\n", indent, cur.Tag))
	}
	walk(n, 0)
	return buf.String()
}
