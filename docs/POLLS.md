# Enquetes (Polls) no QuePasa

Este documento descreve como usar as funcionalidades de enquetes implementadas no QuePasa, incluindo criação de enquetes e processamento de votos descriptografados.

## Funcionalidades Implementadas

### 1. Criação de Enquetes
- ✅ Criação de enquetes com pergunta e opções
- ✅ Suporte a múltiplas seleções
- ✅ Formatação rica das mensagens de enquete
- ✅ Webhooks para enquetes criadas

### 2. Processamento de Votos
- ✅ **Descriptografia automática de votos** usando a API do whatsmeow
- ✅ Identificação das opções selecionadas pelo usuário
- ✅ Exibição das opções votadas em texto claro
- ✅ Fallback para dados criptografados quando a descriptografia falha
- ✅ Webhooks detalhados para votos recebidos

## Como Usar

### Enviando uma Enquete

#### 1. Voto Recebido (Descriptografado)
```json
{
  "id": "3FC371B5ECCA4B677C6E",
  "timestamp": "2025-07-30T13:38:09.4287404-03:00",
  "type": "poll",
  "chat": {
    "id": "xxxxxx@s.whatsapp.net",
    "phone": "+55xxxxxxxxx",
    "title": "xxxxxxx",
    "lid": "xxxxxxxxxx@lid"
  },
  "text": "🗳️ *Voto registrado*\n\n📊 **Qual é sua linguagem favorita?**\n\n👤 xxxxxxx votou\n\n✅ *Opções selecionadas:*\n• Python\n",
  "fromme": false,
  "poll": {
    "question": "Qual é sua linguagem favorita?",
    "options": ["JavaScript", "Python", "Go", "TypeScript"],
    "selections": 1,
    "message_id": "3FA719CEB7BF1208F234"
  },
  "debug": {
    "event": "poll_vote",
    "reason": "vote_decrypted",
    "info": {
      "poll_vote": {
        "poll_id": "3FA719CEB7BF1208F234",
        "voter_id": "xxxxxxxxxx@s.whatsapp.net",
        "voter_name": "xxxxxxx",
        "voted_at": "2025-07-30T16:38:09Z",
        "selected_options": ["Python"],
        "encrypted_payload": "6NhEo8dSA95j1BnsnGNuu...",
        "encrypted_iv": "fP3Vjm3PSpZQOQcg"
      },
      "decryption_successful": true,
      "original_poll_found": true
    }
  },
  "action": "vote",
  "decryption_successful": true
}
```

#### 2. Voto Recebido (Criptografado - Fallback)
```json
{
  "id": "3FC371B5ECCA4B677C6E",
  "timestamp": "2025-07-30T13:38:09.4287404-03:00",
  "type": "poll",
  "text": "🗳️ *Voto registrado*\n\n📊 **Qual é sua linguagem favorita?**\n\n👤 xxxxx votou\n\n🔒 _Voto criptografado (não foi possível descriptografar)_\n\n_Dados criptografados:_\nPayload: 6NhEo8dSA95j1BnsnGNuu...\nIV: fP3Vjm3PSpZQOQcg",
  "debug": {
    "event": "poll_vote",
    "reason": "vote_encrypted",
    "info": {
      "decryption_successful": false
    }
  },
  "action": "vote",
  "decryption_successful": false
}
```

## Implementação Técnica

### Descriptografia de Votos

A implementação usa a função `DecryptPollVote` do whatsmeow para descriptografar automaticamente os votos:

1. **Captura do voto criptografado**: O sistema recebe `encPayload` e `encIV`
2. **Descriptografia automática**: Usa `client.DecryptPollVote()` 
3. **Mapeamento de hashes**: Converte hashes SHA-256 de volta para nomes das opções
4. **Fallback gracioso**: Se a descriptografia falha, mostra dados criptografados

### Tipos de Dados

#### WhatsappPoll
```go
type WhatsappPoll struct {
    Question   string   `json:"question"`
    Options    []string `json:"options"`
    Selections uint     `json:"selections"`
    // ... outros campos
}
```

#### WhatsappPollVote
```go
type WhatsappPollVote struct {
    PollId           string    `json:"poll_id"`
    VoterId          string    `json:"voter_id"`
    VoterName        string    `json:"voter_name"`
    VotedAt          time.Time `json:"voted_at"`
    SelectedOptions  []string  `json:"selected_options"`
    EncryptedPayload string    `json:"encrypted_payload"`
    EncryptedIV      string    `json:"encrypted_iv"`
}
```

