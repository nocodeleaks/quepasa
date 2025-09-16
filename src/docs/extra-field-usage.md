# Campo Extra para Dispatching

O campo `extra` está disponível tanto para **webhooks** quanto para **RabbitMQ** e permite adicionar dados customizados que serão incluídos em todas as mensagens enviadas.

## Usando o Campo Extra

### Para Webhooks

**POST /api/v1/bot/{token}/webhook**

```json
{
  "url": "https://webhook.exemplo.com/quepasa",
  "forwardinternal": false,
  "trackid": "sistema_crm",
  "extra": {
    "cliente_id": "12345",
    "empresa": "MinhaEmpresa",
    "ambiente": "producao",
    "versao": "1.0"
  }
}
```

### Para RabbitMQ

**POST /api/v1/bot/{token}/rabbitmq**

```json
{
  "connection_string": "amqp://user:pass@rabbitmq.exemplo.com:5672/vhost",
  "forwardinternal": false,
  "trackid": "sistema_crm",
  "extra": {
    "cliente_id": "12345",
    "empresa": "MinhaEmpresa",
    "ambiente": "producao",
    "versao": "1.0"
  }
}
```

## Como o Extra é Enviado

Quando uma mensagem é despachada via webhook ou RabbitMQ, o payload incluirá o campo `extra`:

### Payload do Webhook
```json
{
  "id": "3EB0796DC45C27BE9D8E",
  "timestamp": "2025-09-15T10:30:00Z",
  "type": "text",
  "text": "Olá!",
  "fromMe": false,
  "chat": {
    "id": "5511999999999@s.whatsapp.net",
    "title": "João Silva"
  },
  "extra": {
    "cliente_id": "12345",
    "empresa": "MinhaEmpresa",
    "ambiente": "producao",
    "versao": "1.0"
  }
}
```

### Payload do RabbitMQ
```json
{
  "id": "3EB0796DC45C27BE9D8E",
  "timestamp": "2025-09-15T10:30:00Z",
  "type": "text",
  "text": "Olá!",
  "fromMe": false,
  "chat": {
    "id": "5511999999999@s.whatsapp.net",
    "title": "João Silva"
  },
  "extra": {
    "cliente_id": "12345",
    "empresa": "MinhaEmpresa",
    "ambiente": "producao",
    "versao": "1.0"
  }
}
```

## Casos de Uso

1. **Identificação de Cliente**: Incluir dados do cliente para facilitar integrações
2. **Versionamento**: Rastrear versão da API ou sistema integrado
3. **Ambiente**: Distinguir entre desenvolvimento, teste e produção
4. **Metadados**: Qualquer informação adicional necessária para o sistema receptor

## Validação

- O campo `extra` aceita qualquer estrutura JSON válida
- Pode ser `null` para remover dados extras existentes
- Não há limite de profundidade ou tamanho específico
- É preservado exatamente como enviado

## Atualização

Para atualizar apenas o campo `extra`, envie a mesma requisição POST com os dados atualizados. O sistema irá sobrescrever a configuração existente.

## Remoção

Para remover o campo `extra`, envie:
```json
{
  "url": "https://webhook.exemplo.com/quepasa",
  "extra": null
}
```
