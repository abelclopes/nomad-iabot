---
name: "WebChat Channel"
description: "Skill for web-based chat interface integration"
version: "1.0.0"
integration: "webchat"
security_level: "medium"
---

# WebChat Channel Skills

## Objetivo
Este skill define as operações permitidas para a integração do WebChat, garantindo que a interface web opere de forma segura e controlada.

## Operações Permitidas

### Mensagens

#### 1. Receber Mensagens via HTTP POST
- **Endpoint**: `/api/v1/chat`
- **Método**: POST
- **Descrição**: Processar mensagens enviadas pelo frontend web
- **Parâmetros**:
  - `message` (obrigatório): Texto da mensagem
  - `session_id` (opcional): ID da sessão
  - `user_id` (opcional): ID do usuário
- **Restrições**: 
  - Validar token JWT se autenticação estiver habilitada
  - Limitar tamanho da mensagem (max 10KB)
  - Rate limiting por IP
- **Exemplo**: 
  ```json
  {
    "message": "Liste meus work items",
    "session_id": "session-123",
    "user_id": "user@example.com"
  }
  ```

#### 2. Enviar Respostas via HTTP
- **Descrição**: Retornar respostas processadas para o frontend
- **Formato**: JSON
- **Restrições**: 
  - Máximo 100KB por resposta
  - Sanitizar HTML/JavaScript malicioso
  - Não incluir dados sensíveis
- **Exemplo**:
  ```json
  {
    "response": "Encontrei 3 work items...",
    "session_id": "session-123",
    "timestamp": "2024-01-30T12:00:00Z"
  }
  ```

### Autenticação

#### 3. Validação de Token JWT
- **Descrição**: Validar token JWT em requisições autenticadas
- **Header**: `Authorization: Bearer <token>`
- **Restrições**: 
  - Token deve ser válido e não expirado
  - Verificar assinatura com JWT_SECRET
  - Rejeitar tokens malformados
- **Exemplo**: `Authorization: Bearer eyJhbGc...`

### Health Check

#### 4. Health Check
- **Endpoint**: `/health`
- **Método**: GET
- **Descrição**: Verificar status do sistema
- **Restrições**: Endpoint público (sem autenticação)
- **Exemplo**: `GET /health` → `{"status": "ok"}`

### Ferramentas

#### 5. Listar Ferramentas Disponíveis
- **Endpoint**: `/api/v1/tools`
- **Método**: GET
- **Descrição**: Listar ferramentas disponíveis para o agente
- **Restrições**: Requer autenticação
- **Exemplo**: Retorna lista de tools disponíveis

## Regras de Segurança

### Prevenção de Prompt Injection
1. **Input Sanitization**: Sanitizar todas as mensagens de entrada
2. **Rate Limiting**: 100 requisições/minuto por IP
3. **JWT Authentication**: Validar tokens em endpoints protegidos
4. **CORS**: Configurar CORS apropriadamente
5. **Content Security Policy**: Headers de segurança apropriados

### Controle de Acesso
- ✅ Verificar JWT em endpoints protegidos
- ✅ Log de todas as requisições para auditoria
- ✅ Rate limiting por IP address
- ✅ Timeout de 30 segundos por requisição
- ✅ Validar Content-Type (application/json)

### Operações NÃO Permitidas
- ❌ Processar requisições sem rate limiting
- ❌ Executar JavaScript do cliente no servidor
- ❌ Acessar arquivos do sistema
- ❌ Modificar configurações do servidor
- ❌ Fazer requisições para URLs externas arbitrárias
- ❌ Expor tokens ou credenciais em respostas
- ❌ Aceitar uploads de arquivos não sanitizados
- ❌ Executar comandos do sistema operacional

## API Endpoints

### POST /api/v1/chat
```http
POST /api/v1/chat HTTP/1.1
Content-Type: application/json
Authorization: Bearer <token>

{
  "message": "string",
  "session_id": "string (optional)",
  "user_id": "string (optional)"
}
```

**Response:**
```json
{
  "response": "string",
  "session_id": "string",
  "timestamp": "ISO 8601 datetime"
}
```

### GET /health
```http
GET /health HTTP/1.1
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": 12345
}
```

### GET /api/v1/tools
```http
GET /api/v1/tools HTTP/1.1
Authorization: Bearer <token>
```

**Response:**
```json
{
  "tools": [
    {
      "name": "devops_list_my_workitems",
      "description": "List Azure DevOps work items..."
    }
  ]
}
```

## Tratamento de Erros

### Códigos de Status HTTP
- `200 OK`: Sucesso
- `400 Bad Request`: Parâmetros inválidos
- `401 Unauthorized`: Token inválido ou ausente
- `429 Too Many Requests`: Rate limit excedido
- `500 Internal Server Error`: Erro no servidor

### Formato de Erro
```json
{
  "error": "string",
  "code": "ERROR_CODE",
  "message": "Descrição amigável do erro"
}
```

## Exemplo de Uso

```bash
# Chat sem autenticação (se permitido)
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Olá! Me fale sobre você."}'

# Chat com autenticação
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "message": "Liste meus work items",
    "session_id": "session-123"
  }'

# Health check
curl http://localhost:8080/health
```

## Configuração Necessária

Para usar este skill, as seguintes variáveis de ambiente devem estar configuradas:
- `SERVER_PORT`: Porta do servidor (padrão: 8080)
- `JWT_SECRET`: Chave secreta para assinar tokens JWT (se auth habilitada)
- `RATE_LIMIT_REQUESTS`: Número de requisições permitidas por minuto (padrão: 100)
- `ENABLE_AUTH`: Habilitar autenticação JWT (true/false)
- `CORS_ALLOWED_ORIGINS`: Origens permitidas para CORS

## Segurança Adicional

### Headers de Segurança
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`

### CORS
Configurar CORS apropriadamente para permitir apenas origens confiáveis:
```go
cors.AllowedOrigins = []string{"https://app.example.com"}
```

### Rate Limiting
Implementar rate limiting robusto:
- Por IP address
- Por user_id (se autenticado)
- Limites diferentes para endpoints diferentes
