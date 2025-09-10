# Métricas de Webhook Retry - Documentação

## 📊 Métricas Implementadas

### Métricas Básicas de Webhook

#### `quepasa_webhooks_sent_total`
- **Tipo**: Counter
- **Descrição**: Total de webhooks enviados (todas as tentativas)
- **Uso**: Monitora volume total de requests de webhook

#### `quepasa_webhook_send_errors_total`
- **Tipo**: Counter
- **Descrição**: Total de webhooks que falharam completamente
- **Uso**: Monitora taxa de falha geral de webhooks

### Métricas Específicas de Retry

#### `quepasa_webhook_retry_attempts_total`
- **Tipo**: Counter
- **Descrição**: Total de tentativas de retry (não inclui primeira tentativa)
- **Uso**: Monitora quantas vezes o sistema teve que fazer retry

#### `quepasa_webhook_retries_successful_total`
- **Tipo**: Counter
- **Descrição**: Total de webhooks que tiveram sucesso após retry
- **Uso**: Monitora eficácia do sistema de retry

#### `quepasa_webhook_retry_failures_total`
- **Tipo**: Counter
- **Descrição**: Total de webhooks que falharam mesmo após todos os retries
- **Uso**: Monitora casos onde retry não foi suficiente

### Métricas de Performance

#### `quepasa_webhook_duration_seconds`
- **Tipo**: Histogram
- **Descrição**: Duração total de entrega do webhook (incluindo retries)
- **Uso**: Monitora latência e performance do sistema

## 📈 Exemplos de Queries do Prometheus

### Taxa de Sucesso Global
```promql
# Taxa de sucesso de webhooks
(quepasa_webhooks_sent_total - quepasa_webhook_send_errors_total) / quepasa_webhooks_sent_total * 100
```

### Eficácia do Sistema de Retry
```promql
# Quantos webhooks foram salvos pelo retry
quepasa_webhook_retries_successful_total / quepasa_webhook_retry_attempts_total * 100
```

### Taxa de Retry
```promql
# Porcentagem de webhooks que precisaram de retry
quepasa_webhook_retry_attempts_total / quepasa_webhooks_sent_total * 100
```

### Latência Média de Webhooks
```promql
# Tempo médio de entrega de webhooks
rate(quepasa_webhook_duration_seconds_sum[5m]) / rate(quepasa_webhook_duration_seconds_count[5m])
```

### Webhooks Falhando Mesmo com Retry
```promql
# Rate de webhooks que falharam mesmo após retry
rate(quepasa_webhook_retry_failures_total[5m])
```

## 📊 Dashboard Grafana Sugerido

### Painel 1: Visão Geral
- **Webhook Success Rate**: Taxa de sucesso global
- **Webhooks Sent**: Total de webhooks enviados (gauge)
- **Retry Rate**: Taxa de webhooks que precisaram retry

### Painel 2: Sistema de Retry
- **Retry Success Rate**: Eficácia do sistema de retry
- **Retry Attempts**: Tentativas de retry ao longo do tempo
- **Retry Failures**: Falhas mesmo após retry

### Painel 3: Performance
- **Webhook Latency**: Histogram de latência
- **Average Response Time**: Tempo médio de resposta
- **95th Percentile**: P95 de latência

### Painel 4: Alertas
- **Failed Webhooks**: Webhooks falhando
- **High Retry Rate**: Taxa de retry muito alta
- **Slow Webhooks**: Webhooks muito lentos

## 🚨 Alertas Recomendados

### Alerta: Taxa de Falha Alta
```yaml
- alert: WebhookHighFailureRate
  expr: rate(quepasa_webhook_send_errors_total[5m]) / rate(quepasa_webhooks_sent_total[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "Taxa de falha de webhook alta"
    description: "{{ $value | humanizePercentage }} dos webhooks estão falhando"
```

### Alerta: Sistema de Retry Ineficaz
```yaml
- alert: WebhookRetryIneffective
  expr: rate(quepasa_webhook_retry_failures_total[5m]) / rate(quepasa_webhook_retry_attempts_total[5m]) > 0.5
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Sistema de retry não está funcionando"
    description: "{{ $value | humanizePercentage }} dos retries estão falhando"
```

### Alerta: Latência Alta
```yaml
- alert: WebhookHighLatency
  expr: histogram_quantile(0.95, rate(quepasa_webhook_duration_seconds_bucket[5m])) > 30
  for: 3m
  labels:
    severity: warning
  annotations:
    summary: "Latência de webhook alta"
    description: "P95 de latência está em {{ $value }}s"
```

## 🔍 Monitoramento em Ação

### Cenário 1: Sistema Funcionando Normalmente
```
quepasa_webhooks_sent_total: 1000
quepasa_webhook_send_errors_total: 10
quepasa_webhook_retry_attempts_total: 50
quepasa_webhook_retries_successful_total: 45
```
- **Taxa de sucesso**: 99%
- **Taxa de retry**: 5%
- **Eficácia do retry**: 90%

### Cenário 2: Sistema Externo Instável
```
quepasa_webhooks_sent_total: 1000
quepasa_webhook_send_errors_total: 100
quepasa_webhook_retry_attempts_total: 300
quepasa_webhook_retries_successful_total: 200
```
- **Taxa de sucesso**: 90%
- **Taxa de retry**: 30%
- **Eficácia do retry**: 67%

### Cenário 3: Sistema Externo Fora do Ar
```
quepasa_webhooks_sent_total: 1000
quepasa_webhook_send_errors_total: 800
quepasa_webhook_retry_attempts_total: 2400
quepasa_webhook_retries_successful_total: 0
```
- **Taxa de sucesso**: 20%
- **Taxa de retry**: 240%
- **Eficácia do retry**: 0%

## 🎯 Benefícios do Monitoramento

1. **Visibilidade**: Vê exatamente como o sistema está performando
2. **Detecção Precoce**: Identifica problemas antes que afetem usuários
3. **Otimização**: Dados para ajustar configurações de retry
4. **SLA**: Métricas para acordos de nível de serviço
5. **Debugging**: Facilita investigação de problemas

## 🔧 Como Usar

1. **Configure Prometheus** para coletar métricas do QuePasa
2. **Importe dashboards** no Grafana
3. **Configure alertas** baseados nas métricas
4. **Monitore regularmente** as métricas de webhook
5. **Ajuste configurações** baseado nos dados coletados

As métricas estão instrumentadas no código e serão coletadas automaticamente quando o sistema de retry estiver ativo!
