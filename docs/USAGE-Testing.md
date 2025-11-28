# Sistema de Testes - QuePasa API

## Visão Geral

O sistema de testes do QuePasa API utiliza bancos de dados SQLite em memória para testes isolados e rápidos. Este sistema permite testar funcionalidades da API sem depender de configurações externas ou bancos de dados reais.

## Arquitetura

### Arquivos Principais

- **`src/api/testing_setup.go`** - Funções helpers para configuração de testes
- **`src/api/api_handlers+*_test.go`** - Arquivos de teste específicos
- **`src/models/qp_database.go`** - Construtores exportados para interfaces de banco

### Componentes

1. **Banco de Dados em Memória** - SQLite `:memory:` para isolamento
2. **Schema Automático** - Criação automática de tabelas necessárias
3. **Helpers de Configuração** - Funções para setup/teardown rápido
4. **Construtores Exportados** - Interfaces de banco acessíveis aos testes

## Como Usar

### 1. Configuração Básica de Teste

```go
package api

import (
    "testing"
    models "github.com/nocodeleaks/quepasa/models"
)

func TestMinhaFuncionalidade(t *testing.T) {
    // Setup do ambiente de teste
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // Seu código de teste aqui...
}
```

### 2. Criando Usuários de Teste

```go
func TestComUsuario(t *testing.T) {
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // Criar usuário de teste
    user := CreateTestUser(t, "testuser", "testpass123")

    // Usar o usuário...
    t.Logf("Usuário criado: %s", user.Username)
}
```

### 3. Criando Servidores de Teste

```go
func TestComServidor(t *testing.T) {
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // Criar usuário e servidor
    user := CreateTestUser(t, "testuser", "testpass123")
    server := CreateTestServer(t, "test-token-123", user.Username)

    // Verificar servidor
    if server.Token != "test-token-123" {
        t.Errorf("Token esperado: test-token-123, recebido: %s", server.Token)
    }
}
```

### 4. Configurando Master Key Temporária

```go
func TestComMasterKey(t *testing.T) {
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // Configurar master key temporária
    testMasterKey := "my-test-master-key"
    cleanup := SetupTestMasterKey(t, testMasterKey)
    defer cleanup()

    // Usar master key nos testes...
}
```

### 5. Teste Completo de API

```go
func TestEndpointCompleto(t *testing.T) {
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // Criar dados de teste
    user := CreateTestUser(t, "apiuser", "apipass")
    server := CreateTestServer(t, "api-token", user.Username)

    // Configurar master key se necessário
    cleanup := SetupTestMasterKey(t, "test-master-key")
    defer cleanup()

    // Fazer requisição HTTP
    req := httptest.NewRequest(http.MethodGet, "/info", nil)
    req.Header.Set("X-QUEPASA-TOKEN", "api-token")
    rec := httptest.NewRecorder()

    // Executar handler
    GetInformationController(rec, req)

    // Verificar resposta
    if rec.Code != http.StatusOK {
        t.Errorf("Status esperado: 200, recebido: %d", rec.Code)
    }

    var response models.QpInfoResponse
    if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
        t.Fatalf("Erro ao decodificar resposta: %v", err)
    }

    if !response.Success {
        t.Error("Resposta deveria ser bem-sucedida")
    }
}
```

## Funções Helpers Disponíveis

### Setup e Cleanup

- **`SetupTestDatabase(t *testing.T) *sqlx.DB`** - Cria banco em memória
- **`SetupTestService(t *testing.T)`** - Inicializa WhatsappService
- **`CleanupTestDatabase(t *testing.T)`** - Limpa banco após teste

### Criação de Dados

- **`CreateTestUser(t *testing.T, username, password string) *models.QpUser`** - Cria usuário
- **`CreateTestServer(t *testing.T, token, username string) *models.QpWhatsappServer`** - Cria servidor
- **`SetupTestMasterKey(t *testing.T, masterKey string) func()`** - Configura master key (retorna cleanup)

### Utilitários

- **`GetTestDataDir(t *testing.T) string`** - Diretório para arquivos temporários
- **`CreateTestDatabase(t *testing.T, name string) *sqlx.DB`** - Banco em arquivo (persistente)

## Padrões de Teste

### Estrutura Básica

```go
func TestNomeDoTeste(t *testing.T) {
    // 1. Setup
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // 2. Arrange (preparar dados)
    user := CreateTestUser(t, "user", "pass")
    server := CreateTestServer(t, "token", user.Username)

    // 3. Act (executar ação)
    req := httptest.NewRequest(http.MethodGet, "/endpoint", nil)
    rec := httptest.NewRecorder()
    HandlerFunction(rec, req)

    // 4. Assert (verificar resultado)
    if rec.Code != http.StatusOK {
        t.Errorf("Erro: esperado 200, recebido %d", rec.Code)
    }
}
```

### Testes com Subtestes

