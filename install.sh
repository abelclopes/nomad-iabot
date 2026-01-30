#!/bin/bash

# ==============================================================================
# Nomad Agent - Instalador Automatizado
# ==============================================================================
# Este script baixa, instala e configura o Nomad Agent na máquina do usuário
#
# Uso: curl -fsSL https://raw.githubusercontent.com/abelclopes/nomad-iabot/main/install.sh | bash
#      ou: bash install.sh
# ==============================================================================

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variáveis de instalação
INSTALL_DIR="${INSTALL_DIR:-$HOME/nomad-iabot}"
REPO_URL="https://github.com/abelclopes/nomad-iabot.git"
SERVICE_NAME="nomad-agent"
GO_MIN_VERSION="1.22"

# ==============================================================================
# Funções auxiliares
# ==============================================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[AVISO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERRO]${NC} $1"
}

# ==============================================================================
# Verificações do sistema
# ==============================================================================

check_os() {
    log_info "Verificando sistema operacional..."
    
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        log_success "Sistema Linux detectado"
        return 0
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        log_success "Sistema macOS detectado"
        return 0
    else
        log_error "Sistema operacional não suportado: $OSTYPE"
        log_info "Este instalador suporta apenas Linux e macOS"
        return 1
    fi
}

check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

check_git() {
    log_info "Verificando Git..."
    
    if check_command git; then
        GIT_VERSION=$(git --version | awk '{print $3}')
        log_success "Git encontrado (versão $GIT_VERSION)"
        return 0
    else
        log_error "Git não está instalado"
        log_info "Instale o Git:"
        log_info "  Ubuntu/Debian: sudo apt-get install git"
        log_info "  CentOS/RHEL: sudo yum install git"
        log_info "  macOS: brew install git"
        return 1
    fi
}

check_go() {
    log_info "Verificando Go..."
    
    if check_command go; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        
        # Comparar versões
        if version_compare "$GO_VERSION" "$GO_MIN_VERSION"; then
            log_success "Go encontrado (versão $GO_VERSION)"
            return 0
        else
            log_error "Go versão $GO_VERSION é muito antiga (mínimo: $GO_MIN_VERSION)"
            return 1
        fi
    else
        log_error "Go não está instalado"
        log_info "Instale o Go $GO_MIN_VERSION ou superior:"
        log_info "  Visite: https://golang.org/dl/"
        log_info "  Ubuntu/Debian: sudo snap install go --classic"
        log_info "  macOS: brew install go"
        return 1
    fi
}

