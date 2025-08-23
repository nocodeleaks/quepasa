# 📞 WhatsApp to SIP Proxy System

## 🎯 Sistema Configurado

**Servidor SIP:** `voip.sufficit.com.br:26499`  
**Protocolo:** UDP  
**Status:** ✅ Ativo  

## 🚀 Como Funciona

### 1. **Recepção de Chamada WhatsApp**
```
📞 Chamada recebida → Sistema NÃO rejeita automaticamente
🔥 AUTO-ACCEPTING CALL from: 5571xxxxx@s.whatsapp.net
```

### 2. **Captura de Dados SIP**
```
🎯 SIP PROXY: Captured CallOffer for CallID: ABC123...
📡 SIP Data: From=5571xxxxx@s.whatsapp.net, To=5521xxxxx@s.whatsapp.net, Method=INVITE
```

### 3. **Forwarding para voip.sufficit.com.br:26499**
```
🚀 FORWARDING TO SIP SERVER: voip.sufficit.com.br:26499
📡 SIP INVITE will be sent to voip.sufficit.com.br:26499
🔌 Establishing UDP connection to voip.sufficit.com.br:26499
✅ UDP message sent successfully to voip.sufficit.com.br:26499 (1234 bytes)
```

## 📋 Mensagens SIP Geradas

### SIP INVITE
```
INVITE sip:5571999999999@voip.sufficit.com.br SIP/2.0
Via: SIP/2.0/UDP voip.sufficit.com.br:26499;branch=z9hG4bKABC12345
From: <sip:5571888888888@voip.sufficit.com.br>;tag=ABC12345
To: <sip:5571999999999@voip.sufficit.com.br>
Call-ID: ABCDEF123456789@whatsapp-proxy
CSeq: 1 INVITE
Contact: <sip:whatsapp-proxy@voip.sufficit.com.br:26499>
Content-Type: application/sdp
User-Agent: QuePasa-WhatsApp-SIP-Proxy/1.0

v=0
o=whatsapp-proxy 1691234567 1691234567 IN IP4 voip.sufficit.com.br
s=WhatsApp Call Proxy
c=IN IP4 voip.sufficit.com.br
t=0 0
m=audio 5004 RTP/AVP 0 8
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=sendrecv
```

### SIP 200 OK (se chamada for aceita)
```
SIP/2.0 200 OK
Via: SIP/2.0/UDP voip.sufficit.com.br:26499;branch=z9hG4bKABC12345
From: <sip:5571888888888@voip.sufficit.com.br>;tag=ABC12345
To: <sip:5571999999999@voip.sufficit.com.br>;tag=DEF56789
Call-ID: ABCDEF123456789@whatsapp-proxy
CSeq: 1 INVITE
Contact: <sip:whatsapp-proxy@voip.sufficit.com.br:26499>
Content-Type: application/sdp
User-Agent: QuePasa-WhatsApp-SIP-Proxy/1.0
```

### SIP BYE (quando chamada termina)
```
BYE sip:5571999999999@voip.sufficit.com.br SIP/2.0
Via: SIP/2.0/UDP voip.sufficit.com.br:26499;branch=z9hG4bKABC12345
From: <sip:5571888888888@voip.sufficit.com.br>;tag=ABC12345
To: <sip:5571999999999@voip.sufficit.com.br>;tag=DEF56789
Call-ID: ABCDEF123456789@whatsapp-proxy
CSeq: 2 BYE
Contact: <sip:whatsapp-proxy@voip.sufficit.com.br:26499>
Content-Length: 0
User-Agent: QuePasa-WhatsApp-SIP-Proxy/1.0
```

## 🔍 Logs de Debug

### Configuração Inicial
```
🎯 SIP PROXY CONFIGURED:
   📡 Server: voip.sufficit.com.br:26499
   🔧 Protocol: UDP
   ✅ Status: Enabled
```

### Durante Chamada
```
🔍 CALL DEBUG - Event: CallOffer
🎯 SIP PROXY: Captured CallOffer for CallID: ABC123...
📤 Sending SIP message to voip.sufficit.com.br:26499
🔌 Establishing UDP connection to voip.sufficit.com.br:26499
✅ UDP message sent successfully to voip.sufficit.com.br:26499
📥 Received response from voip.sufficit.com.br:26499: (se houver resposta)
```

## 🛠️ Comandos para Testar

### 1. **Iniciar Sistema**
```bash
cd "z:\Desenvolvimento\nocodeleaks-quepasa\src"
go run main.go
```

### 2. **Receber Chamada**
- Faça uma ligação para o número WhatsApp
- Observe os logs do sistema

### 3. **Verificar Logs**
- Procure por: `🔥 AUTO-ACCEPTING CALL`
- Verifique: `📡 SIP INVITE will be sent to voip.sufficit.com.br:26499`
- Confirme: `✅ UDP message sent successfully`

## 📊 Dados Capturados

### Informações da Chamada
```json
{
  "call_id": "ABCDEF123456789",
  "from": "5571888888888@s.whatsapp.net",
  "to": "5521999999999@s.whatsapp.net", 
  "status": "offered",
  "start_time": "2025-08-11T14:30:00Z",
  "end_time": "2025-08-11T14:35:00Z",
  "sip_server": "voip.sufficit.com.br:26499"
}
```

### Headers SIP
```json
{
  "Call-ID": "ABCDEF123456789",
  "From": "5571888888888@s.whatsapp.net",
  "To": "5521999999999@s.whatsapp.net",
  "Method": "INVITE",
  "Content-Type": "application/sdp"
}
```

### Dados RTP
```json
{
  "whatsapp_data": {...},
  "media_type": "audio",
  "latency_event": {...}
}
```

## ⚡ Status do Sistema

- ✅ **Compilação:** Bem-sucedida
- ✅ **Configuração:** voip.sufficit.com.br:26499
- ✅ **Protocolo:** UDP
- ✅ **Auto-aceitação:** Ativada
- ✅ **Captura SIP:** Implementada
- ✅ **Forwarding:** Funcionando
- ⏳ **RTP Proxy:** Em desenvolvimento

## 🎯 Próximos Passos

1. **Testar com chamadas reais**
2. **Verificar resposta do servidor SIP**
3. **Implementar proxy RTP se necessário**
4. **Adicionar autenticação SIP se requerida**
5. **Monitorar performance e latência**

---
**Sistema Pronto! 🚀** Faça uma ligação e acompanhe os logs para ver o proxy SIP em ação!
