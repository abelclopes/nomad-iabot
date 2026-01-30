---
name: "Azure DevOps Integration"
description: "Skill for managing Azure DevOps work items, pipelines, repositories, and boards"
version: "1.0.0"
integration: "azure_devops"
security_level: "high"
---

# Azure DevOps Integration Skills

## Objetivo
Este skill define as operações permitidas para a integração com Azure DevOps, garantindo que o agente opere apenas dentro dos limites seguros e definidos.

## Operações Permitidas

### Work Items

#### 1. Listar Work Items Atribuídos ao Usuário
- **Comando**: `devops_list_my_workitems`
- **Descrição**: Lista work items (tasks, bugs, stories) atribuídos ao usuário atual que não estão fechados
- **Parâmetros**: Nenhum
- **Restrições**: Apenas work items do usuário atual
- **Exemplo**: "Liste meus work items"

#### 2. Obter Detalhes de Work Item
- **Comando**: `devops_get_workitem`
- **Descrição**: Obtém detalhes completos de um work item específico
- **Parâmetros**: 
  - `id` (obrigatório): ID numérico do work item
- **Restrições**: Apenas work items que o usuário tem permissão de visualizar
- **Exemplo**: "Mostre detalhes do work item #123"

#### 3. Criar Work Item
- **Comando**: `devops_create_workitem`
- **Descrição**: Cria um novo work item
- **Parâmetros**:
  - `type` (obrigatório): Task, Bug, User Story, Feature, Epic
  - `title` (obrigatório): Título do work item
  - `description` (opcional): Descrição em HTML
  - `assigned_to` (opcional): Email ou nome do responsável
  - `priority` (opcional): 1 (mais alta) a 4 (mais baixa)
  - `tags` (opcional): Array de tags
  - `parent_id` (opcional): ID do work item pai
- **Restrições**: 
  - Tipos de work item limitados aos definidos
  - Prioridade deve estar entre 1-4
  - Título não pode estar vazio
- **Exemplo**: "Crie uma task com título 'Implementar login'"

#### 4. Atualizar Work Item
- **Comando**: `devops_update_workitem`
- **Descrição**: Atualiza um work item existente
- **Parâmetros**:
  - `id` (obrigatório): ID do work item
  - `title` (opcional): Novo título
  - `state` (opcional): Novo estado (New, Active, Resolved, Closed)
  - `assigned_to` (opcional): Novo responsável
  - `priority` (opcional): Nova prioridade (1-4)
- **Restrições**: 
  - Apenas work items que o usuário tem permissão de editar
  - Estados devem ser válidos para o tipo de work item
- **Exemplo**: "Mude o estado do work item #123 para Closed"

#### 5. Consultar Work Items (WIQL)
- **Comando**: `devops_query_workitems`
- **Descrição**: Consulta work items usando WIQL (Work Item Query Language)
- **Parâmetros**:
  - `query` (obrigatório): Query WIQL válida
- **Restrições**: 
  - Query deve ser sintaxe WIQL válida
  - Não permitir queries que retornem dados sensíveis
  - Limitar resultados a um número razoável
- **Exemplo**: "Liste todos os bugs abertos do projeto"

### Pipelines

#### 6. Listar Pipelines
- **Comando**: `devops_list_pipelines`
- **Descrição**: Lista todos os pipelines no projeto
- **Parâmetros**: Nenhum
- **Restrições**: Apenas pipelines que o usuário tem permissão de visualizar
- **Exemplo**: "Liste os pipelines disponíveis"

#### 7. Executar Pipeline
- **Comando**: `devops_run_pipeline`
- **Descrição**: Dispara a execução de um pipeline
- **Parâmetros**:
  - `pipeline_id` (obrigatório): ID do pipeline
  - `branch` (opcional): Branch git (padrão: refs/heads/main)
  - `variables` (opcional): Variáveis do pipeline como chave-valor
- **Restrições**: 
  - Apenas pipelines que o usuário tem permissão de executar
  - Branch deve existir no repositório
  - Variáveis devem seguir formato key-value
- **Exemplo**: "Execute o pipeline #5 na branch develop"

### Repositórios

#### 8. Listar Repositórios
- **Comando**: `devops_list_repos`
- **Descrição**: Lista todos os repositórios Git no projeto
- **Parâmetros**: Nenhum
- **Restrições**: Apenas repositórios que o usuário tem permissão de visualizar
- **Exemplo**: "Liste os repositórios do projeto"

### Boards

#### 9. Listar Boards
- **Comando**: `devops_list_boards`
- **Descrição**: Lista todos os boards (Kanban) do projeto
- **Parâmetros**:
  - `team` (opcional): Nome do time
- **Restrições**: Apenas boards que o usuário tem permissão de visualizar
- **Exemplo**: "Liste os boards do time DevOps"

## Regras de Segurança

### Prevenção de Prompt Injection
1. **Validação de Entrada**: Todos os parâmetros devem ser validados antes da execução
2. **Sanitização**: Remover caracteres especiais e comandos maliciosos
3. **Whitelist de Comandos**: Apenas comandos listados acima são permitidos
4. **Rate Limiting**: Limitar número de chamadas por minuto
5. **Autenticação**: Sempre verificar permissões do usuário

### Operações NÃO Permitidas
- ❌ Deletar work items
- ❌ Deletar pipelines
- ❌ Deletar repositórios
- ❌ Modificar permissões de usuários
- ❌ Acessar dados de outros projetos sem permissão
- ❌ Executar comandos arbitrários
- ❌ Modificar configurações de segurança
- ❌ Exportar dados sensíveis em massa

## Tratamento de Erros
- Sempre retornar mensagens de erro amigáveis
- Não expor detalhes internos do sistema
- Log de todas as operações para auditoria
- Retry automático para falhas temporárias

## Exemplo de Uso

```
Usuário: "Liste meus work items abertos"
Agente: Executa devops_list_my_workitems
Resultado: Lista formatada de work items

Usuário: "Crie uma task para implementar o login"
Agente: Executa devops_create_workitem com type="Task", title="Implementar login"
Resultado: "Created work item #456: Implementar login"
```

## Configuração Necessária

Para usar este skill, as seguintes variáveis de ambiente devem estar configuradas:
- `AZURE_DEVOPS_PAT`: Personal Access Token com permissões adequadas
- `AZURE_DEVOPS_ORGANIZATION`: Nome da organização
- `AZURE_DEVOPS_PROJECT`: Nome do projeto padrão
- `AZURE_DEVOPS_API_VERSION`: Versão da API (padrão: 7.0)

## Permissões Necessárias no PAT
- Work Items: Read & Write
- Code: Read
- Build: Read & Execute
- Project and Team: Read
