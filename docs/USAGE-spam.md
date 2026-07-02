# Uso do endpoint `/spam`

Este documento descreve o comportamento atual do endpoint `/spam` no QuePasa e deve ser atualizado sempre que o fluxo de envio em massa mudar.

## Status atual

O `/spam` envia uma mensagem usando uma seção WhatsApp disponível no serviço.

Rotas registradas hoje:

- `POST /spam`
- `POST /current/spam`
- `POST /v4/spam`

O comentário antigo no código mencionando `/v4/bot/{token}/spam` não representa a rota registrada atualmente.

## Autenticação

O endpoint exige a master key do QuePasa.

A master key pode ser enviada por:

- query string: `?masterkey=...`
- form field: `masterkey`
- header: `X-QUEPASA-MASTERKEY`

Não documentar valores reais de master key em exemplos, issues ou logs.

## Corpo da requisição

O corpo segue a mesma base do envio comum (`/send`). Exemplo mínimo:

```json
{
  "chatId": "5511999999999@s.whatsapp.net",
  "text": "Mensagem de teste"
}
```

Também podem existir campos aceitos pelo fluxo comum de envio, como anexos, URL e conteúdo codificado, porque após escolher uma seção o `/spam` delega para o mesmo handler usado por `/send`.

## Como a seção é escolhida

O `/spam` consulta a tabela `spam_sections` antes de escolher a seção.

Fluxo com `spam_sections` vazia:

1. valida a master key;
2. chama `runtime.GetFirstReadySession`;
3. usa a primeira sessão em memória com status `Ready`.

Consequências:

- este é o comportamento legado;
- a seleção pode não ser determinística, porque depende da iteração das sessões carregadas em memória.

Fluxo com `spam_sections` configurada:

1. valida a master key;
2. carrega `spam_sections` ordenada por `position`;
3. ignora itens desativados;
4. usa somente seções presentes nessa tabela;
5. escolhe a primeira seção configurada que esteja em memória com status `Ready`.

Se a tabela tiver itens, mas nenhum item ativo estiver `Ready`, o endpoint retorna erro. Nesse cenário não há fallback para todas as sessões, porque a existência de itens na tabela indica seleção explícita.

Para envio por seção específica fora dessa fila, usar `/send` com o `token` da seção.

## Tabela `spam_sections`

Campos:

- `token`: token da seção em `servers.token`; chave primária;
- `position`: ordem de tentativa no `/spam`;
- `enabled`: permite pausar uma seção sem removê-la;
- `label`: rótulo opcional para a UI;
- `created_at`;
- `updated_at`.

Os registros são removidos automaticamente quando a seção correspondente em `servers` é apagada.

## Administração

O app administrativo fica em:

```http
GET /apps/spam/
```

O app usa a master key localmente na tela. Se `QUEPASA_MASTERKEY` não estiver configurada no serviço, a tela mostra bloqueio e não tenta operar a fila.

Endpoints administrativos:

- `GET /api/spam/status`: informa se a master key está configurada e se a requisição atual está destravada;
- `GET /api/spam/sections`: lista a fila configurada;
- `POST /api/spam/sections/search`: pesquisa seções cadastradas em qualquer usuário/contexto;
- `POST /api/spam/sections`: adiciona uma seção à fila;
- `PATCH /api/spam/sections`: altera `enabled`, `position` ou `label`;
- `DELETE /api/spam/sections?token=...`: remove uma seção da fila;
- `POST /api/spam/sections/reorder`: grava a ordem da fila.

Todos os endpoints administrativos, exceto `GET /api/spam/status`, exigem master key.

## Erros esperados

Quando a master key está ausente/inválida, o endpoint responde com erro usando HTTP `423 Locked`.

Quando a fila explícita está configurada mas nenhuma seção ativa está `Ready`, o endpoint responde com erro informando que não há seção de spam configurada pronta.

## Mudanças planejadas

Proposta para seleção explícita posterior:

- permitir envio por uma seção:

```http
POST /spam?token=TOKEN_DA_SECAO
```

- permitir envio por várias seções:

```json
{
  "tokens": [
    "TOKEN_DA_SECAO_1",
    "TOKEN_DA_SECAO_2"
  ],
  "chatId": "5511999999999@s.whatsapp.net",
  "text": "Mensagem de teste"
}
```

Regras desejadas:

- se `token` for informado, usar somente aquela seção;
- se `tokens` for informado, usar somente as seções listadas;
- validar que cada seção existe e está `Ready`;
- retornar erro claro quando nenhuma seção informada estiver apta;
- manter compatibilidade temporária com o comportamento legado somente se isso for decidido explicitamente.

## Checklist de evolução

- [x] Documentar o comportamento legado atual.
- [x] Definir contrato da tabela `spam_sections` e do app `/apps/spam/`.
- [x] Criar migração da tabela `spam_sections`.
- [x] Criar provider SQL para `spam_sections`.
- [x] Criar endpoints administrativos `/api/spam/*`.
- [x] Alterar `POST /spam` para respeitar a tabela quando configurada.
- [x] Criar app `/apps/spam/` com layout próprio.
- [x] Adicionar testes unitários do provider.
- [x] Adicionar testes unitários dos controllers administrativos.
- [x] Atualizar exemplos deste documento após a implementação.
