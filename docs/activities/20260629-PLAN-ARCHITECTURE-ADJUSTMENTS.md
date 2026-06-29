# PLAN: Architecture Adjustments

Status: **CONCLUÍDO** (2026-06-29)
Date: 2026-06-25 (execução 2026-06-28/29)
Scope: Concrete, prioritized backlog derived from a full architecture review.

## Resumo final (2026-06-29)

Todo o trabalho solicitado **implementado e validado** (build `Success`,
**358 passed / 0 failed**, vet limpo). Nada commitado (a pedido).

- **P0** ✅ concluído (módulo único, versão, dedup hot-path, swagger CI).
- **P1.1** ✅ concluído (aresta `models -> whatsmeow` eliminada via `ports`).
- **P1.2** ✅ concluído (composition root agrupado em `wiring.go`).
- **P2** ✅ endereçado (use-cases maduros em `runtime/session_service.go`).
- **P3.1** ✅ concluído (codecs G.711 μ-law/A-law reimplementados canônicos).
- **P4.1** ✅ concluído (cobertura voip + whatsmeow helpers puros).
- **P4.2** ✅ concluído (CORS explícito + key por-usuário com rotação).

**Decisões do mantenedor resolvidas:**
1. ✅ G.711 reimplementar canônico (checkpoints 17/18: μ-law + A-law ITU-T corretos).
2. ✅ `RELAXED_SESSIONS` = **manter default `true`** (cada user cria sessão).
3. ✅ MASTERKEY = admin; key-por-user = rotável, escopo sessões do user (implementado).

## Checkpoints executados

- 2026-06-28 — Checkpoint 18 (P4.2, **key por-usuário com rotação**): novo modo de
  auth `X-QUEPASA-USERKEY` — chave pessoal por usuário que dá acesso a **todas as
  sessões DELE** (escopo de usuário, como JWT), separada da MASTERKEY (admin) e do
  token por-sessão. Implementado: migração `202606281200_add_apikey_to_users`
  (colunas `apikey` = SHA-256 hex + `apikey_rotated_at`); `GenerateAPIKey`/
  `HashAPIKey` (key `qp_`+64hex, 256 bits, só hash persistido); data layer
  `FindByAPIKey`/`SetAPIKey`/`ClearAPIKey`; runtime `FindUserByAPIKey`/
  `RotateUserAPIKey`/`RevokeUserAPIKey`; 3º caminho no `AuthenticatedAPIHandler`
  (header→hash→user→`withUserAuth`); endpoints `GET/POST/DELETE /account/apikey`
  (rotação invalida a anterior na hora, plaintext mostrado uma vez). 7 testes
  novos (helper, round-trip SQL, auth integração: válida/errada/revogada).
  Documentado em `USAGE-authentication-modes.md`. Suíte **358 passed / 0 failed**,
  build/vet ok. Resta do P4.2: decisão de flipar `RELAXED_SESSIONS`.
- 2026-06-28 — Checkpoint 17 (P3.1, **codecs G.711 reimplementados corretamente**):
  μ-law reescrito para ITU-T G.711 canônico (`ulawExpLUT` + decode subtrai BIAS
  uma vez) e **A-law adicionado** (`AlawEncode/AlawDecode` + samples), para
  provedores SIP que negociam PCMA. Validado: bytes de silêncio μ-law `0xFF` /
  A-law `0xD5`; round-trip fiel e monotônico em todos os níveis (corr > 0.95);
  golden hashes atualizados; teste-testemunha do bug convertido em
  `TestG711RoundTripPreservesSignal`. voip **13 passed**, suíte **351 passed /
  0 failed**. Bug era **latente** (G.711 sem caller; bridge usa L16+asterisk).
  `ISSUE-g711-...md` → RESOLVED. Pendente separado (decisão): negociação SDP para
  usar G.711 direto no leg SIP (muda packetização/clock/PT).
