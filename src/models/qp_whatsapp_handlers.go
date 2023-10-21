package models

import (
	"fmt"
	"strings"
	"sync"
	"time"

	whatsapp "github.com/nocodeleaks/quepasa/whatsapp"
)

// Serviço que controla os servidores / bots individuais do whatsapp
type QPWhatsappHandlers struct {
	server *QpWhatsappServer

	messages     map[string]whatsapp.WhatsappMessage
	sync         *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
	syncRegister *sync.Mutex

	// Appended events handler
	aeh []interface {
		Handle(*whatsapp.WhatsappMessage)
	}
}

func (handler *QPWhatsappHandlers) HandleGroups() bool {
	return handler.server.HandleGroups
}

func (handler *QPWhatsappHandlers) HandleBroadcast() bool {
	return handler.server.HandleBroadcast
}

//#region EVENTS FROM WHATSAPP SERVICE

func (handler *QPWhatsappHandlers) Message(msg *whatsapp.WhatsappMessage) {

	// skipping groups if choosed
	if !handler.HandleGroups() && msg.FromGroup() {
		return
	}

	// skipping broadcast if choosed
	if !handler.HandleBroadcast() && msg.FromBroadcast() {
		return
	}

	// messages sended with chat title
	if len(msg.Chat.Title) == 0 {
		msg.Chat.Title = handler.server.GetChatTitle(msg.Chat.Id)
	}

	if len(msg.InReply) > 0 {
		cached, err := handler.GetMessage(msg.InReply)
		if err == nil {
			maxlength := ENV.SynopsisLength() - 4
			if uint64(len(cached.Text)) > maxlength {
				msg.Synopsis = cached.Text[0:maxlength] + " ..."
			} else {
				msg.Synopsis = cached.Text
			}
		}
	}

	handler.server.Log.Tracef("msg recebida/(enviada por outro meio) em models: %s", msg.Id)
	handler.appendMsgToCache(msg)
}

// does not cache msg, only update status and webhook dispatch
func (handler *QPWhatsappHandlers) Receipt(msg *whatsapp.WhatsappMessage) {
	ids := strings.Split(msg.Text, ",")
	for _, element := range ids {
		cached, err := handler.GetMessage(element)
		if err == nil {
			// update status
			handler.server.Log.Tracef("msg recebida/(enviada por outro meio) em models: %s", cached.Id)
		}
	}

	// Executando WebHook de forma assincrona
	handler.Trigger(msg)
}

/*
<summary>

	Event on:
		* User Logged Out from whatsapp app
		* Maximum numbers of devices reached
		* Banned
		* Token Expired

</summary>
*/
func (handler *QPWhatsappHandlers) LoggedOut(reason string) {

	// one step at a time
	if handler.server != nil {

		msg := "logged out !"
		if len(reason) > 0 {
			msg += " reason: " + reason
		}

		handler.server.Log.Warn(msg)

		// marking unverified and wait for more analyses
		handler.server.MarkVerified(false)
	}
}

//#endregion
//region MESSAGE CONTROL REGION HANDLE A LOCK

// Salva em cache e inicia gatilhos assíncronos
func (handler *QPWhatsappHandlers) appendMsgToCache(msg *whatsapp.WhatsappMessage) {

	handler.sync.Lock() // Sinal vermelho para atividades simultâneas
	// Apartir deste ponto só se executa um por vez

	normalizedId := msg.Id
	normalizedId = strings.ToUpper(normalizedId) // ensure that is an uppercase string before save

	// saving on local normalized cache, do not afect remote msgs
	handler.messages[normalizedId] = *msg

	handler.sync.Unlock() // Sinal verde !

	// Executando WebHook de forma assincrona
	handler.Trigger(msg)
}

func (handler *QPWhatsappHandlers) GetMessages(timestamp time.Time) (messages []whatsapp.WhatsappMessage) {
	handler.sync.Lock() // Sinal vermelho para atividades simultâneas
	// Apartir deste ponto só se executa um por vez

	for _, item := range handler.messages {
		if item.Timestamp.After(timestamp) {
			messages = append(messages, item)
		}
	}

	handler.sync.Unlock() // Sinal verde !
	return
}

func (handler *QPWhatsappHandlers) GetMessagesByPrefix(id string) (messages []whatsapp.WhatsappMessage) {
	handler.sync.Lock() // Sinal vermelho para atividades simultâneas
	// Apartir deste ponto só se executa um por vez

	for _, item := range handler.messages {
		if strings.HasPrefix(item.Id, id) {
			messages = append(messages, item)
		}
	}

	handler.sync.Unlock() // Sinal verde !
	return
}

// Get a single message if exists
func (handler *QPWhatsappHandlers) GetMessage(id string) (msg whatsapp.WhatsappMessage, err error) {
	handler.sync.Lock() // Sinal vermelho para atividades simultâneas
	// Apartir deste ponto só se executa um por vez

	normalizedId := id
	normalizedId = strings.ToUpper(normalizedId) // ensure that is an uppercase string before save

	// getting from local normalized cache, do not afect remote msgs
	msg, ok := handler.messages[normalizedId]
	if !ok {
		err = fmt.Errorf("message not present on handlers (cache) id: %s", normalizedId)
	}

	handler.sync.Unlock() // Sinal verde !
	return msg, err
}

//endregion
//region EVENT HANDLER TO INTERNAL USE, GENERALY TO WEBHOOK

func (handler *QPWhatsappHandlers) Trigger(payload *whatsapp.WhatsappMessage) {
	for _, handler := range handler.aeh {
		go handler.Handle(payload)
	}
}

// Register an event handler that triggers on a new message received on cache
func (handler *QPWhatsappHandlers) Register(evt interface {
	Handle(*whatsapp.WhatsappMessage)
}) {
	handler.sync.Lock() // Sinal vermelho para atividades simultâneas

	if !handler.IsRegistered(evt) {
		handler.aeh = append(handler.aeh, evt)
	}

	handler.sync.Unlock()
}

// Removes an specific event handler
func (handler *QPWhatsappHandlers) UnRegister(evt interface {
	Handle(*whatsapp.WhatsappMessage)
}) {
	handler.sync.Lock() // Sinal vermelho para atividades simultâneas

	newHandlers := []interface {
		Handle(*whatsapp.WhatsappMessage)
	}{}
	for _, v := range handler.aeh {
		if v != evt {
			newHandlers = append(handler.aeh, evt)
		}
	}

	// updating
	handler.aeh = newHandlers

	handler.sync.Unlock()
}

// Removes an specific event handler
func (handler *QPWhatsappHandlers) Clear() {
	handler.sync.Lock() // Sinal vermelho para atividades simultâneas

	// updating
	handler.aeh = nil

	handler.sync.Unlock()
}

// Indicates that has any event handler registered
func (handler *QPWhatsappHandlers) IsAttached() bool {
	return len(handler.aeh) > 0
}

// Indicates that if an specific hanlder is registered
func (handler *QPWhatsappHandlers) IsRegistered(evt interface{}) bool {
	for _, v := range handler.aeh {
		if v == evt {
			return true
		}
	}

	return false
}

//endregion

func (handler *QPWhatsappHandlers) GetTotal() int {
	return len(handler.messages)
}
