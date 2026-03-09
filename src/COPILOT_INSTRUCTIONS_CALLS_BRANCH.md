# 🎯 COPILOT INSTRUCTIONS - CALLS BRANCH

## 📋 OBJETIVO PRINCIPAL DA BRANCH CALLS

**META:** Implementar um proxy de chamadas WhatsApp-SIP funcional que aceite automaticamente chamadas do WhatsApp e estabeleça o fluxo RTP de áudio com um servidor SIP.

### 🔥 OBJETIVO ESPECÍFICO ATUAL

**DESAFIO:** Fazer o WhatsApp **realmente aceitar** chamadas automaticamente e estabelecer o handshake para iniciar o tráfego de áudio RTP.

#### O QUE JÁ FUNCIONA ✅
- ✅ Detecção de chamadas WhatsApp (CallOffer)
- ✅ Envio para servidor SIP (voip.example.com:26499)
- ✅ Resposta 200 OK do servidor SIP
- ✅ Integração SIP configurada corretamente
- ✅ Terminação de chamadas (BYE/CANCEL) funcionando
- ✅ Sistema sem avisos/warnings

#### O QUE NÃO FUNCIONA ❌
- ❌ **WhatsApp continua tocando em todos os dispositivos** após "aceitação"
- ❌ Chamada não para de tocar (indica que WhatsApp não reconheceu a aceitação)
- ❌ Fluxo RTP de áudio não é estabelecido
- ❌ Handshake de aceitação não é confirmado pelo WhatsApp

### 🧠 CONTEXTO TÉCNICO

#### Limitações Conhecidas
- WhatsApp Business API **NÃO** tem método oficial `AcceptCall()`
- Existe apenas `RejectCall()` oficial
- WhatsApp Web **TEM** funcionalidade de aceitar chamadas
- Precisamos **descobrir e replicar** o comportamento do WhatsApp Web

#### Arquitetura Atual
```
WhatsApp CallOffer → SIP INVITE → SIP Server (200 OK) → Tentativa Accept → ❌ Falha
```

#### Arquitetura Desejada
```
WhatsApp CallOffer → SIP INVITE → SIP Server (200 OK) → WhatsApp Accept ✅ → RTP Audio Flow
```

### 🎯 ESTRATÉGIAS IMPLEMENTADAS

#### Múltiplas Abordagens de Aceitação
1. **STRATEGY 1:** Busca por método oficial `AcceptCall()` via reflection
2. **STRATEGY 2:** Protocolo manual com diferentes estruturas de nós
3. **STRATEGY 3:** Resposta com informações de transporte RTP
4. **STRATEGY 4:** Simulação do comportamento do WhatsApp Web

#### Estruturas de Nós Testadas
- `preaccept` + `accept` (atual)
- `accept` com atributo `count`
- `accept` com atributo `media`
- `transport` com informações RTP
- `media` com status ativo

### 🔬 PRÓXIMOS PASSOS

#### Investigação Necessária
- [ ] Analisar tráfego de rede do WhatsApp Web durante aceitação
- [ ] Testar diferentes sequências de nós (preaccept/accept/transport)
- [ ] Verificar se eventos `CallAccept` são disparados
- [ ] Experimentar com atributos adicionais nos nós

#### Melhorias de Debug
- [ ] Monitor de eventos `CallAccept` em tempo real
- [ ] Logs detalhados de todas as tentativas
- [ ] Análise de resposta do servidor WhatsApp

### 📁 ARQUIVOS PRINCIPAIS

#### Core da Funcionalidade
- `whatsmeow_call_manager.go` - Lógica principal de aceitação
- `sip_proxy_integration.go` - Integração com SIP
- `whatsmeow_handlers.go` - Handlers de eventos
- `whatsmeow_sip_call_manager.go` - Gerenciamento SIP-WhatsApp

#### Arquivos de Teste/Backup
- `call_accept_*.go` - Métodos experimentais
- `whatsmeow_call_answer_experimental.go` - Testes avançados

### 🚨 INSTRUÇÕES PARA COPILOT

#### Quando Retomar Esta Conversa:
1. **FOQUE** no objetivo: fazer WhatsApp aceitar chamadas e estabelecer RTP
2. **NÃO** sugira rejeitar chamadas - queremos aceitar!
3. **SEMPRE** teste as estratégias múltiplas implementadas
4. **MONITORE** se a chamada para de tocar nos outros dispositivos
5. **ANALISE** logs para eventos `CallAccept`

#### Contexto de Desenvolvimento:
- Branch: `calls`
- Ambiente: Windows + Go
- Servidor SIP: voip.example.com:26499
- Biblioteca: whatsmeow + sipgo

#### Comportamento Esperado:
```
📞 Chamada chega → 🔄 Estratégias de aceitação → ✅ Para de tocar → 🎵 Áudio RTP flui
```

---
**NOTA:** Este é um proxy SIP-WhatsApp. O objetivo é aceitar chamadas WhatsApp e roteá-las para SIP com áudio funcionando!
