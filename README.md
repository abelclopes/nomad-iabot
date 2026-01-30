# Nomad Agent 🤖

Um assistente AI seguro e modular com foco em APIs locais e integração com Azure DevOps e Trello.

## 🚀 Funcionalidades

- **LLM Local**: Suporte a Ollama, LM Studio, LocalAI, vLLM
- **Azure DevOps**: Gerenciamento completo de Work Items, Pipelines, Repos e Boards
- **Trello**: Gerenciamento de boards, listas e cards
- **Multi-Canal**: WebChat e Telegram
- **Segurança**: JWT Auth, Rate Limiting, Allowlist de usuários
- **Docker-First**: Build otimizado ~15MB

## 📋 Pré-requisitos

- Go 1.22+ (para desenvolvimento local)
- Docker & Docker Compose
- Ollama ou outro servidor LLM local
- Azure DevOps PAT (opcional)
- Trello API Key e Token (opcional)

## ⚡ Início Rápido

### Instalação Automática (Recomendado)

Use o instalador automático que baixa, configura e instala o Nomad Agent:

```bash
# Opção 1: Clone e execute localmente (mais seguro)
git clone https://github.com/abelclopes/nomad-iabot.git
cd nomad-iabot
bash install.sh

# Opção 2: Download e execução direta
# ⚠️ AVISO: Revise o script antes de executar com curl | bash
curl -fsSL https://raw.githubusercontent.com/abelclopes/nomad-iabot/main/install.sh | bash
```

O instalador irá:
- ✅ Verificar dependências (Go, Git)
- ✅ Baixar/atualizar o repositório
- ✅ Configurar o arquivo `.env` interativamente
- ✅ Compilar o binário
- ✅ Criar serviço systemd (Linux, opcional)
- ✅ Gerar JWT secret automaticamente

### Instalação Manual

### 1. Clone e Configure

```bash
git clone https://github.com/abelclopes/nomad-iabot.git
cd nomad-iabot
cp .env.example .env
```

### 2. Configure o .env

Edite `.env` com suas configurações:

```env
# LLM
LLM_PROVIDER=ollama
LLM_BASE_URL=http://localhost:11434
LLM_MODEL=llama3.2

# Segurança
JWT_SECRET=sua-chave-secreta-aqui

# Azure DevOps (opcional)
AZURE_DEVOPS_PAT=seu-pat-aqui
AZURE_DEVOPS_ORGANIZATION=sua-org
AZURE_DEVOPS_PROJECT=seu-projeto

# Trello (opcional)
TRELLO_API_KEY=sua-api-key-aqui
TRELLO_TOKEN=seu-token-aqui
```

### 3. Execute

**Instalação automática:**
```bash
# Se instalou via instalador automático
sudo systemctl start nomad-agent  # Linux com systemd
# ou
cd $HOME/nomad-iabot
./start.sh
```

**Com Docker:**
```bash
docker-compose up -d
```

**Desenvolvimento local:**
```bash
go run ./cmd/nomad
```

### 4. Teste

```bash
# Health check
curl http://localhost:8080/health

# Chat (sem auth)
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Olá! Me fale sobre você."}'
```

## 🏗️ Estrutura do Projeto

