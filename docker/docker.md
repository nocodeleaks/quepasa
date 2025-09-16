# QuePasa Docker Installation Guide

## English Version

### Prerequisites
- Docker and Docker Compose installed
- Basic knowledge of environment variables
- Access to configure your domain/network

### Installation Steps

1. **Clone or download the project**
   ```bash
   git clone https://github.com/nocodeleaks/quepasa.git
   cd quepasa
   ```

2. **Configure environment variables**
   ```bash
   cd docker
   cp .env.example .env
   # Edit .env file with your specific configurations
   ```

3. **Edit the `.env` file with your settings**
   - **DOMAIN**: Set your domain (e.g., `quepasa.yourdomain.com`)
   - **EMAIL**: Set your administrator email
   - **MASTERKEY**: Change the default master key for security
   - **QUEPASA_BASIC_AUTH_PASSWORD**: Set a strong password
   - **DBPASSWORD**: Set a secure database password
   - **SIGNING_SECRET**: Change the default signing secret
   - **WEBSOCKETSSL**: Set to `true` if using HTTPS/SSL
   - **LOGLEVEL**: Adjust logging level (ERROR, WARN, INFO, DEBUG, TRACE)
   - **TZ**: Set your timezone

4. **Optional: Review docker-compose.yml**
   - The compose file now uses environment variables from `.env`
   - Includes PostgreSQL database service
   - Configured with health checks and proper networking

5. **Build and run the container**
   ```bash
   # Option 1: Build and run in one command
   docker-compose up -d --build
   
   # Option 2: Build first, then run
   docker-compose build
   docker-compose up -d
   
   # Using newer Docker Compose syntax
   docker compose up -d --build
   ```

6. **Verify installation**
   ```bash
   # Check container status
   docker-compose ps
   
   # Check logs
   docker-compose logs -f quepasa
   ```

### Important Configuration Notes

- **Environment File**: All configurations are now in `.env` file for better management
- **Database**: PostgreSQL service included with automatic setup
- **First Setup**: Set `ACCOUNTSETUP=true` for initial configuration
- **Security**: Change all default passwords and secrets in `.env`
- **SSL**: Set `WEBSOCKETSSL=true` if using HTTPS
- **Network**: Uses internal Docker network `quepasa_network`
- **Ports**: Default port 31000, configurable via `QUEPASA_EXTERNAL_PORT`

### Environment Variables Overview

The `.env` file contains all necessary configurations organized in sections:
- **Basic Config**: Domain, setup flags, master key
- **Authentication**: Email, passwords, auth settings  
- **WhatsApp Features**: Groups, broadcasts, calls, receipts
- **Logging**: Log levels for application and WhatsApp
- **Database**: PostgreSQL connection settings
- **Performance**: Cache, memory, sync settings
- **Debug**: Various debugging options

---

## Versão em Português

### Pré-requisitos
- Docker e Docker Compose instalados
- Conhecimento básico de variáveis de ambiente
- Acesso para configurar seu domínio/rede

### Passos de Instalação

1. **Clone ou baixe o projeto**
   ```bash
   git clone <url-do-repositorio>
   cd quepasa
   ```

2. **Configure as variáveis de ambiente**
   ```bash
   cd docker
   cp .env.example .env
   # Edite o arquivo .env com suas configurações específicas
   ```

3. **Edite o arquivo `.env` com suas configurações**
   - **DOMAIN**: Configure seu domínio (ex: `quepasa.seudominio.com`)
   - **EMAIL**: Defina seu email de administrador
   - **MASTERKEY**: Altere a chave mestra padrão por segurança
   - **QUEPASA_BASIC_AUTH_PASSWORD**: Defina uma senha forte
   - **DBPASSWORD**: Defina uma senha segura para o banco
   - **SIGNING_SECRET**: Altere o segredo de assinatura padrão
   - **WEBSOCKETSSL**: Defina como `true` se usar HTTPS/SSL
   - **LOGLEVEL**: Ajuste o nível de log (ERROR, WARN, INFO, DEBUG, TRACE)
   - **TZ**: Defina seu fuso horário

4. **Opcional: Revise o docker-compose.yml**
   - O arquivo compose agora usa variáveis de ambiente do `.env`
   - Inclui serviço de banco PostgreSQL
   - Configurado com health checks e networking adequado

5. **Construa e execute o container**
   ```bash
   # Opção 1: Construir e executar em um comando
   docker-compose up -d --build
   
   # Opção 2: Construir primeiro, depois executar
   docker-compose build
   docker-compose up -d
   
   # Usando sintaxe mais nova do Docker Compose
   docker compose up -d --build
   ```

6. **Verifique a instalação**
   ```bash
   # Verificar status do container
   docker-compose ps
   
   # Verificar logs
   docker-compose logs -f quepasa
   ```

### Notas Importantes de Configuração

- **Arquivo de Ambiente**: Todas as configurações estão no arquivo `.env` para melhor gestão
- **Banco de Dados**: Serviço PostgreSQL incluído com configuração automática
- **Primeira Configuração**: Defina `ACCOUNTSETUP=true` para configuração inicial
- **Segurança**: Altere todas as senhas e segredos padrão no `.env`
- **SSL**: Defina `WEBSOCKETSSL=true` se usar HTTPS
- **Rede**: Usa rede interna do Docker `quepasa_network`
- **Portas**: Porta padrão 31000, configurável via `QUEPASA_EXTERNAL_PORT`

### Visão Geral das Variáveis de Ambiente

O arquivo `.env` contém todas as configurações necessárias organizadas em seções:
- **Configuração Básica**: Domínio, flags de setup, chave mestra
- **Autenticação**: Email, senhas, configurações de auth
- **Recursos WhatsApp**: Grupos, broadcasts, chamadas, recibos
- **Logging**: Níveis de log para aplicação e WhatsApp
- **Banco de Dados**: Configurações de conexão PostgreSQL
- **Performance**: Cache, memória, configurações de sync
- **Debug**: Várias opções de debugging

---

## Troubleshooting / Solução de Problemas

### Common Issues / Problemas Comuns

#### Database Connection Issues / Problemas de Conexão com Banco
```bash
# Check database container
docker-compose ps

# View database logs
docker-compose logs db
```

#### Permission Issues / Problemas de Permissão
```bash
# Fix volume permissions
docker-compose down
sudo chown -R $USER:$USER quepasa_volume
docker-compose up -d
```

#### Port Conflicts / Conflitos de Porta
- Check if port 31000 is already in use
- Modify `QUEPASA_EXTERNAL_PORT` in .env file
- Update docker-compose.yml port mappings

#### Container Won't Start / Container Não Inicia
```bash
# Check detailed logs
docker-compose logs --details quepasa

# Rebuild without cache
docker-compose build --no-cache
docker-compose up -d
```

### Useful Commands / Comandos Úteis

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Restart specific service
docker-compose restart quepasa

# Access container shell
docker-compose exec quepasa sh

# View real-time logs
docker-compose logs -f --tail=100 quepasa
```

### Health Check / Verificação de Saúde

The container includes a health check endpoint:
```
http://your-domain:31000/healthapi
```

This endpoint should return status information about the QuePasa service.