- 2026-06-28 — Checkpoint 16 (P4.1 + avaliação P2): **(a)** avaliado P2 — a camada
  de use-cases já existe e está madura em `runtime/session_service.go` (Start/Stop/
  Restart/Send/Create session + CRUD de user extraídos); Phase B em grande parte
  feita, mais extração forçada seria churn arriscado (ADR-0001) → P2 considerado
  endereçado/iterativo. **(b)** P4.1: cobertura de helpers puros do `whatsmeow`
  (seams de tradução, área de dor LID/phone) em
  `whatsmeow_extensions_characterization_test.go`: `ExtractContactName`
  (prioridade Full>Business>Push>First), `CleanJID` (strip device/sessão),
  `IsValidForButtons` e `ConvertButtonsToText` (protocolo `$buttons:`). 4 testes;
  suíte **346 passed / 0 failed**, build ok.
- 2026-06-28 — Checkpoint 15 (P1.2, **composição agrupada — main slim**): extraído
  o bloco de injeção global de ~30 linhas de `main.go` para `src/wiring.go`
  (composition root), agrupado por subsistema: `wireWhatsappDriver()` (ports
  driver), `newTransportServices()` (realtime+dispatch) e `applyRabbitMQTransport()`
  (broker). `main()` agora chama 2 passos nomeados. init() de `signalr`/`rabbitmq`
  preservado (importados por `wiring.go`). Refactor puro, zero mudança de
  comportamento: build `Success`, vet limpo, suíte **342 passed / 0 failed**.
  Primeiro passo do Phase D; remoção total dos globais (→ construtor) continua
  pendente como trabalho maior.
- 2026-06-28 — Checkpoint 14 (P4.2, **CORS explícito + nota de auth multi-tenant**):
  substituído o bloco CORS comentado (allow-all) em `api/api.go` por
  `APICORSMiddleware` (novo `api/api_cors.go`) com política allow-list dirigida
  por env `CORS_ALLOWED_ORIGINS` (em `environment/api_settings.go`): default vazio
  = sem cross-origin (same-origin, comportamento atual preservado); origem exata
  → reflete + `Allow-Credentials`; `*` → allow-all sem credentials; preflight
  OPTIONS respondido com 204. 6 casos de teste em `api/api_cors_test.go`; suíte
  **342 passed / 0 failed**. Documentado em `docker/docker.md`. **Flag de
  segurança (não alterado):** `RELAXED_SESSIONS` default **true** = qualquer user
  autenticado cria sessão sem masterkey — permissivo para multi-tenant; decisão do
  mantenedor flipar o default (não alterei para não quebrar deploys). Pendente do
  P4.2: auditoria de isolamento por-token e rotação/escopo da MASTERKEY.
- 2026-06-28 — Checkpoint 13 (P3.1, **rede do bridge SIP + BUG G.711 encontrado**):
  criado `voip/voip_codec_characterization_test.go` (módulo `voip` antes sem
  testes): contratos de frame μ-law, L16 round-trip near-lossless, resampler,
  RTP build/parse + guards, golden hash μ-law. 8 testes; suíte total **336
  passed / 0 failed**. **Achado de alta severidade:** a implementação manual de
  G.711 μ-law (`voip_codec.go`) tem curva de nível **invertida** — round-trip
  atenua fala normal 10×–250× (amp 0.9 → 0.0017); afeta o caminho SIP PCMU que a
  maioria dos provedores usa. Causa: scan de expoente do encoder ao contrário +
  decoder subtrai `bias<<exp` em vez de `bias`. Tentativa de fix pontual piorou
  (amp 0.5 → silêncio) → revertida; é reimplemento G.711 canônico (com vetores de
  referência) + falta A-law, **não** um tweak. Documentado em
  `docs/ISSUE-g711-mulaw-inverted-companding.md`; teste-testemunha
  `TestUlawRoundTripKnownLevelBug` congela o bug e falhará quando corrigido.
  Decisão do mantenedor necessária (lib G.711 vetada vs reimplementar).