```go
func TestEndpointAutenticacao(t *testing.T) {
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    user := CreateTestUser(t, "user", "pass")
    server := CreateTestServer(t, "token", user.Username)

    t.Run("SemAutenticacao", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/info", nil)
        rec := httptest.NewRecorder()
        GetInformationController(rec, req)

        if rec.Code != http.StatusNoContent {
            t.Errorf("Esperado 204, recebido %d", rec.Code)
        }
    })

    t.Run("ComToken", func(t *testing.T) {
        req := httptest.NewRequest(http.MethodGet, "/info", nil)
        req.Header.Set("X-QUEPASA-TOKEN", "token")
        rec := httptest.NewRecorder()
        GetInformationController(rec, req)

        if rec.Code != http.StatusOK {
            t.Errorf("Esperado 200, recebido %d", rec.Code)
        }
    })
}
```

## Boas Práticas

### 1. Isolamento
- Sempre use `SetupTestService(t)` no início
- Sempre use `defer CleanupTestDatabase(t)` no final
- Cada teste deve ser independente

### 2. Nomenclatura
- Use `TestNomeDaFuncionalidade` para funções de teste
- Use `TestNomeDaFuncionalidade_Cenario` para subtestes
- Prefixe com tipo: `TestAPI_`, `TestModel_`, etc.

### 3. Dados de Teste
- Use nomes únicos para evitar conflitos
- Prefira dados simples e previsíveis
- Documente dados especiais nos comentários

### 4. Verificações
- Verifique códigos HTTP corretos
- Valide estrutura da resposta JSON
- Teste casos de erro e sucesso
- Use `t.Logf()` para debug quando necessário

### 5. Performance
- Testes devem ser rápidos (< 1s por teste)
- Use banco em memória sempre que possível
- Evite sleeps desnecessários

## Executando Testes

### Todos os Testes da API

```bash
cd src/api
go test -v
```

### Teste Específico

```bash
cd src/api
go test -v -run TestInfoEndpoint
```

### Com Cobertura

```bash
cd src/api
go test -v -cover
```

### Com Perfil de Performance

```bash
cd src/api
go test -v -bench=. -benchmem
```

## Exemplos Completos

### Teste de Autenticação

Veja `api_handlers+InformationController_test.go` para exemplo completo de testes de autenticação incluindo:
- Sem autenticação
- Com token de bot
- Com master key (header e query)
- Token inválido
- Prioridade de autenticação

### Exemplos Práticos

Veja `docs/testing_examples.go` para exemplos detalhados de:
- Teste de endpoint de envio de mensagem
- Cenários múltiplos de autenticação
- Validação de dados
- Testes de performance (benchmark)
- Upload de arquivos

### Teste de Endpoint

```go
func TestSendMessage(t *testing.T) {
    SetupTestService(t)
    defer CleanupTestDatabase(t)

    // Criar servidor
    user := CreateTestUser(t, "user", "pass")
    server := CreateTestServer(t, "token", "user")

    // Preparar payload
    payload := models.QpSendRequest{
        ChatId: "5511999999999@c.us",
        Text:   "Mensagem de teste",
    }
    body, _ := json.Marshal(payload)

    // Fazer requisição
    req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-QUEPASA-TOKEN", "token")
    rec := httptest.NewRecorder()

    // Executar
    SendMessageController(rec, req)

    // Verificar
    if rec.Code != http.StatusOK {
        t.Errorf("Erro no envio: %d", rec.Code)
    }
}
```

## Troubleshooting

### Erro: "Database not initialized"
- Certifique-se de chamar `SetupTestService(t)` no início do teste

### Erro: "Server not found"
- Verifique se criou o servidor com `CreateTestServer()`
- Confirme que o token está correto

### Erro: "User not found"
- Use `CreateTestUser()` para criar usuário antes do servidor

### Testes Lentos
- Use banco em memória (`SetupTestService`)
- Evite operações desnecessárias
- Paralelize testes quando possível

### Conflitos de Dados
- Use nomes únicos em cada teste
- Não dependa de estado entre testes
- Use `CleanupTestDatabase()` sempre

## Extensões Futuras

### Testes de Integração
- Adicionar testes com banco real
- Testes de API end-to-end
- Testes de carga

### Helpers Adicionais
- `CreateTestGroup()` - Para testes de grupos
- `CreateTestContact()` - Para testes de contatos
- `MockWhatsAppConnection()` - Para simular conexões

### Relatórios
- Cobertura de código automática
- Relatórios de performance
- Integração com CI/CD

## Contribuição

Para adicionar novos testes:

1. Crie arquivo `*_test.go` no diretório apropriado
2. Use as funções helpers do `testing_setup.go`
3. Siga os padrões estabelecidos
4. Execute `go test -v` para validar
5. Atualize esta documentação se necessário

## Referências

- [Go Testing](https://golang.org/pkg/testing/)
- [httptest](https://golang.org/pkg/net/http/httptest/)
- [sqlx](https://github.com/jmoiron/sqlx)
- [SQLite](https://www.sqlite.org/)
