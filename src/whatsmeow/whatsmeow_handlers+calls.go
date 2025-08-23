// =============================================================================
// Arquivo: whatsmeow_handlers+calls.go
// Propósito:
//
//	Centralizar APENAS o tratamento BÁSICO (mínimo) dos eventos de chamadas
//	recebidos do WhatsApp (CallOffer, CallAccept, CallReject, Transport, etc.).
//	Aqui deve existir somente:
//	  - Logging padronizado de cada evento
//	  - Extração simples de campos essenciais (quando necessário)
//	  - Acionamento de rotinas triviais (ex: marcar estado, aceitar direto, fila leve)
//	NÃO colocar aqui lógica complexa, fluxos extensos, handshake detalhado,
//	montagem de nós (nodes) avançados, análise profunda de media ou integração
//	externa pesada. Qualquer processamento mais elaborado deve ser extraído
//	para arquivos/funções específicas (ex: call_accept_flow.go, call_transport_parser.go).
//
// Diretriz:
//
//	Mantenha este arquivo pequeno, legível e focado em orquestrar o básico.
//	Ele é a “camada fina” entre o evento bruto e módulos especializados.
//
// =============================================================================
package whatsmeow

import (
	"fmt"

	"github.com/nocodeleaks/quepasa/library"
	"go.mau.fi/whatsmeow/types/events"
)

func (source *WhatsmeowHandlers) HandleCallOffer(evt *events.CallOffer) {
	if source == nil || evt == nil {
		return
	}

	logentry := source.GetLogger()
	logentry.Infof("[CALL] Offer: from=%s callID=%s ts=%v", evt.From, evt.CallID, evt.Timestamp)

	callOffer := NewWhatsmeowCallOffer(evt)
	if !callOffer.IsValid() {
		logentry.Debugf("[CALL] Offer not valid (expired or not joinable): callID=%s", evt.CallID)
		return
	}

	if source.WhatsmeowConnection != nil {
		if cm := source.WhatsmeowConnection.GetCallManager(); cm != nil {
			cm.AcceptCall(evt.From, evt.CallID)
		} else {
			logentry.Debug("[CALL] CallManager indisponível no momento da oferta")
		}
	}
}

func (source *WhatsmeowHandlers) HandleCallOfferNotice(evt *events.CallOfferNotice) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] OfferNotice: from=%s callID=%s", evt.From, evt.CallID)
	// OfferNotice pode chegar antes/depois — apenas log; fluxo principal já tratado no Offer
}

func (source *WhatsmeowHandlers) HandleCallRelayLatency(evt *events.CallRelayLatency) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] RelayLatency: from=%s callID=%s data=%v", evt.From, evt.CallID, evt.Data)
}

func (source *WhatsmeowHandlers) HandleCallTerminate(evt *events.CallTerminate) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] Terminate: from=%s callID=%s reason=%v", evt.From, evt.CallID, evt.Reason)
}

func (source *WhatsmeowHandlers) HandleCallAccept(evt *events.CallAccept) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] Accept: from=%s callID=%s", evt.From, evt.CallID)
}

func (source *WhatsmeowHandlers) HandleCallReject(evt *events.CallReject) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] Reject: from=%s callID=%s", evt.From, evt.CallID)
}

func (source *WhatsmeowHandlers) HandleCallTransport(evt *events.CallTransport) {
	if source == nil || evt == nil {
		return
	}

	logentry := source.GetLogger()
	logentry.Infof("[CALL] Transport: from=%s callID=%s ts=%v", evt.From, evt.CallID, evt.Timestamp)

	callTransport := NewWhatsmeowCallTransport(evt)
	if !callTransport.IsValid() {
		logentry.Debugf("[CALL] Offer not valid (expired or not joinable): callID=%s", evt.CallID)
		return
	}

	json := library.ToJson(evt)
	logentry.Infof("[CALL] Transport JSON: %s", json)

	return

	// Log básico somente. Parsing profundo mover para outro arquivo (ex: call_transport_parse.go)
	size := 0
	if evt.Data != nil {
		size = len(fmt.Sprintf("%v", evt.Data))
	}
	logentry.Infof("[CALL] Transport: from=%s callID=%s dataSize=%d", evt.From, evt.CallID, size)
}

func (source *WhatsmeowHandlers) HandleCallUnknown(evt *events.UnknownCallEvent) {
	if source == nil || evt == nil {
		return
	}
	logentry := source.GetLogger()
	logentry.Infof("[CALL] UnknownCallEvent: raw=%+v", evt)
}