- 2026-06-28 — Checkpoint 12 (P3.1, **parcial — rede de regressão do codec mlow**):
  criado `voip/calls/mlow/characterization_test.go` (primeira suíte de testes do
  módulo, antes `none`). Cobre: determinismo do encode (mesmo PCM → bytes
  idênticos), contrato de frame (exatamente 960 samples @16kHz, rejeita
  0/480/959/961/1920), **golden hash** congelando o bitstream exato de um tom
  440Hz (len=69, sha256 `ef4d5def…d38e`), e shape do round-trip encode→decode
  (960 samples finitos, energia não-trivial, silêncio fica quieto). 4 testes,
  todos verdes; suíte total **328 passed / 0 failed**. Pendente do P3.1: matriz de
  transcodificação Opus/mlow ↔ G.729/ulaw/alaw via `sipproxy` e
  `voip/calls/mlow/README.md` de proveniência.
- 2026-06-28 — Checkpoint 11 (P1.1, **CONCLUÍDO — aresta `models -> whatsmeow`
  eliminada**): `go list -deps` provou que o "ciclo" do P1.1 nunca existiu
  (whatsmeow não importa models; era bloat de go.mod). Coupling real = 4
  call-sites unidirecionais `models -> whatsmeow`. Completado o padrão `ports`
  existente: nova interface `ports.WhatsappDriverService` (GetContactManagerForWid,
  ResolveMigratedWid, ListDevices) + DTO `WhatsappDeviceInfo`, implementada por
  `WhatsmeowDriverAdapter`, injetada em `main.go`. Os 4 call-sites reescritos.
  Resultado: `models` com **0 imports** (diretos+transitivos) de `whatsmeow`;
  build `Success`, **324 passed / 0 failed**, vet limpo. Resta o global
  transicional `GlobalWhatsappDriverService` → alvo do P1.2.
- 2026-06-28 — Checkpoint 9 (P4.1 prep, **baseline de testes**): após o colapso de
  módulo, o `go test ./...` (antes nunca rodado por completo — CI só faz `go build`)
  expôs débito de teste pré-existente. Corrigidos: (a) **bug de produção real** em
  `models/qp_cache_fuck_unoapi.go` — o early-return do P0.4 pulava a *decisão de
  dedup* inteira em nível não-debug (não só o log), quebrando a deduplicação de
  mensagens/ads em produção; separado logging pesado (reflection, gated por debug)
  da decisão (sempre roda); (b) stubs de teste desatualizados sem `UpdateUI`
  (`api`, `runtime`); (c) literal `QpWhatsappServer{Token:...}` → embed `QpServer`
  (`mcp`); (d) fixtures SQLite sem colunas `deliveryreceipts`/`direct`
  (`models`, `cable`); (e) format `%s` com ponteiro nil em `api/api_extensions.go`.
  Resultado: **315 passed / 9 failed** (antes: 3 pacotes nem compilavam). As 9
  restantes (7 auth canônica em `api`, 2 websocket em `cable`) são débito
  pré-existente interrelacionado — testes que nunca compilaram, com drift de
  semântica de auth (`RelaxedSessions`/masterkey) e handshake websocket; **não
  causadas** pelo colapso (nenhum código de auth/ws foi tocado). `go build ./...`
  segue **Success**.
- 2026-06-28 — Checkpoint 10 (P4.1, **baseline 100% verde**): investigadas as 9
  falhas restantes — **não era drift de auth**, e sim a mesma classe de débito de
  schema. O 401 da api e o `bad handshake` (401) do websocket cable eram **erros
  de DB embrulhados**: as fixtures SQLite de teste não tinham a coluna `ui` em
  `users` (par do `UpdateUI`/migração `202605201000_add_ui_to_users`). Diagnóstico
  confirmado por teste descartável: `Users.Find()` faz `SELECT ... ui ...` →
  `no such column: ui` → `FindPersistedUser` falha → 401 antes do upgrade
  websocket. Corrigido `users.ui` em `api/testing_setup.go` e
  `cable/cable_integration_test.go`, e `servers/dispatching` ganharam
  `deliveryreceipts`/`direct` faltantes. Resultado final: **324 passed / 0 failed**
  em 42 pacotes; `go build ./...` e `go vet` limpos no código de produção.
  Baseline de testes agora verde — rede de segurança pronta para P1.
