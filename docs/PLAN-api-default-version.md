## Plan: Configurable Default API Version

Adicionar uma variável de ambiente para controlar qual versão responde no alias sem versão (`/api/...`), mantendo todas as versões explícitas funcionando em paralelo. A recomendação é separar o alias base do registro versionado, permitir `v4` e `v5` como defaults suportados para o alias sem versão, expor essa escolha em `settings` e `preview`, e migrar o frontend `src/apps/vuejs` para usar explicitamente `/api/v5/...` para evitar quebra quando `/api/...` apontar para `v4`.

**Steps**
1. Fase 1 — Modelar a nova configuração de ambiente em `src/environment/api_settings.go`.
   - Adicionar constante da env, por exemplo `API_DEFAULT_VERSION`.
   - Adicionar campo em `environment.APISettings` para a versão default do alias sem versão.
   - Carregar com default inicial alinhado à decisão de rollout (pelo seu cenário, `v4` durante a migração; se a equipe preferir rollout neutro, manter `v5` e mudar apenas por env).
   - Validar/normalizar o valor para evitar aliases inválidos. Recomendação: suportar explicitamente `v4` e `v5` para o alias base; manter `v3` apenas como rota explicitamente versionada, porque seu shape (`/v3/bot/{token}`) não combina com o contrato canônico sem versão.
2. Fase 2 — Expor a configuração no environment discovery.
   - Atualizar `src/environment/environment_settings_preview.go` para incluir um campo público como `default_api_version` no preview.
   - Garantir que o mesmo campo apareça em `settings` autenticado automaticamente via `APISettings`.
   - Revisar `src/api/api_handlers+EnvironmentController.go` apenas para validar que a sanitização continua correta e que nenhum segredo novo é exposto.
3. Fase 3 — Desacoplar alias sem versão do registro fixo atual. *depende da Fase 1*
   - Hoje `src/api/v5/routes.go` sempre monta `""` e `"/v5"`, e `src/api/legacy/routes.go` sempre monta `""`, `"/current"` e `"/v4"`.
   - Refatorar isso para separar:
     - rotas explicitamente versionadas, sempre ativas;
     - alias sem versão, montado apenas para a versão escolhida por env.
   - A forma mais segura é introduzir registradores mais granulares, por exemplo:
     - mount da família canônica somente em `"/v5"`;
     - mount da família legada somente em `"/v4"` e `"/current"`;
     - um mount adicional para `""` decidido por `environment.Settings.API.DefaultVersion`.
   - O ponto central dessa orquestração deve ficar em `src/api/api.go`, porque ali já existe a decisão de montagem sob `API_PREFIX`.
4. Fase 4 — Preservar compatibilidade de rotas explícitas. *pode ocorrer em paralelo com parte da Fase 3*
   - Manter sempre funcionando, independentemente do default:
     - `/api/v5/...` (canônica atual)
     - `/api/v4/...` e `/api/current/...` (legado atual)
     - `/api/v3/bot/{token}`
   - Manter também o comportamento existente de `API_PREFIX`, ou seja, a mudança controla somente qual versão responde em `/<prefix>/...`, sem alterar o prefixo configurável.
5. Fase 5 — Migrar o frontend Vue para versão explícita. *depende do desenho de roteamento da Fase 3*
   - Atualizar `src/apps/vuejs/client/src/services/api.ts` para não depender semanticamente do alias `/api/...` quando o objetivo for a API canônica nova.
   - Recomendação: reescrever chamadas canônicas do SPA para `/api/v5/...` antes da substituição por `apiBase`, preservando apenas a resolução de prefixo configurado.
   - Validar se há componentes ou testes que assumem `/api/...` diretamente e ajustar o contrato do frontend para a v5 explícita.
6. Fase 6 — Atualizar testes de roteamento e discovery. *depende das Fases 2–5*
   - Ajustar `src/api/api_route_registration_test.go` para validar:
     - aliases explícitos por versão continuam montados;
     - o alias sem versão muda conforme o env configurado;
     - `/api/v5/...` e `/api/v4/...` permanecem estáveis.
   - Atualizar `src/api/api_environment_discovery_test.go` e/ou `src/api/api_handlers+EnvironmentController_test.go` para verificar `default_api_version` em `preview` e `settings`.
   - Revisar `src/frontend_canonical_routes_test.go` e demais testes que assumem `/api/...` como v5 implícita.
   - Adicionar teste do loader de `APISettings` para o novo env e seu default.