```
nomad-agent/
├── cmd/
│   └── nomad/          # Entry point
├── internal/
│   ├── agent/          # Core do agente AI
│   ├── channels/       # Canais (Telegram, WebChat)
│   ├── config/         # Configurações
│   ├── devops/         # Azure DevOps integration
│   ├── trello/         # Trello integration
│   ├── gateway/        # HTTP server & handlers
│   └── llm/            # Cliente LLM
├── skills/             # Agent Skills - Configuração de segurança
│   ├── README.md       # Documentação dos skills
│   ├── azure_devops_skills.md
│   ├── telegram_skills.md
│   ├── webchat_skills.md
│   └── llm_skills.md
├── web/
│   └── dist/           # WebChat frontend (TODO)
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## 🔧 Configuração

### Desinstalação

Para desinstalar completamente o Nomad Agent:

```bash
cd $HOME/nomad-iabot
bash uninstall.sh
```

O desinstalador irá:
- ⛔ Parar o serviço e processos em execução
- 🗑️ Remover o serviço systemd (se existir)
- 📁 Remover o diretório de instalação
- 💾 Fazer backup do arquivo `.env`

### Provedores LLM Suportados

| Provedor | URL Padrão | Notas |
|----------|------------|-------|
| Ollama | `http://localhost:11434` | Recomendado |
| LM Studio | `http://localhost:1234` | OpenAI-compatible |
| LocalAI | `http://localhost:8080` | OpenAI-compatible |
| vLLM | `http://localhost:8000` | OpenAI-compatible |

### Azure DevOps

Crie um PAT em: `https://dev.azure.com/{org}/_usersSettings/tokens`

Permissões necessárias:
- **Work Items**: Read & Write
- **Code**: Read
- **Build**: Read & Execute
- **Project and Team**: Read

### Trello

1. Obtenha sua API Key em: `https://trello.com/app-key`
2. Gere um Token clicando em "Token" na mesma página
3. Configure no `.env`:
   ```env
   TRELLO_API_KEY=sua-api-key
   TRELLO_TOKEN=seu-token
   ```

Permissões do Token:
- O token precisa ter acesso de leitura e escrita aos boards que você deseja gerenciar

### Telegram Bot

1. Fale com [@BotFather](https://t.me/BotFather)
2. Crie um bot com `/newbot`
3. Copie o token para `TELEGRAM_BOT_TOKEN`
4. Adicione seu ID em `TELEGRAM_ALLOWED_USERS`

## 📡 API Reference

### Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| GET | `/health` | Health check |
| POST | `/api/v1/chat` | Enviar mensagem |
| GET | `/api/v1/tools` | Listar ferramentas |
| POST | `/api/v1/devops/workitems` | Criar work item |
| GET | `/api/v1/devops/workitems/{id}` | Buscar work item |
| POST | `/api/v1/devops/workitems/query` | Query WIQL |

### Exemplo de Chat

```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "message": "Liste os bugs abertos do projeto",
    "session_id": "session-123"
  }'
```

## 🔐 Segurança

- **JWT Auth**: Tokens assinados com HS256
- **Rate Limiting**: 100 requests/min por IP
- **Input Validation**: Sanitização de entrada
- **No Shell Execution**: Ferramentas sandboxadas
- **User Allowlist**: Controle de acesso no Telegram
- **Agent Skills**: Configuração de segurança por integração (veja `skills/`)

### Agent Skills Pattern

O projeto utiliza o padrão **Agent Skills** para garantir que o agente opere apenas dentro dos limites seguros e definidos, prevenindo falhas de segurança como prompt injection.

Cada integração possui um arquivo de skill que documenta:
- ✅ Operações permitidas
- ❌ Operações proibidas
- 🔒 Regras de segurança
- 📝 Exemplos de uso

**Skills disponíveis:**
- `skills/azure_devops_skills.md` - Operações do Azure DevOps
- `skills/telegram_skills.md` - Operações do Telegram
- `skills/webchat_skills.md` - Operações do WebChat
- `skills/llm_skills.md` - Operações do LLM

Para mais informações, consulte [skills/README.md](skills/README.md).

## 🐳 Docker

### Build Manual

```bash
docker build -t nomad-agent .
```

### Com Ollama incluído

```bash
docker-compose --profile with-ollama up -d
```

### Apenas o Agent

```bash
docker-compose up -d nomad-agent
```

## 📝 Roadmap

- [ ] WebChat frontend
- [ ] WebSocket streaming
- [ ] Histórico de conversas persistente
- [ ] Mais integrações (GitHub, GitLab)
- [ ] Plugin system
- [ ] Multi-tenancy

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'Add nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📄 Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.