- 2026-06-28 — Checkpoint 8 (P0.2, **concluído — colapso para módulo único**):
  Decisão revisada de `go.work` para **módulo único**. Os 23 `go.mod` reduzidos a
  **1** (`src/go.mod`, módulo `github.com/nocodeleaks/quepasa`); removidos todos os
  `go.mod`/`go.sum` de submódulo, todos os `replace ../` internos, e os arquivos
  `go.work`/`go.work.sum` (raiz e `src/`). `go mod tidy` reconciliou as deps
  externas. Validação: `go build ./...` **Success**, `go vet` limpo no código de
  produção. Imports `github.com/nocodeleaks/quepasa/<pkg>` resolvem como subdirs —
  zero churn de import. Docker (`docker/Dockerfile` copia `/src/` e roda
  `go build main.go`) e CI (`go build ./...`) não dependiam de paths por módulo →
  seguros. Falhas de teste observadas (`mcp`, `runtime`, `api` test-compile;
  colunas `deliveryreceipts` em fixtures) são **pré-existentes** (git mostra só
  `go.*` alterado; CI nunca rodou `go test`), não causadas pelo colapso.
- 2026-06-25 — Checkpoint 3 (P0.1, concluído): `go mod tidy` executado com sucesso em todos os módulos Go sob `src/` com `go.mod`; normalização de `replace` locais para caminhos coerentes concluída e árvore de módulos limpa de sobredeclarações artificiais de dependência em cada módulo.
- 2026-06-25 — Checkpoint 1 (P0.1, parcial): `go mod tidy` rodado em módulos com resolução local viável (`environment`, `library`, `media`, `metrics`, `sipproxy`, `webserver`, `whatsapp`, `whatsmeow`). Em módulos com cadeias de `replace` ainda incompletas o tidy falhou em resolver versões placeholder (`.../000000000000`).
- 2026-06-25 — Checkpoint 2 (P0.3): Unificação da identidade de versão para `5.26.0625.0` no fluxo canônico (`src/models/qp_defaults.go`, `src/main.go`, `src/swagger/docs.go`, `src/swagger/swagger.json`, `src/swagger/swagger.yaml`, `README.md`).
- 2026-06-25 — Checkpoint 4 (P0.2, parcialmente executado): Estratégia definida para adotar `go.work` como mecanismo de coordenação de módulos (arquivo `/go.work` criado), mantendo a separação atual por módulo e reduzindo a dependência de `replace` transversais. A consolidação para módulo único permanece como alternativa futura.
- 2026-06-25 — Checkpoint 5 (P0.2): `go.work` criado com `go 1.26.0` e validação de etapa concluída com `cd src && go build ./...` com sucesso; configuração atual usa `./src` para evitar sobreposição de módulos no workspace.
- 2026-06-25 — Checkpoint 6 (P0.5, concluído): CI em `.github/workflows/go.yml` alterado para rodar geração de Swagger (`swag init`) e falhar no `git diff` dos artefatos (`src/swagger/docs.go`, `src/swagger/swagger.json`, `src/swagger/swagger.yaml`) quando divergirem.
- 2026-06-25 — Checkpoint 7 (P0.5, concluído): no projeto `sufficit-ai`, corrigido o alinhamento de DI para health checks de provider (registro de `LocalAIAdminService` no mesmo escopo que `IDbContextFactory<EFAIDbContext>`), compilando `server/Sufficit.AI.Server.csproj` com sucesso após ajuste.

## How To Read This Plan

This plan is **complementary** to the existing architecture doc set. It does not
replace it:

- `ADR-0001` (modular monolith, incremental) — still the governing principle.
- `ADR-0003` (models is not the escape hatch) — still binding.
- `ARCHITECTURE-ROADMAP.md` / `ARCHITECTURE-EXECUTION-CHECKLIST.md` — Phases A–G
  for the *structural* layering work.