version_compare() {
    local version1=$1
    local version2=$2
    
    # Remove prefixo 'v' se existir
    version1=${version1#v}
    version2=${version2#v}
    
    # Compara as versões
    if [ "$(printf '%s\n' "$version1" "$version2" | sort -V | head -n1)" = "$version2" ]; then
        return 0
    else
        return 1
    fi
}

# ==============================================================================
# Funções de instalação
# ==============================================================================

clone_or_update_repo() {
    log_info "Baixando/atualizando repositório..."
    
    if [ -d "$INSTALL_DIR/.git" ]; then
        log_info "Repositório já existe em $INSTALL_DIR"
        cd "$INSTALL_DIR"
        
        # Salvar alterações locais se existirem
        if ! git diff-index --quiet HEAD -- 2>/dev/null; then
            log_warn "Alterações locais detectadas, fazendo stash..."
            git stash push -m "Auto stash antes da atualização - $(date)"
            log_info "Para recuperar suas alterações: 'cd $INSTALL_DIR && git stash list && git stash pop'"
        fi
        
        log_info "Atualizando repositório..."
        git pull origin main
        log_success "Repositório atualizado"
    else
        log_info "Clonando repositório para $INSTALL_DIR..."
        git clone "$REPO_URL" "$INSTALL_DIR"
        cd "$INSTALL_DIR"
        log_success "Repositório clonado"
    fi
}

configure_env() {
    log_info "Configurando arquivo .env..."
    
    if [ -f "$INSTALL_DIR/.env" ]; then
        log_warn "Arquivo .env já existe"
        read -p "Deseja reconfigurar? (s/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Ss]$ ]]; then
            log_info "Mantendo configuração existente"
            return 0
        fi
    fi
    
    # Copiar .env.example
    cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
    
    echo ""
    log_info "=========================================="
    log_info "Configuração do Nomad Agent"
    log_info "=========================================="
    echo ""
    
    # Porta do Gateway
    read -p "Porta do Gateway (padrão: 8080): " GATEWAY_PORT
    GATEWAY_PORT=${GATEWAY_PORT:-8080}
    
    # Portable sed across Linux and macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/GATEWAY_PORT=.*/GATEWAY_PORT=$GATEWAY_PORT/" "$INSTALL_DIR/.env"
    else
        sed -i "s/GATEWAY_PORT=.*/GATEWAY_PORT=$GATEWAY_PORT/" "$INSTALL_DIR/.env"
    fi
    
    # Provedor LLM
    echo ""
    log_info "Provedores LLM disponíveis:"
    echo "  1) ollama (recomendado)"
    echo "  2) lmstudio"
    echo "  3) localai"
    echo "  4) vllm"
    echo "  5) openai"
    read -p "Escolha o provedor LLM (1-5, padrão: 1): " LLM_CHOICE
    LLM_CHOICE=${LLM_CHOICE:-1}
    
    case $LLM_CHOICE in
        1) LLM_PROVIDER="ollama"; LLM_DEFAULT_URL="http://localhost:11434" ;;
        2) LLM_PROVIDER="lmstudio"; LLM_DEFAULT_URL="http://localhost:1234" ;;
        3) LLM_PROVIDER="localai"; LLM_DEFAULT_URL="http://localhost:8080" ;;
        4) LLM_PROVIDER="vllm"; LLM_DEFAULT_URL="http://localhost:8000" ;;
        5) LLM_PROVIDER="openai"; LLM_DEFAULT_URL="https://api.openai.com/v1" ;;
        *) LLM_PROVIDER="ollama"; LLM_DEFAULT_URL="http://localhost:11434" ;;
    esac
    
    read -p "URL do LLM (padrão: $LLM_DEFAULT_URL): " LLM_BASE_URL
    LLM_BASE_URL=${LLM_BASE_URL:-$LLM_DEFAULT_URL}
    
    read -p "Modelo LLM (padrão: qwen3:latest): " LLM_MODEL
    LLM_MODEL=${LLM_MODEL:-qwen3:latest}
    
    # Portable sed across Linux and macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|LLM_PROVIDER=.*|LLM_PROVIDER=$LLM_PROVIDER|" "$INSTALL_DIR/.env"
        sed -i '' "s|LLM_BASE_URL=.*|LLM_BASE_URL=$LLM_BASE_URL|" "$INSTALL_DIR/.env"
        sed -i '' "s|LLM_MODEL=.*|LLM_MODEL=$LLM_MODEL|" "$INSTALL_DIR/.env"
    else
        sed -i "s|LLM_PROVIDER=.*|LLM_PROVIDER=$LLM_PROVIDER|" "$INSTALL_DIR/.env"
        sed -i "s|LLM_BASE_URL=.*|LLM_BASE_URL=$LLM_BASE_URL|" "$INSTALL_DIR/.env"
        sed -i "s|LLM_MODEL=.*|LLM_MODEL=$LLM_MODEL|" "$INSTALL_DIR/.env"
    fi
    
    # JWT Secret
    echo ""
    JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)
    
    # Validate JWT secret
    if [ -z "$JWT_SECRET" ] || [ ${#JWT_SECRET} -lt 32 ]; then
        log_error "Falha ao gerar JWT secret"
        return 1
    fi
    
    log_info "Gerando JWT secret aleatório..."
    
    # Portable sed across Linux and macOS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" "$INSTALL_DIR/.env"
    else
        sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" "$INSTALL_DIR/.env"
    fi
    
    # Azure DevOps (opcional)
    echo ""
    read -p "Configurar Azure DevOps? (s/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        read -s -p "Azure DevOps PAT (entrada oculta): " AZURE_DEVOPS_PAT
        echo
        read -p "Azure DevOps Organization: " AZURE_DEVOPS_ORGANIZATION
        read -p "Azure DevOps Project: " AZURE_DEVOPS_PROJECT
        
        # Portable sed across Linux and macOS
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|AZURE_DEVOPS_PAT=.*|AZURE_DEVOPS_PAT=$AZURE_DEVOPS_PAT|" "$INSTALL_DIR/.env"
            sed -i '' "s|AZURE_DEVOPS_ORGANIZATION=.*|AZURE_DEVOPS_ORGANIZATION=$AZURE_DEVOPS_ORGANIZATION|" "$INSTALL_DIR/.env"
            sed -i '' "s|AZURE_DEVOPS_PROJECT=.*|AZURE_DEVOPS_PROJECT=$AZURE_DEVOPS_PROJECT|" "$INSTALL_DIR/.env"
        else
            sed -i "s|AZURE_DEVOPS_PAT=.*|AZURE_DEVOPS_PAT=$AZURE_DEVOPS_PAT|" "$INSTALL_DIR/.env"
            sed -i "s|AZURE_DEVOPS_ORGANIZATION=.*|AZURE_DEVOPS_ORGANIZATION=$AZURE_DEVOPS_ORGANIZATION|" "$INSTALL_DIR/.env"
            sed -i "s|AZURE_DEVOPS_PROJECT=.*|AZURE_DEVOPS_PROJECT=$AZURE_DEVOPS_PROJECT|" "$INSTALL_DIR/.env"
        fi
    fi
    
    # Telegram Bot (opcional)
    echo ""
    read -p "Configurar Telegram Bot? (s/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Ss]$ ]]; then
        read -s -p "Telegram Bot Token (entrada oculta): " TELEGRAM_BOT_TOKEN
        echo
        read -p "Telegram Allowed Users (separados por vírgula): " TELEGRAM_ALLOWED_USERS
        
        # Portable sed across Linux and macOS
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|TELEGRAM_BOT_TOKEN=.*|TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN|" "$INSTALL_DIR/.env"
            sed -i '' "s|TELEGRAM_ALLOWED_USERS=.*|TELEGRAM_ALLOWED_USERS=$TELEGRAM_ALLOWED_USERS|" "$INSTALL_DIR/.env"
        else
            sed -i "s|TELEGRAM_BOT_TOKEN=.*|TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN|" "$INSTALL_DIR/.env"
            sed -i "s|TELEGRAM_ALLOWED_USERS=.*|TELEGRAM_ALLOWED_USERS=$TELEGRAM_ALLOWED_USERS|" "$INSTALL_DIR/.env"
        fi
    fi
    
    log_success "Arquivo .env configurado"
}

build_binary() {
    log_info "Compilando binário..."
    
    cd "$INSTALL_DIR"
    
    # Build
    go build -o nomad ./cmd/nomad
    
    if [ -f "$INSTALL_DIR/nomad" ]; then
        chmod +x "$INSTALL_DIR/nomad"
        log_success "Binário compilado com sucesso"
        return 0
    else
        log_error "Falha ao compilar binário"
        return 1
    fi
}

create_systemd_service() {
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        log_info "Systemd service não disponível no macOS"
        return 0
    fi
    
    echo ""
    read -p "Deseja criar um serviço systemd para auto-start? (s/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        log_info "Pulando criação do serviço systemd"
        return 0
    fi
    
    log_info "Criando serviço systemd..."
    
    # Criar arquivo de serviço
    SERVICE_FILE="/tmp/${SERVICE_NAME}.service"
    
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Nomad Agent - AI Assistant
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/nomad
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    # Copiar para systemd e habilitar
    sudo cp "$SERVICE_FILE" "/etc/systemd/system/${SERVICE_NAME}.service"
    sudo systemctl daemon-reload
    sudo systemctl enable "$SERVICE_NAME"
    
    rm -f "$SERVICE_FILE"
    
    log_success "Serviço systemd criado e habilitado"
    log_info "Use 'sudo systemctl start $SERVICE_NAME' para iniciar"
    log_info "Use 'sudo systemctl status $SERVICE_NAME' para verificar status"
}

create_launcher_scripts() {
    log_info "Criando scripts de lançamento..."
    
    # Script de start
    cat > "$INSTALL_DIR/start.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
./nomad
EOF
    chmod +x "$INSTALL_DIR/start.sh"
    
    # Script de stop (para systemd ou processo manual)
    cat > "$INSTALL_DIR/stop.sh" << EOF
#!/bin/bash
INSTALL_DIR="$INSTALL_DIR"
if command -v systemctl &> /dev/null && systemctl is-active --quiet nomad-agent; then
    sudo systemctl stop nomad-agent
elif pgrep -f "\$INSTALL_DIR/nomad" > /dev/null; then
    # Kill specific nomad process using full path
    pkill -f "\$INSTALL_DIR/nomad"
    echo "Nomad process stopped"
else
    echo "No nomad process found"
fi
EOF
    chmod +x "$INSTALL_DIR/stop.sh"
    
    log_success "Scripts de lançamento criados"
}

verify_installation() {
    log_info "Verificando instalação..."
    
    # Verificar binário
    if [ ! -f "$INSTALL_DIR/nomad" ]; then
        log_error "Binário não encontrado"
        return 1
    fi
    
    # Verificar .env
    if [ ! -f "$INSTALL_DIR/.env" ]; then
        log_error "Arquivo .env não encontrado"
        return 1
    fi
    
    # Verificar que o binário pode mostrar ajuda (teste mais significativo)
    cd "$INSTALL_DIR"
    if ./nomad -h > /dev/null 2>&1 || ./nomad --help > /dev/null 2>&1; then
        log_success "Binário funcional"
    fi
    
    log_success "Instalação verificada"
}

# ==============================================================================
# Função principal
# ==============================================================================

main() {
    echo ""
    echo "=========================================="
    echo "  Nomad Agent - Instalador"
    echo "=========================================="
    echo ""
    
    # Verificações do sistema
    check_os || exit 1
    check_git || exit 1
    check_go || exit 1
    
    echo ""
    
    # Instalação
    clone_or_update_repo || exit 1
    configure_env || exit 1
    build_binary || exit 1
    create_launcher_scripts || exit 1
    create_systemd_service
    verify_installation || exit 1
    
    # Mensagem final
    echo ""
    log_success "=========================================="
    log_success "  Instalação concluída com sucesso!"
    log_success "=========================================="
    echo ""
    log_info "Diretório de instalação: $INSTALL_DIR"
    log_info ""
    log_info "Para iniciar o Nomad Agent:"
    
    if [[ "$OSTYPE" == "linux-gnu"* ]] && systemctl is-enabled "$SERVICE_NAME" &>/dev/null; then
        log_info "  sudo systemctl start $SERVICE_NAME"
        log_info ""
        log_info "Para verificar o status:"
        log_info "  sudo systemctl status $SERVICE_NAME"
        log_info ""
        log_info "Para ver os logs:"
        log_info "  sudo journalctl -u $SERVICE_NAME -f"
    else
        log_info "  cd $INSTALL_DIR"
        log_info "  ./start.sh"
        log_info ""
        log_info "Ou diretamente:"
        log_info "  cd $INSTALL_DIR"
        log_info "  ./nomad"
    fi
    
    echo ""
    log_info "Para testar a instalação:"
    GATEWAY_PORT=$(grep "^GATEWAY_PORT=" "$INSTALL_DIR/.env" 2>/dev/null | cut -d= -f2 | tr -d ' ')
    GATEWAY_PORT=${GATEWAY_PORT:-8080}
    log_info "  curl http://localhost:${GATEWAY_PORT}/health"
    echo ""
    log_info "Documentação completa: $INSTALL_DIR/README.md"
    echo ""
}

# Executar instalação
main "$@"
