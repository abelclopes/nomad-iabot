#!/bin/bash

# ==============================================================================
# Nomad Agent - Desinstalador
# ==============================================================================
# Este script remove completamente o Nomad Agent da máquina do usuário
#
# Uso: bash uninstall.sh
# ==============================================================================

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variáveis
INSTALL_DIR="${INSTALL_DIR:-$HOME/nomad-iabot}"
SERVICE_NAME="nomad-agent"

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
# Funções de desinstalação
# ==============================================================================

stop_service() {
    log_info "Parando serviço..."
    
    if command -v systemctl &> /dev/null; then
        if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
            sudo systemctl stop "$SERVICE_NAME" || log_warn "Não foi possível parar o serviço"
            log_success "Serviço parado"
        fi
    fi
    
    # Parar processos manuais
    if pgrep -f "$INSTALL_DIR/nomad" > /dev/null; then
        log_info "Parando processos em execução..."
        pkill -f "$INSTALL_DIR/nomad" || true
        sleep 2
        log_success "Processos parados"
    fi
}

remove_systemd_service() {
    log_info "Removendo serviço systemd..."
    
    if command -v systemctl &> /dev/null; then
        if [ -f "/etc/systemd/system/${SERVICE_NAME}.service" ]; then
            sudo systemctl disable "$SERVICE_NAME" 2>/dev/null || true
            sudo rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
            sudo systemctl daemon-reload
            log_success "Serviço systemd removido"
        else
            log_info "Serviço systemd não encontrado"
        fi
    fi
}

remove_installation_dir() {
    log_info "Removendo diretório de instalação..."
    
    if [ -d "$INSTALL_DIR" ]; then
        # Fazer backup do .env se existir
        if [ -f "$INSTALL_DIR/.env" ]; then
            BACKUP_FILE="$HOME/nomad-agent.env.backup.$(date +%Y%m%d_%H%M%S)"
            cp "$INSTALL_DIR/.env" "$BACKUP_FILE"
            log_info "Backup do .env salvo em: $BACKUP_FILE"
        fi
        
        rm -rf "$INSTALL_DIR"
        log_success "Diretório de instalação removido"
    else
        log_info "Diretório de instalação não encontrado"
    fi
}

# ==============================================================================
# Função principal
# ==============================================================================

main() {
    echo ""
    echo "=========================================="
    echo "  Nomad Agent - Desinstalador"
    echo "=========================================="
    echo ""
    
    log_warn "Este script irá remover completamente o Nomad Agent"
    log_warn "Diretório: $INSTALL_DIR"
    echo ""
    
    read -p "Tem certeza que deseja continuar? (s/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        log_info "Desinstalação cancelada"
        exit 0
    fi
    
    echo ""
    
    # Desinstalação
    stop_service
    remove_systemd_service
    remove_installation_dir
    
    # Mensagem final
    echo ""
    log_success "=========================================="
    log_success "  Desinstalação concluída!"
    log_success "=========================================="
    echo ""
    log_info "O Nomad Agent foi removido completamente"
    
    if [ -f "$HOME/nomad-agent.env.backup."* ]; then
        echo ""
        log_info "Backup da configuração salvo em:"
        ls -1 "$HOME"/nomad-agent.env.backup.* 2>/dev/null | head -n 1
    fi
    
    echo ""
}

# Executar desinstalação
main "$@"