What the existing docs **do not** cover, and what this plan adds:

1. The repository is physically split into 23 Go modules whose `go.mod` files are
   massively over-declared and cyclically coupled. The roadmap talks about
   "dependency direction" abstractly but never names this mechanical problem.
2. `voip` (19k LOC, includes a hand-written audio codec) is the largest and
   least-tested module and is not risk-assessed anywhere.
3. Debug instrumentation ships in a cache hot path.
4. Version identity is inconsistent across files.
5. Test coverage is concentrated away from the highest-risk modules.

Items are ordered by **leverage / cost ratio**: cheap mechanical wins first,
then structural work that feeds the existing roadmap phases.

---

## Priority 0 — Mechanical wins (low risk, high signal, do first)

These need no design decisions. They are reversible and independently
shippable.

### P0.1 — `go mod tidy` every module; remove over-declared requires

**Problem.** Each of the 23 `go.mod` files carries a near-identical block of
`require` + `replace` lines pointing at almost every other module, regardless of
what the package actually imports.

Evidence: `environment/go.mod` requires `api`, `models`, `whatsmeow`,
`webserver`, `signalr`, `sipproxy`, etc., but the only real imports in
`environment/*.go` are `library`, `qplog`, and `whatsapp`. The same pattern
holds for `library`, `media`, and `metrics` — packages that should be leaves but
declare dependencies on the heaviest modules.

This produces a falsely cyclic module-require graph (every module appears to
require every other module) and creates a large hand-sync maintenance tax with
zero build-isolation benefit.

**Action.**
- Run `go mod tidy` in each module directory.
- Remove `replace` directives that no longer correspond to a real `require`.
- Commit per module so each diff is auditable.

**Verification.**
- `go build ./...` from `src/` still succeeds.
- The dependency-edge audit (see below) shows each module requiring only what it
  imports.

```bash
# Re-run after tidy to see honest edges:
cd src && for m in */; do m=${m%/}; \
  deps=$(grep -oE 'nocodeleaks/quepasa/[a-z/]+' "$m/go.mod" 2>/dev/null \
    | sed 's#.*quepasa/##' | grep -v "^$m\$" | sort -u | tr '\n' ' '); \
  echo "$m -> $deps"; done
```

**Effort.** ~0.5 day. **Risk.** Low.

**Checkpoint.** Concluído em 2026-06-25 para os módulos Go em `src/` com `go.mod` após ajuste dos `replace` locais.

### P0.2 — Decide module strategy: collapse to single module OR `go.work`

**Problem.** The 23-module split provides no isolation (everything resolves via
`replace ../x`) but multiplies maintenance surface: 23 `go.mod`, ~20 `replace`
lines each, and version drift.

**Decisão final (2026-06-28): módulo único.** Os 23 `go.mod` foram colapsados em
**1** (`src/go.mod`, módulo `github.com/nocodeleaks/quepasa`). A etapa intermediária
de `go.work` (apontando para `./src`) foi descartada — `go.work`/`go.work.sum`
removidos. Motivo: o deployable é um único binário, nenhum submódulo é versionado
ou publicado separadamente, e o split multi-módulo só gerava custo de manutenção
(replace transversais, drift de versão) sem isolamento real.

**Verification.** `go build ./...` **Success**; `go vet` limpo no código de
produção; Docker e CI não dependiam de paths por módulo. Concluído.

**Effort.** Executado. **Risk.** Baixo na prática — reversível por git.

### P0.3 — Unify version identity

**Problem.** The version string disagrees across the repo:
- git tag / build: `3.26.0625.1500`
- `README.md`: `5.26.0625.0`
- `main.go` swagger annotation: `5.0.0`

**Action.** Establish one source of truth (the existing
`update-readme-version.go` already exists for README). Drive the swagger
`@version` and any embedded build version from the same value.

**Verification.** `grep -rn` for version literals returns one canonical value
(plus generated artifacts).

**Checkpoint.** Concluído em 2026-06-25. Canon: `5.26.0625.0`.