## Informações Necessárias da Enquete Original

Para descriptografar votos com sucesso, o sistema precisa ter acesso às seguintes informações da **enquete original**:

### 1. Dados Obrigatórios ✅
```go
type WhatsappPoll struct {
    Question   string   `json:"question"`   // ✅ OBRIGATÓRIO: Pergunta da enquete
    Options    []string `json:"options"`    // ✅ OBRIGATÓRIO: Lista exata das opções
    MessageId  string   `json:"message_id"` // ✅ OBRIGATÓRIO: ID da mensagem original
}
```

### 2. Como Funciona a Descriptografia

O processo de descriptografia segue estes passos:

1. **Recebe voto criptografado**: `encPayload` + `encIV`
2. **Busca enquete original**: Usa `pollCreationMessageKey.ID` para encontrar a enquete
3. **Descriptografia pelo whatsmeow**: `client.DecryptPollVote()` retorna hashes SHA-256
4. **Mapeamento de opções**: Converte hashes de volta para nomes das opções

### 3. Exemplo do Mapeamento
```go
// Para cada opção da enquete original:
options := []string{"JavaScript", "Python", "Go", "TypeScript"}

// O whatsmeow retorna hashes SHA-256:
selectedHashes := [][]byte{0x1a2b3c...} // Hash do "Python"

// Nosso código mapeia de volta:
for _, option := range options {
    optionHash := sha256.Sum256([]byte(option))
    if bytes.Equal(selectedHash, optionHash[:]) {
        selectedOptions = append(selectedOptions, option) // "Python"
    }
}
```

### 4. Cache de Enquetes

O sistema mantém um cache interno das enquetes criadas:
- **Chave**: `message_id` da enquete original
- **Valor**: Estrutura `WhatsappPoll` completa
- **TTL**: Baseado nas configurações do whatsmeow

### 5. Cenários de Falha

❌ **Descriptografia falha quando**:
- Enquete original não está no cache
- Opções da enquete foram modificadas
- Chaves de criptografia foram perdidas
- Enquete é muito antiga (> TTL do cache)

✅ **Descriptografia funciona quando**:
- Enquete original está em cache
- Todas as opções originais estão disponíveis
- Message ID corresponde exatamente
- Chaves de criptografia estão válidas

## Limitações e Considerações

1. **Cache de enquetes**: Sistema depende do cache interno do whatsmeow para manter enquetes ativas
2. **Descriptografia**: Requer acesso às chaves de criptografia e à enquete original completa  
3. **Histórico**: Votos antigos podem falhar se a enquete original saiu do cache
4. **Webhook timing**: Delay entre criação e primeiro voto pode causar falhas temporárias
5. **Mapeamento exato**: As opções devem corresponder **exatamente** às originais (case-sensitive)

## Logs e Debug

Para debugar problemas de enquetes, verifique os logs:

```bash
# Criação de enquetes
grep "poll created" /var/log/quepasa.log

# Processamento de votos
grep "poll vote processed" /var/log/quepasa.log

# Tentativas de descriptografia
grep "decrypt poll vote" /var/log/quepasa.log

# Falhas de descriptografia
grep "failed to decrypt poll vote" /var/log/quepasa.log

# Cache de enquetes
grep "original poll found" /var/log/quepasa.log
```

## Troubleshooting

### ❓ Voto não descriptografa
```json
{
  "debug": {
    "reason": "vote_encrypted", 
    "decryption_successful": false
  }
}
```
**Possíveis causas:**
- Enquete original não está no cache
- Message ID não corresponde
- Enquete muito antiga
- Chaves de criptografia perdidas

**Solução:** Verifique logs para `"original poll found": false`

### ❓ Opções não mapeiam corretamente
```json
{
  "selected_options": [],
  "decryption_successful": false
}
```
**Possíveis causas:**
- Opções da enquete foram alteradas
- Hash SHA-256 não confere
- Encoding diferente das strings

**Solução:** Confirme que as opções são exatamente iguais às originais

### ❓ Cache de enquetes vazio
```bash
grep "handler or client is nil" /var/log/quepasa.log
```
**Possíveis causas:**
- Cliente whatsmeow não inicializado
- Perda de conexão durante criação
- Restart do serviço

**Solução:** Reinicie a conexão WhatsApp ou recrie a enquete