7. Fase 7 — Atualizar documentação operacional. *depende da implementação fechada*
   - Documentar a nova env em `src/environment/README.md`.
   - Atualizar `docs/USAGE-environment-discovery.md` com o novo campo no payload.
   - Atualizar `README.md` e, se necessário, `docs/USAGE-authentication-modes.md` para deixar claro que:
     - `/api/v4/...` e `/api/v5/...` coexistem;
     - `/api/...` aponta para a versão definida por env;
     - o SPA oficial usa `v5` explícita para não ser afetado pelo alias default.
8. Fase 8 — Verificação final.
   - Rodar testes focados de API e roteamento.
   - Validar manualmente exemplos como:
     - `/api/system/version` respondendo pela versão default configurada;
     - `/api/v4/health` permanecendo funcional;
     - `/api/v5/system/version` permanecendo funcional;
     - `/api/system/environment` mostrando `default_api_version` em `preview` e `settings`.

**Relevant files**
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\environment\api_settings.go` — declarar/carregar `API_DEFAULT_VERSION` em `APISettings`.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\environment\environment_settings_preview.go` — expor `default_api_version` no preview público.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api.go` — ponto principal para decidir qual versão recebe o alias sem versão sob `API_PREFIX`.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_handlers_v5.go` — hoje registra a família canônica com alias vazio; provável refatoração para mount explícito e/ou parametrizado.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\v5\routes.go` — separar mount de alias vazio e mount de alias versionado (`/v5`).
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_handlers.go` — família legada atual (`v4`) e possível ponto para expor mount explícito.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\legacy\routes.go` — hoje monta `""`, `"/current"`, `"/v4"`; precisa desacoplar o alias vazio.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\apps\vuejs\client\src\services\api.ts` — migrar SPA para `v5` explícita.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_route_registration_test.go` — cobertura de aliases/versionamento.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_environment_discovery_test.go` — cobertura de `default_api_version` no endpoint de environment.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\api\api_handlers+EnvironmentController_test.go` — asserts adicionais de discovery, se necessário.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\frontend_canonical_routes_test.go` — revisar expectativas do frontend sobre `/api/...`.
- `z:\Desenvolvimento\nocodeleaks-quepasa\src\environment\README.md` — documentar a nova env.
- `z:\Desenvolvimento\nocodeleaks-quepasa\docs\USAGE-environment-discovery.md` — documentar o novo campo exposto.
- `z:\Desenvolvimento\nocodeleaks-quepasa\README.md` — documentar comportamento do alias default da API.

**Verification**
1. Rodar testes de API conforme a convenção do repositório: `cd src/api` e executar testes focados de registro de rota, environment discovery e controllers relacionados.
2. Rodar build do backend em `src` para validar que a reorganização do registro de rotas não quebrou a montagem global.
3. Validar manualmente, com a env apontando para `v4`, que:
   - `/api/system/version` e demais endpoints sem versão caem no comportamento legado esperado;
   - `/api/v5/system/version` continua acessível;
   - `/api/v4/health` continua acessível.
4. Validar manualmente, com a env apontando para `v5`, que o comportamento volta ao padrão canônico atual.
5. Validar o SPA Vue após a migração para `/api/v5/...`, confirmando que ele continua funcional independentemente do valor de `API_DEFAULT_VERSION`.
6. Se houver alteração em payload documentado ou anotações expostas, regenerar/validar Swagger de acordo com a convenção do repositório.

**Decisions**
- Confirmado: o SPA oficial deve ser migrado para usar `v5` explícita, evitando conflitos com o alias `/api/...`.
- Confirmado: a versão default da API deve aparecer tanto em `settings` quanto em `preview` do endpoint de environment.
- Recomendação arquitetural: limitar o alias sem versão configurável a `v4` e `v5`; manter `v3` somente como rota explicitamente versionada.
- Incluído no escopo: coexistência total das rotas explicitamente versionadas.
- Excluído do escopo: remover versões antigas, mudar `API_PREFIX`, ou alterar o contrato explícito de `/api/v3/...`, `/api/v4/...`, `/api/v5/...`.

**Further Considerations**
1. Valor default da nova env: para atender ao rollout de migração, a recomendação é usar `v4` como default inicial em produção, mas isso deve ser decisão consciente porque muda o comportamento histórico atual de `/api/...`.
2. Nomenclatura: `API_DEFAULT_VERSION` comunica melhor a intenção do que algo ligado a “canonical”, porque o alias pode apontar temporariamente para a família legada.
3. Se a equipe quiser endurecer a configuração, é recomendável tratar valores inválidos com fallback claro e log explícito no startup para facilitar suporte operacional.