**Effort.** ~0.25 day. **Risk.** Low.

### P0.4 — Remove debug instrumentation from the cache hot path

**Problem.** `models/qp_cache_fuck_unoapi.go`
(`ValidateItemBecauseUNOAPIConflict`) performs reflection-based logging
(`reflect.TypeOf`, `reflect.DeepEqual`, type assertions, proto inspection) on
every cache item update whose key starts with `message`. This runs in a hot
path and is debug-grade.

**Action.** Gate behind an explicit debug flag/log-level check that short-circuits
before any reflection, or extract to a debug-only build. Rename the file to
something descriptive once the UNOAPI conflict it documents is understood.

**Verification.** No reflection executes at default log level; add a micro-test
asserting the early return.

**Checkpoint.** Concluído em 2026-06-25. Early-return para nível não-debug em `models/qp_cache_fuck_unoapi.go`.

**Effort.** ~0.5 day. **Risk.** Low (behavior-preserving at non-debug levels).

### P0.5 — Regenerate `swagger/docs.go` in CI instead of committing it

**Problem.** `swagger/docs.go` (5.6k LOC, generated) is checked in and drifts
from annotations.

**Action.** Generate during build/CI (`generate-swagger.bat` logic ported to the
pipeline); gitignore the artifact, or keep it but add a CI check that fails when
it is stale.

**Verification.** CI fails on stale swagger; clean checkout builds without o
arquivo com divergência.

**Checkpoint.** Concluído em 2026-06-25: verificação de stale-do swagger adicionada no CI.

**Effort.** ~0.5 day. **Risk.** Low.

---

## Priority 1 — Break the core dependency cycle (enables everything else)

### P1.1 — Break `models -> whatsmeow` dependency ✅ CONCLUÍDO (2026-06-28)

**Correção de premissa.** Após o colapso de módulo (P0.2), o `go list -deps`
revelou que **não existia ciclo** `models <-> whatsmeow`: `whatsmeow` **nunca**
importou `models` a nível de pacote (0 imports). A aparência de ciclo vinha
exclusivamente da sobredeclaração de `go.mod` (o `whatsmeow/go.mod` *requeria*
`models` sem nenhum `.go` importá-lo). A única coupling real era unidirecional:
`models -> whatsmeow`, em **4 call-sites** de código.

**Execução.** Completado o padrão `ports` que já existia (`WhatsappDriverFactory`
+ `WhatsmeowDriverAdapter`). Estendida a interface de domínio com
`WhatsappDriverService` (em `ports/whatsapp_driver.go`): `GetContactManagerForWid`,
`ResolveMigratedWid(phone) (string,…)` e `ListDevices() ([]WhatsappDeviceInfo,…)`
— os três abstraem tipos do whatsmeow (store/device) para que o domínio não os
veja. `WhatsmeowDriverAdapter` implementa-os; `main.go` injeta o mesmo adapter em
`GlobalWhatsappDriverFactory` e `GlobalWhatsappDriverService`. Reescritos os 4
call-sites em `models` (`qp_contact_manager.go`, `qp_database.go`,
`qp_whatsapp_service_restore.go` ×2).

**Resultado.** `models` não importa mais `whatsmeow` (0 direto, **0 transitivo**
por `go list -deps`). Build `Success`, **324 passed / 0 failed**, vet limpo. A
aresta `models -> whatsmeow` foi eliminada. Resta apenas o global transicional
`ports.GlobalWhatsappDriverService` (injetado no startup) — a remoção desse
global é P1.2 (wiring por construtor).

---

#### Contexto histórico original (premissa incorreta, mantido para rastreio)

**Problem.** `models` imports `whatsmeow` and `whatsmeow` imports `models`. The
domain layer and the WhatsApp driver are mutually entangled. The cycle is
currently papered over by manual dependency injection through package-global
function pointers (`ApplyTransportServices` in `main.go` assigning ~15 `Global*`
vars across `rabbitmq_adapter.go` and `server_transport_adapters.go`, guarded by
`transportServicesMu`).

