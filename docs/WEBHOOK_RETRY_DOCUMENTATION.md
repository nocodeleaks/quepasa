# Sistema de Retry de Webhooks - Documentação Oficial

## 📋 Índice
- [Visão Geral](#visão-geral)
- [Instalação e Configuração](#instalação-e-configuração)
- [Como Funciona](#como-funciona)
- [Configurações de Environment](#configurações-de-environment)
- [Exemplos Práticos](#exemplos-práticos)
- [Logs e Monitoramento](#logs-e-monitoramento)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)

---

## ⚠️ IMPORTANTE: Sistema Condicional

**O sistema de retry de webhooks é OPCIONAL e ativado apenas quando configurado:**

1. **SEM `WEBHOOK_RETRY_COUNT` no .env**: 
   - ✅ Usa comportamento original
   - ✅ Uma tentativa única
   - ✅ Compatível com sistemas existentes

2. **COM `WEBHOOK_RETRY_COUNT` no .env**:
   - ✅ Ativa sistema de retry automático
   - ✅ Múltiplas tentativas conforme configurado
   - ✅ Logs detalhados de retry

**Esta abordagem garante compatibilidade total com sistemas existentes.**

---

## 🎯 Visão Geral

O **Sistema de Retry de Webhooks** é uma funcionalidade que aumenta a confiabilidade do envio de mensagens para sistemas externos no QuePasa. Quando um webhook falha, o sistema automaticamente tenta reenviar a mensagem seguindo configurações personalizáveis.

### Principais Benefícios
- ✅ **Maior Confiabilidade**: Recuperação automática de falhas temporárias
- ✅ **Configurável**: Ajuste o comportamento por ambiente
- ✅ **Compatível**: Funciona com código existente sem mudanças
- ✅ **Observável**: Logs detalhados de todas as tentativas

---

## ⚙️ Instalação e Configuração

### 1. Variáveis de Environment

Adicione estas variáveis ao seu arquivo `.env`:

```bash
# Sistema de Retry de Webhooks
WEBHOOK_RETRY_COUNT=3    # Número de tentativas após falha inicial
WEBHOOK_RETRY_DELAY=1    # Segundos entre tentativas
WEBHOOK_TIMEOUT=10       # Timeout por requisição (segundos)
```

### 2. Valores Padrão
Quando `WEBHOOK_RETRY_COUNT` está definida, o sistema usa estes padrões para variáveis não configuradas:
- **WEBHOOK_RETRY_COUNT**: Valor definido pelo usuário
- **WEBHOOK_RETRY_DELAY**: 1 segundo (se não definido)
- **WEBHOOK_TIMEOUT**: 10 segundos (se não definido)

**Comportamento sem configuração:**
Se `WEBHOOK_RETRY_COUNT` não estiver definida, o sistema usa o comportamento original (uma tentativa apenas).

### 3. Ativação
O sistema de retry é **condicional** e é ativado apenas quando a variável `WEBHOOK_RETRY_COUNT` estiver definida no arquivo `.env`.

- **Sem `WEBHOOK_RETRY_COUNT` definida**: Usa comportamento original (uma tentativa apenas)
- **Com `WEBHOOK_RETRY_COUNT` definida**: Ativa o sistema de retry automático

---

## 🔄 Como Funciona

### Fluxo de Execução

```
1. Tentativa Inicial
   ↓
2. Falhou? → Aguarda delay → Retry
   ↓
3. Sucesso? → ✅ FIM
   ↓
4. Falhou? → Aguarda delay → Retry
   ↓
5. Esgotar tentativas? → ❌ ERRO FINAL
```

### Condições de Falha
- Status HTTP ≠ 200
- Timeout na requisição
- Erro de conexão/rede
- Erro na criação da requisição

### Condições de Sucesso
- Status HTTP = 200
- Resposta recebida dentro do timeout

---

## 🛠️ Configurações de Environment

### WEBHOOK_RETRY_COUNT
```bash
# Número de retries após primeira falha
WEBHOOK_RETRY_COUNT=3

# Exemplo: 3 = 4 tentativas totais
# - 1 tentativa inicial
# - 3 tentativas de retry
```

**Valores Recomendados:**
- **Desenvolvimento**: `1` (rápido para testes)
- **Produção**: `3` (balanceado)
- **Alta Disponibilidade**: `5` (máxima confiabilidade)

### WEBHOOK_RETRY_DELAY
```bash
# Segundos entre tentativas
WEBHOOK_RETRY_DELAY=2

# Aguarda 2 segundos antes de cada retry
```

**Valores Recomendados:**
- **Sistemas Rápidos**: `1` segundo
- **Sistemas Normais**: `2-3` segundos
- **Sistemas Lentos**: `5-10` segundos

### WEBHOOK_TIMEOUT
```bash
# Timeout por requisição
WEBHOOK_TIMEOUT=15

# Cada tentativa aguarda no máximo 15 segundos
```

**Valores Recomendados:**
- **APIs Rápidas**: `5-10` segundos
- **APIs Normais**: `10-15` segundos
- **APIs Lentas**: `20-30` segundos

---

## 🔀 Modos de Operação

### Modo Original (Sem Retry)
```bash
# Arquivo .env SEM WEBHOOK_RETRY_COUNT definida
WEBAPIPORT=31000
# ... outras configurações
```

**Comportamento:**
- Uma tentativa única de envio
- Falha imediata em caso de erro
- Comportamento original do sistema

### Modo Retry Básico
```bash
# Arquivo .env COM WEBHOOK_RETRY_COUNT definida
WEBHOOK_RETRY_COUNT=3
# WEBHOOK_RETRY_DELAY e WEBHOOK_TIMEOUT usam valores padrão
```

**Comportamento:**
- 3 tentativas de retry (4 tentativas totais)
- Delay de 1 segundo entre tentativas (padrão)
- Timeout de 10 segundos por tentativa (padrão)

### Modo Retry Personalizado
```bash
# Arquivo .env com configuração completa
WEBHOOK_RETRY_COUNT=5
WEBHOOK_RETRY_DELAY=2
WEBHOOK_TIMEOUT=15
```

**Comportamento:**
- 5 tentativas de retry (6 tentativas totais)
- Delay de 2 segundos entre tentativas
- Timeout de 15 segundos por tentativa

---

## 📝 Exemplos Práticos

### Ambiente de Desenvolvimento
```bash
# Configuração rápida para testes
WEBHOOK_RETRY_COUNT=1
WEBHOOK_RETRY_DELAY=1
WEBHOOK_TIMEOUT=5
```

### Ambiente de Produção
```bash
# Configuração balanceada
WEBHOOK_RETRY_COUNT=3
WEBHOOK_RETRY_DELAY=2
WEBHOOK_TIMEOUT=15
```

### Ambiente de Alta Disponibilidade
```bash
# Máxima confiabilidade
WEBHOOK_RETRY_COUNT=5
WEBHOOK_RETRY_DELAY=3
WEBHOOK_TIMEOUT=30
```

### Sem Retry (Debug)
```bash
# Desabilita retry para debugging
WEBHOOK_RETRY_COUNT=0
WEBHOOK_RETRY_DELAY=1
WEBHOOK_TIMEOUT=10
```

---

## 📊 Logs e Monitoramento

### Sucesso na Primeira Tentativa
```
INFO posting webhook
DEBUG posting webhook payload: {"message":...}
DEBUG webhook success on attempt 1
INFO webhook posted successfully
```

### Sucesso Após Retry (Timeout)
```
INFO posting webhook
DEBUG posting webhook payload: {"message":...}
WARN webhook request error (attempt 1/4): Post "https://webhook.com": context deadline exceeded
INFO webhook retry attempt 1/3 after 2s delay
WARN webhook returned status 502 (attempt 2/4)
INFO webhook retry attempt 2/3 after 2s delay
DEBUG webhook success on attempt 3
INFO webhook posted successfully
```

### Falha Não-Retryable (404)
```
INFO posting webhook
DEBUG posting webhook payload: {"message":...}
WARN webhook returned status 404 (attempt 1/4)
INFO error is not retryable, stopping attempts
ERROR webhook failed after 1 attempts: the requested url do not return 200 status code
```

### Timeout com Retry
```
INFO posting webhook
DEBUG posting webhook payload: {"message":...}
WARN webhook request error (attempt 1/4): Post "https://fluxo.com/webhook": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
INFO webhook retry attempt 1/3 after 1s delay
WARN webhook request error (attempt 2/4): Post "https://fluxo.com/webhook": context deadline exceeded
INFO webhook retry attempt 2/3 after 1s delay
DEBUG webhook success on attempt 3
INFO webhook posted successfully
```

### Falha Após Todos os Retries
```
INFO posting webhook
DEBUG posting webhook payload: {"message":...}
WARN webhook request error (attempt 1/4): connection refused
INFO webhook retry attempt 1/3 after 2s delay
WARN webhook request error (attempt 2/4): connection refused
INFO webhook retry attempt 2/3 after 2s delay
WARN webhook request error (attempt 3/4): connection refused
INFO webhook retry attempt 3/3 after 2s delay
WARN webhook request error (attempt 4/4): connection refused
WARN max retry attempts reached
ERROR webhook failed after 4 attempts: connection refused
```

### Métricas para Monitoramento
- **Taxa de Sucesso**: % webhooks que succedem
- **Tentativas Médias**: Número médio de tentativas até sucesso
- **Tempo de Delivery**: Tempo total incluindo retries
- **Timeouts**: Frequência de timeouts

---

## 🧠 Lógica Inteligente de Retry

### Quando o Sistema Faz Retry

O sistema **NÃO** tenta reenviar em todos os tipos de erro. Ele é inteligente e só faz retry em situações que podem ser recuperáveis:

#### ✅ **Casos que FAZEM Retry:**
- **Timeouts**: `context deadline exceeded`, `Client.Timeout exceeded`
- **Erros de Rede**: `connection refused`, `connection reset`, `no such host`
- **Status 5xx**: Erros de servidor (500, 502, 503, etc.)
- **Status 3xx**: Redirecionamentos não tratados
- **Outros status ≠ 200**: Respostas inesperadas

#### ❌ **Casos que NÃO fazem Retry:**
- **Status 4xx**: Erros de cliente (400, 401, 403, 404, etc.) - são permanentes
- **URL Malformada**: Erros na criação da requisição
- **Status 200**: Sucesso - não precisa retry

### Exemplos Práticos

#### Timeout (FAZ Retry):
```
error: Post "https://webhook.com": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
```
**Comportamento**: Faz retry porque pode ser problema temporário

#### Status 404 (NÃO faz Retry):
```
webhook returned status 404
```
**Comportamento**: Para imediatamente porque é erro permanente

#### Status 500 (FAZ Retry):
```
webhook returned status 500
```
**Comportamento**: Faz retry porque servidor pode estar temporariamente indisponível

---

## 🔧 Troubleshooting

### Problema: Muitos Retries
**Sintomas:**
- Logs excessivos de retry
- Sistema de destino sobrecarregado

**Soluções:**
```bash
# Reduzir tentativas
WEBHOOK_RETRY_COUNT=1

# Ou aumentar delay
WEBHOOK_RETRY_DELAY=5
```

### Problema: Timeouts Frequentes
**Sintomas:**
- Muitos erros de timeout nos logs
- Webhooks falhando por tempo

**Soluções:**
```bash
# Aumentar timeout
WEBHOOK_TIMEOUT=30

# Ou reduzir carga
WEBHOOK_RETRY_COUNT=2
WEBHOOK_RETRY_DELAY=3
```

### Problema: Alta Latência
**Sintomas:**
- Delivery muito lento
- Muitos retries desnecessários

**Soluções:**
```bash
# Configuração mais agressiva
WEBHOOK_RETRY_COUNT=2
WEBHOOK_RETRY_DELAY=1
WEBHOOK_TIMEOUT=10
```

### Problema: Sistema Externo Instável
**Sintomas:**
- Falhas intermitentes
- Status 5xx frequentes

**Soluções:**
```bash
# Mais tentativas com delay maior
WEBHOOK_RETRY_COUNT=5
WEBHOOK_RETRY_DELAY=5
WEBHOOK_TIMEOUT=20
```

---

## ❓ FAQ

### Q: O sistema funciona sem configuração?
**A:** Não, o sistema de retry é ativado apenas quando `WEBHOOK_RETRY_COUNT` está definida. Sem essa variável, usa o comportamento original (uma tentativa).

### Q: É compatível com código existente?
**A:** Totalmente! Não requer mudanças no código atual.

### Q: Como desabilitar o retry?
**A:** Remova ou comente a variável `WEBHOOK_RETRY_COUNT` do arquivo `.env`. Alternativamente, configure `WEBHOOK_RETRY_COUNT=0` para desabilitar retries mas manter outras configurações.

### Q: O payload muda entre tentativas?
**A:** Não, o payload é idêntico em todas as tentativas.

### Q: Os headers são mantidos?
**A:** Sim, todos os headers são mantidos:
- `User-Agent: Quepasa`
- `X-QUEPASA-WID: {wid}`
- `Content-Type: application/json`

### Q: Há impacto na performance?
**A:** Mínimo em caso de sucesso. Em caso de falha, aumenta o tempo total devido aos retries.

### Q: Como testar a configuração?
**A:** Use um servidor HTTP de teste que retorne diferentes status codes.

### Q: Funciona com HTTPS?
**A:** Sim, funciona com HTTP e HTTPS.

### Q: O que acontece se o servidor retornar 4xx?
**A:** Trata como falha e faz retry. Futuramente pode ser otimizado para não fazer retry em 4xx.

### Q: Como monitorar a eficácia?
**A:** Analise os logs para:
- Taxa de sucesso na primeira tentativa
- Número médio de retries até sucesso
- Frequência de falhas totais

---

## 📞 Suporte

Para dúvidas ou problemas:
1. Verifique os logs de webhook
2. Ajuste as configurações conforme os exemplos
3. Teste com diferentes valores
4. Consulte a seção de troubleshooting

**Configuração Recomendada Inicial:**
```bash
WEBHOOK_RETRY_COUNT=3
WEBHOOK_RETRY_DELAY=2
WEBHOOK_TIMEOUT=15
```

---

*Esta documentação refere-se ao Sistema de Retry de Webhooks implementado no QuePasa.*
