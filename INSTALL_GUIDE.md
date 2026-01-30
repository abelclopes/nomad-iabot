# Guia de Instalação do Nomad Agent

## Instalação Rápida

### Opção 1: Instalação Automática (Recomendada)

```bash
# Clone o repositório
git clone https://github.com/abelclopes/nomad-iabot.git
cd nomad-iabot

# Execute o instalador
bash install.sh
```

O instalador irá guiá-lo através de um processo interativo que inclui:

1. **Verificação de pré-requisitos**
   - Go 1.22 ou superior
   - Git

2. **Configuração do ambiente**
   - Porta do gateway (padrão: 8080)
   - Provedor LLM (Ollama, LM Studio, LocalAI, vLLM, OpenAI)
   - URL do LLM
   - Modelo LLM
   - JWT secret (gerado automaticamente)
   - Azure DevOps (opcional)
   - Telegram Bot (opcional)

3. **Compilação e instalação**
   - Download das dependências Go
   - Compilação do binário
   - Criação de scripts auxiliares
   - Configuração de serviço systemd (Linux, opcional)

### Opção 2: Instalação via curl

⚠️ **AVISO**: Este método baixa e executa o script diretamente. Revise sempre o script antes de executá-lo desta forma.

```bash
curl -fsSL https://raw.githubusercontent.com/abelclopes/nomad-iabot/main/install.sh | bash
```

## Exemplo de Uso

### Instalação Completa

```bash
$ bash install.sh

==========================================
  Nomad Agent - Instalador
==========================================

[INFO] Verificando sistema operacional...
[OK] Sistema Linux detectado
[INFO] Verificando Git...
[OK] Git encontrado (versão 2.40.0)
[INFO] Verificando Go...
[OK] Go encontrado (versão 1.22.0)

[INFO] Clonando repositório...
[OK] Repositório clonado

[INFO] Configurando arquivo .env...
==========================================
Configuração do Nomad Agent
==========================================

Porta do Gateway (padrão: 8080): 8080

Provedores LLM disponíveis:
  1) ollama (recomendado)
  2) lmstudio
  3) localai
  4) vllm
  5) openai
Escolha o provedor LLM (1-5, padrão: 1): 1

URL do LLM (padrão: http://localhost:11434): 
Modelo LLM (padrão: qwen3:latest): llama3.2

[INFO] Gerando JWT secret aleatório...
[OK] Arquivo .env configurado

[INFO] Compilando binário...
[OK] Binário compilado com sucesso

[INFO] Criando scripts de lançamento...
[OK] Scripts de lançamento criados

Deseja criar um serviço systemd para auto-start? (s/N): s
[INFO] Criando serviço systemd...
[OK] Serviço systemd criado e habilitado

[INFO] Verificando instalação...
[OK] Instalação verificada

==========================================
  Instalação concluída com sucesso!
==========================================

Para iniciar o Nomad Agent:
  sudo systemctl start nomad-agent

Para verificar o status:
  sudo systemctl status nomad-agent

Para testar a instalação:
  curl http://localhost:8080/health
```

## Iniciando o Nomad Agent

### Com systemd (Linux)

```bash
# Iniciar o serviço
sudo systemctl start nomad-agent

# Verificar status
sudo systemctl status nomad-agent

# Ver logs
sudo journalctl -u nomad-agent -f

# Parar o serviço
sudo systemctl stop nomad-agent
```

### Manualmente

```bash
# Usando o script auxiliar
cd $HOME/nomad-iabot
./start.sh

# Ou diretamente
cd $HOME/nomad-iabot
./nomad
```

## Testando a Instalação

```bash
# Health check
curl http://localhost:8080/health

# Enviar mensagem de teste
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Olá! Como você pode me ajudar?"}'
```

## Desinstalação

Para remover completamente o Nomad Agent:

```bash
cd $HOME/nomad-iabot
bash uninstall.sh
```

O desinstalador irá:
- Parar todos os processos em execução
- Remover o serviço systemd
- Remover o diretório de instalação
- Criar um backup do arquivo .env

## Atualizando

Para atualizar para a versão mais recente:

```bash
cd $HOME/nomad-iabot
bash install.sh
```

O instalador detectará a instalação existente e oferecerá a opção de atualizar ou reconfigurar.

## Localização Personalizada

Por padrão, o Nomad Agent é instalado em `$HOME/nomad-iabot`. Para usar um diretório diferente:

```bash
export INSTALL_DIR=/caminho/customizado
bash install.sh
```

## Solução de Problemas

### Go não está instalado

```bash
# Ubuntu/Debian
sudo snap install go --classic

# CentOS/RHEL
sudo yum install golang

# macOS
brew install go
```

### Git não está instalado

```bash
# Ubuntu/Debian
sudo apt-get install git

# CentOS/RHEL
sudo yum install git

# macOS
brew install git
```

### Erro de compilação

Certifique-se de que:
1. A versão do Go é 1.22 ou superior: `go version`
2. As variáveis de ambiente do Go estão configuradas: `echo $GOPATH`
3. Você tem conexão com a internet para baixar dependências

### Porta já em uso

Se a porta 8080 já estiver em uso, escolha uma porta diferente durante a configuração ou edite o arquivo `.env`:

```bash
cd $HOME/nomad-iabot
nano .env
# Altere GATEWAY_PORT=8080 para outra porta
```

## Suporte

Para mais informações, consulte:
- [README.md](README.md) - Documentação completa
- [Issues no GitHub](https://github.com/abelclopes/nomad-iabot/issues) - Relatar problemas