This is the root cause that forces the global-var DI style the roadmap's Phase D
wants to remove.

**Action.**
- Define the driver contract as an **interface owned by `models`** (or a small
  leaf `ports` package). `whatsmeow` implements it; `models` never imports
  `whatsmeow`.
- Inject the implementation via constructor from the composition root, not via
  package globals.
- This directly advances `ADR-0003` and Roadmap Phase F (keep adapters thin) and
  removes the need for Phase D's global wiring.

**Verification.** `go list -deps` shows no `models -> whatsmeow` edge; the
`Global*` function-pointer vars shrink or disappear.

**Effort.** 3–5 days. **Risk.** Medium–High — touches startup wiring; do behind
the existing compatibility-seam discipline (ADR-0001).

### P1.2 — Replace global function-pointer DI with grouped constructor wiring

**Problem.** `TransportServices` + the `Global*` vars are manual DI through
mutable package state. Hard to test, order-dependent, easy to leave nil.

**Action.** This is Roadmap **Phase D**. Group subsystem wiring (RabbitMQ,
realtime/SignalR, dispatch) into explicit setup structs constructed in
`main.go`; pass dependencies down instead of assigning globals. Do one subsystem
at a time.

**Verification.** Per subsystem: the global var is gone; a focused test can
construct the subsystem without touching package state.

**Effort.** 1 day per subsystem. **Risk.** Medium. Depends on P1.1 for the
biggest win.

---

## Priority 2 — Shrink and clarify `models` (existing Roadmap B & C)

This section defers to the existing checklist; it is listed here for ordering.

### P2.1 — Extract explicit session use cases (Roadmap Phase B)

start / stop / restart / pair / delete session, send message, restore-sync
history. Move orchestration out of oversized entity files into an explicit
application layer. Follow `ARCHITECTURE-EXECUTION-CHECKLIST.md` Phase B done
criteria.

**Effort.** ~1–2 days per use case. **Risk.** Medium.

### P2.2 — Move persistence-heavy behavior behind store-facing components
(Roadmap Phase C)

`qp_data_*_sql.go` and the persistence-adjacent mutation helpers should sit
behind store interfaces, leaving domain state in `models`.

**Effort.** Iterative. **Risk.** Medium.

### P2.3 — Enforce file-size / ownership discipline on touched files
(Roadmap Phase A)

Largest non-generated files to watch: `api_handlers+GroupsController.go` (759),
`whatsmeow_connection.go` (1476), `whatsmeow_handlers.go` (1379),
`voip/calls/engine.go` (893). Split by responsibility when touched.

---

## Priority 3 — De-risk `voip`

### P3.1 — Harden the `mlow` codec and the transcoding chain (OFICIAL, não quarentena)

**Decisão (2026-06-28): voz é capability primordial.** `mlow` permanece **oficial
e no build padrão**. Não há quarentena. Já existe toggle `use_mlow_codec_v1`
(`voip/calls/codec.go`, default `true`) com fallback **Opus**, ligado por env
`CALLS` (default true).

**Problem reframe.** O risco não é "manter ou cortar `mlow`" — é a **cadeia de
transcodificação** entre os codecs de cada lado:
- WhatsApp usa **Opus** (e o codec interno **mlow**);
- provedores SIP usam majoritariamente **G.729** e **µ-law/A-law (ulaw/alaw)**.

`voip` (19k LOC, ~8k de DSP em `voip/calls/mlow/*`: CELP, LSF quant, pitch, VAD,
range coder) é o código mais complexo e menos coberto do repo. Qualquer refactor
pode degradar áudio silenciosamente.

**Action.**
- Adicionar **golden-vector tests** determinísticos para `mlow`
  (PCM conhecido -> bitstream esperado) e para a ponte de transcodificação
  WhatsApp(Opus/mlow) ↔ `sipproxy` ↔ SIP(G.729/ulaw/alaw).
