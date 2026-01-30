package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/abelclopes/nomad-iabot/internal/config"
	"github.com/abelclopes/nomad-iabot/internal/devops"
	"github.com/abelclopes/nomad-iabot/internal/llm"
	"github.com/abelclopes/nomad-iabot/internal/trello"
)

// Agent is the core AI agent that processes messages and executes tools
type Agent struct {
	config       *config.Config
	logger       *slog.Logger
	llmClient    *llm.Client
	devopsClient *devops.Client
	devopsTool   *devops.Tool
	trelloClient *trello.Client
	trelloTool   *trello.Tool
}

// New creates a new Agent instance
func New(cfg *config.Config, logger *slog.Logger) (*Agent, error) {
	// Create LLM client
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.TimeoutSec)

	agent := &Agent{
		config:    cfg,
		logger:    logger,
		llmClient: llmClient,
	}

	// Initialize Azure DevOps client if configured
	if cfg.AzureDevOps.PAT != "" && cfg.AzureDevOps.Organization != "" {
		devopsClient := devops.NewClient(
			cfg.AzureDevOps.Organization,
			cfg.AzureDevOps.Project,
			cfg.AzureDevOps.PAT,
			cfg.AzureDevOps.APIVersion,
		)
		agent.devopsClient = devopsClient
		agent.devopsTool = devops.NewTool(devopsClient)
		logger.Info("Azure DevOps integration enabled",
			"organization", cfg.AzureDevOps.Organization,
			"project", cfg.AzureDevOps.Project,
		)
	}

	// Initialize Trello client if configured
	if cfg.Trello.APIKey != "" && cfg.Trello.Token != "" {
		trelloClient := trello.NewClient(cfg.Trello.APIKey, cfg.Trello.Token)
		agent.trelloClient = trelloClient
		agent.trelloTool = trello.NewTool(trelloClient)
		logger.Info("Trello integration enabled")
	}

	return agent, nil
}

// ProcessMessage processes an incoming message and returns a response
func (a *Agent) ProcessMessage(ctx context.Context, userID, channel, message string) (string, error) {
	a.logger.Info("processing message",
		"user_id", userID,
		"channel", channel,
		"message_length", len(message),
	)

	// Build system prompt
	systemPrompt := a.buildSystemPrompt()

	// Build messages
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: message},
	}

	// Get available tools
	tools := a.getAvailableTools()

	// Build chat options
	var opts []llm.ChatOption
	if len(tools) > 0 {
		opts = append(opts, llm.WithTools(tools))
	}

	// Get initial response
	resp, err := a.llmClient.Chat(ctx, messages, opts...)
	if err != nil {
		a.logger.Error("LLM request failed", "error", err)
		return "", fmt.Errorf("failed to process message: %w", err)
	}

	// Check if we have choices
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	choice := resp.Choices[0]

	// Process tool calls if any
	maxIterations := 10 // Safety limit
	for i := 0; i < maxIterations && len(choice.ToolCalls) > 0; i++ {
		a.logger.Info("processing tool calls", "count", len(choice.ToolCalls), "iteration", i+1)

		// Add assistant message with tool calls
		messages = append(messages, llm.Message{
			Role:    "assistant",
			Content: choice.Message.Content,
		})

		// Execute each tool call
		for _, tc := range choice.ToolCalls {
			result, err := a.executeTool(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %s", err.Error())
			}

			// Add tool result
			messages = append(messages, llm.Message{
				Role:    "tool",
				Content: result,
			})
		}

		// Get next response
		resp, err = a.llmClient.Chat(ctx, messages, opts...)
		if err != nil {
			a.logger.Error("LLM request failed during tool processing", "error", err)
			return "", fmt.Errorf("failed to process tool results: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no response from LLM")
		}
		choice = resp.Choices[0]
	}

	return choice.Message.Content, nil
}

// buildSystemPrompt creates the system prompt for the agent
func (a *Agent) buildSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString("Você é o Nomad Agent, um assistente AI inteligente e prestativo.\n\n")
	sb.WriteString("## Suas Capacidades\n")
	sb.WriteString("- Responder perguntas de forma clara e objetiva\n")
	sb.WriteString("- Ajudar com tarefas de programação e desenvolvimento\n")

	if a.devopsClient != nil {
		sb.WriteString("- Gerenciar projetos no Azure DevOps (work items, pipelines, repositórios)\n")
		sb.WriteString("\n## Azure DevOps\n")
		sb.WriteString(fmt.Sprintf("Organização: %s\n", a.config.AzureDevOps.Organization))
		sb.WriteString(fmt.Sprintf("Projeto padrão: %s\n", a.config.AzureDevOps.Project))
	}

	if a.trelloClient != nil {
		sb.WriteString("- Gerenciar boards, listas e cards no Trello\n")
		sb.WriteString("\n## Trello\n")
		sb.WriteString("Você pode criar, atualizar e consultar boards, listas e cards do Trello.\n")
	}

	sb.WriteString("\n## Diretrizes\n")
	sb.WriteString("- Seja conciso e direto nas respostas\n")
	sb.WriteString("- Use formatação Markdown quando apropriado\n")
	sb.WriteString("- Quando usar ferramentas, explique o que está fazendo\n")
	sb.WriteString("- Responda no idioma do usuário\n")

	return sb.String()
}

// getAvailableTools returns the list of available tools
func (a *Agent) getAvailableTools() []llm.Tool {
	var tools []llm.Tool

	// Add DevOps tools if available
	if a.devopsTool != nil {
		tools = append(tools, a.devopsTool.GetToolDefinitions()...)
	}

	// Add Trello tools if available
	if a.trelloTool != nil {
		tools = append(tools, a.trelloTool.GetToolDefinitions()...)
	}

	return tools
}

// executeTool executes a tool and returns the result
func (a *Agent) executeTool(ctx context.Context, name string, arguments string) (string, error) {
	a.logger.Info("executing tool", "name", name)

	// Parse arguments
	var args map[string]interface{}
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	// Execute DevOps tools
	if a.devopsTool != nil {
		result, handled, err := a.devopsTool.Execute(ctx, name, args)
		if handled {
			if err != nil {
				return "", err
			}
			return result, nil
		}
	}

	// Execute Trello tools
	if a.trelloTool != nil {
		result, handled, err := a.trelloTool.Execute(ctx, name, args)
		if handled {
			if err != nil {
				return "", err
			}
			return result, nil
		}
	}

	return "", fmt.Errorf("unknown tool: %s", name)
}

// GetDevOpsClient returns the Azure DevOps client
func (a *Agent) GetDevOpsClient() *devops.Client {
	return a.devopsClient
}

// GetDevOpsTool returns the Azure DevOps tool
func (a *Agent) GetDevOpsTool() *devops.Tool {
	return a.devopsTool
}

// GetTrelloClient returns the Trello client
func (a *Agent) GetTrelloClient() *trello.Client {
	return a.trelloClient
}

// GetTrelloTool returns the Trello tool
func (a *Agent) GetTrelloTool() *trello.Tool {
	return a.trelloTool
}

// GetLLMClient returns the LLM client
func (a *Agent) GetLLMClient() *llm.Client {
	return a.llmClient
}
