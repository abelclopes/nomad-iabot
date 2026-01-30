---
name: "Telegram Channel"
description: "Skill for Telegram bot integration and message handling"
version: "1.0.0"
integration: "telegram"
security_level: "high"
---

# Telegram Channel Skills

## Objetivo
Este skill define as operações permitidas para a integração com Telegram, garantindo que o bot opere de forma segura e controlada.

## Operações Permitidas

### Mensagens

#### 1. Receber Mensagens de Texto
- **Descrição**: Processar mensagens de texto enviadas por usuários
- **Restrições**: 
  - Apenas usuários na allowlist podem interagir
  - Verificar ID do usuário antes de processar
  - Limitar tamanho da mensagem
- **Exemplo**: Usuário envia "Olá, como você está?"

#### 2. Enviar Mensagens de Resposta
- **Descrição**: Enviar respostas formatadas aos usuários
- **Restrições**: 
  - Máximo 4096 caracteres por mensagem (limite do Telegram)
  - Usar formatação Markdown quando apropriado
  - Não enviar dados sensíveis
- **Exemplo**: Bot responde "Olá! Posso ajudar com Azure DevOps."

### Comandos

#### 3. Comando /start
- **Descrição**: Inicializar conversa com o bot
- **Resposta**: Mensagem de boas-vindas
- **Restrições**: Disponível para todos os usuários na allowlist
- **Exemplo**: `/start`

#### 4. Comando /help
- **Descrição**: Mostrar comandos disponíveis
- **Resposta**: Lista de comandos e suas descrições
- **Restrições**: Disponível para todos os usuários na allowlist
- **Exemplo**: `/help`

#### 5. Comando /status
- **Descrição**: Verificar status do sistema
- **Resposta**: Status do agente e integrações
- **Restrições**: Apenas usuários autorizados
- **Exemplo**: `/status`

#### 6. Comando /workitems
- **Descrição**: Listar work items do Azure DevOps
- **Resposta**: Lista de work items do usuário
- **Restrições**: Requer integração com Azure DevOps configurada
- **Exemplo**: `/workitems`

## Regras de Segurança

### Prevenção de Prompt Injection
1. **User Allowlist**: Apenas IDs de usuário autorizados podem interagir
2. **Validação de Entrada**: Sanitizar todas as mensagens recebidas
3. **Rate Limiting**: Limitar número de mensagens por usuário
4. **Timeout**: Timeout de 30 segundos para processar mensagens
5. **Sanitização**: Remover comandos shell e caracteres especiais

### Controle de Acesso
- ✅ Verificar user_id contra allowlist em TODAS as mensagens
- ✅ Log de todas as interações para auditoria
- ✅ Bloquear automaticamente usuários não autorizados
- ✅ Notificar administradores sobre tentativas de acesso não autorizado

### Operações NÃO Permitidas
- ❌ Processar mensagens de usuários não autorizados
- ❌ Enviar mensagens para outros chats sem permissão
- ❌ Executar comandos do sistema operacional
- ❌ Acessar arquivos do sistema
- ❌ Modificar configurações do bot
- ❌ Adicionar/remover usuários da allowlist dinamicamente
- ❌ Enviar mensagens em massa (spam)
- ❌ Fazer forward de mensagens sem autorização

## Tratamento de Mensagens

### Fluxo de Processamento
1. Receber mensagem do Telegram
2. Verificar user_id contra allowlist
3. Log da mensagem recebida
4. Sanitizar entrada
5. Processar com o agente AI
6. Formatar resposta
7. Enviar resposta
8. Log da resposta enviada

### Formato de Mensagens
- Usar Markdown para formatação
- Quebrar mensagens longas em múltiplas mensagens
- Adicionar emojis para melhor UX
- Incluir comandos de ação quando relevante

## Exemplo de Uso

```
Usuário (ID: 123456): "/start"
Bot: "👋 Olá! Eu sou o Nomad Agent. Como posso ajudar?"

Usuário (ID: 123456): "Liste meus work items"
Bot: [Processa via agent.ProcessMessage]
Bot: "📋 Encontrei 3 work items:
- #123 [Task] Implementar login (Active)
- #124 [Bug] Corrigir validação (New)
- #125 [Story] Nova feature (Active)"

Usuário não autorizado (ID: 999999): "Olá"
Bot: [Mensagem bloqueada, não processa]
Log: "Blocked unauthorized user: 999999"
```

## Configuração Necessária

Para usar este skill, as seguintes variáveis de ambiente devem estar configuradas:
- `TELEGRAM_BOT_TOKEN`: Token do bot obtido do BotFather
- `TELEGRAM_ALLOWED_USERS`: Lista de user IDs autorizados (separados por vírgula)
- `TELEGRAM_TIMEOUT_SEC`: Timeout para processar mensagens (padrão: 30)

## Obtenção do Token
1. Falar com [@BotFather](https://t.me/BotFather) no Telegram
2. Enviar comando `/newbot`
3. Seguir instruções para criar o bot
4. Copiar o token fornecido

## Obtenção do User ID
1. Falar com [@userinfobot](https://t.me/userinfobot)
2. O bot retornará seu user ID
3. Adicionar o ID em `TELEGRAM_ALLOWED_USERS`

## Limitações do Telegram
- Máximo 4096 caracteres por mensagem
- Rate limit de 30 mensagens por segundo (global)
- Rate limit de 1 mensagem por segundo por chat
- Arquivos até 50MB (bots)