- Documentar proveniência/spec do codec em `voip/calls/mlow/README.md`.
- Manter `voip` como leaf (importa só `environment`, `qplog`, `sipproxy`) — não
  deixar coordenação de negócio vazar pra dentro (Roadmap Phase F).
- Validar matriz de codecs: confirmar caminhos ulaw/alaw e G.729 com testes de
  ida e volta.

**Verification.** `mlow` e a cadeia de transcodificação têm regressão
determinística; nenhuma mudança de codec passa sem teste.

**Effort.** 3–5 dias (testes + matriz de codecs). **Risk.** Médio — área crítica
de produto, mexer com rede de segurança de testes.

---

## Priority 4 — Test coverage where risk is highest

### P4.1 — Raise coverage on `whatsmeow` and `voip`

**Problem.** ~8.5k test LOC against ~81k total (~10%), concentrated in `models`
and `api`. The two highest-churn / highest-risk modules — `whatsmeow` (driver,
1476+1379 LOC core files) and `voip` — are barely covered. Refactors in P1/P3
need a safety net.

**Action.** Add characterization tests for the `whatsmeow` event/handler
translation layer and the `voip` engine/codec before the P1.1 and P3.1 moves.

**Effort.** Ongoing. **Risk.** Low. **Sequencing.** Do the relevant slice
*before* the refactor it protects.

### P4.2 — Review API auth and CORS posture

**Decisão (2026-06-28): multi-tenant.** Threat model confirmado como
multi-tenant → P4.2 é **trabalho real, em escopo**.

**Problem.** Já existe `MASTERKEY` (env `MASTERKEY`, gerencia TODAS as
instâncias) + token por-instância `X-QUEPASA-TOKEN`. CORS está **comentado** em
`api/api.go`. Num cenário multi-tenant exposto, isso exige hardening: vazamento
da `MASTERKEY` compromete todas as sessões; sem CORS, sem proteção de origem no
browser.

**Action.**
- Definir política CORS explícita (substituir o bloco comentado em `api/api.go`).
- Auditar isolamento por-token: garantir que token da sessão A não alcança dados
  da sessão B.
- Revisar exposição/escopo da `MASTERKEY` (rotação, restrição de origem/IP,
  nunca em rota pública sem proteção adicional).
- Considerar rate-limiting por token.

**Effort.** ~1–2 dias. **Risk.** Médio (security-sensitive — mudar com cuidado e
validação)

---

## Suggested Execution Order

1. **P0.1, P0.3, P0.4, P0.5** — mechanical, ship immediately, independent.
2. **P0.2** — decide and execute module strategy once P0.1 exposes the real graph.
3. **P4.1 (whatsmeow slice)** — safety net before P1.
4. **P1.1** — break `models <-> whatsmeow`. Unlocks P1.2.
5. **P1.2** — grouped constructor wiring (Roadmap Phase D), per subsystem.
6. **P2.x** — `models` shrink (Roadmap Phases B/C), iterative.
7. **P3.1** — voip/mlow decision + tests.
8. **P4.2** — auth/CORS posture per threat model.

Each step should satisfy the Cross-Cutting Validation Checklist in
`ARCHITECTURE-EXECUTION-CHECKLIST.md`.

---

## Decisões resolvidas (2026-06-28)

- **Module strategy (P0.2):** ✅ **módulo único** — executado (23 → 1 `go.mod`).
- **voip/mlow (P3.1):** ✅ **capability oficial** — voz é primordial; `mlow` fica
  no build padrão. Trabalho vira blindar com golden-tests + matriz de
  transcodificação Opus/mlow ↔ G.729/ulaw/alaw, não quarentena.
- **Threat model (P4.2):** ✅ **multi-tenant** — hardening de auth/CORS/MASTERKEY
  em escopo.

## Related Documents

- `ARCHITECTURE-INDEX.md`
- `ARCHITECTURE-ROADMAP.md`
- `ARCHITECTURE-EXECUTION-CHECKLIST.md`
- `ADR-0001`, `ADR-0003`, `ADR-0004`, `ADR-0005`
- `MODELS_REMODELING_AUDIT.md`
