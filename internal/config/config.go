package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for Nomad Agent
type Config struct {
	Gateway     GatewayConfig
	LLM         LLMConfig
	Security    SecurityConfig
	AzureDevOps AzureDevOpsConfig
	Trello      TrelloConfig
	Telegram    TelegramConfig
	Tools       ToolsConfig
}

// GatewayConfig holds gateway/server configuration
type GatewayConfig struct {
	HTTPPort    int
	WSPort      int
	Bind        string // IP address to bind to (e.g., "0.0.0.0" for all interfaces, "127.0.0.1" for localhost)
	CORSOrigins []string
}

// LLMConfig holds LLM provider configuration
type LLMConfig struct {
	Provider    string // "ollama", "lmstudio", "localai", "openrouter", "openai"
	BaseURL     string
	Model       string
	APIKey      string // API Key for OpenRouter, OpenAI, etc.
	MaxTokens   int
	Temperature float64
	TimeoutSec  int
}

// SecurityConfig holds security settings
type SecurityConfig struct {
	JWTSecret      string
	RateLimitRPS   int    // requests per second
	RateLimitBurst int    // burst size
	AuthMode       string // "jwt", "api-key", "none"
}

// AzureDevOpsConfig holds Azure DevOps integration settings
type AzureDevOpsConfig struct {
	Enabled      bool
	Organization string
	Project      string
	PAT          string // Personal Access Token
	APIVersion   string
}

// TrelloConfig holds Trello integration settings
type TrelloConfig struct {
	Enabled bool
	APIKey  string
	Token   string
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	Enabled   bool
	BotToken  string
	AllowFrom []int64 // allowed user IDs (empty = all)
}

// ToolsConfig holds tool permissions
type ToolsConfig struct {
	FileRead       FileReadConfig
	CommandExecute CommandExecuteConfig
	WebSearch      WebSearchConfig
}

// FileReadConfig holds file reading permissions
type FileReadConfig struct {
	Enabled          bool
	AllowedPaths     []string
	MaxFileSizeBytes int64
}

// CommandExecuteConfig holds command execution permissions
type CommandExecuteConfig struct {
	Enabled         bool
	AllowedCommands []string
	TimeoutSec      int
}

// WebSearchConfig holds web search settings
type WebSearchConfig struct {
	Enabled bool
	Engine  string // "duckduckgo", "searxng"
	BaseURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Gateway: GatewayConfig{
			HTTPPort:    getEnvInt("GATEWAY_PORT", 8080),
			WSPort:      getEnvInt("GATEWAY_WS_PORT", 8081),
			Bind:        getEnv("GATEWAY_HOST", "0.0.0.0"),
			CORSOrigins: getEnvSlice("GATEWAY_CORS_ORIGINS", []string{"http://localhost:*"}),
		},
		LLM: LLMConfig{
			Provider:    getEnv("LLM_PROVIDER", "ollama"),
			BaseURL:     getEnv("LLM_BASE_URL", "http://localhost:11434"),
			Model:       getEnv("LLM_MODEL", "llama3.2"),
			APIKey:      getEnv("LLM_API_KEY", ""),
			MaxTokens:   getEnvInt("LLM_MAX_TOKENS", 4096),
			Temperature: getEnvFloat("LLM_TEMPERATURE", 0.7),
			TimeoutSec:  getEnvInt("LLM_TIMEOUT", 120),
		},
		Security: SecurityConfig{
			JWTSecret:      getEnv("JWT_SECRET", ""),
			RateLimitRPS:   getEnvInt("RATE_LIMIT_RPS", 10),
			RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 20),
			AuthMode:       getEnv("AUTH_MODE", "jwt"),
		},
		AzureDevOps: AzureDevOpsConfig{
			Enabled:      getEnvBool("AZURE_DEVOPS_ENABLED", false),
			Organization: getEnv("AZURE_DEVOPS_ORGANIZATION", ""),
			Project:      getEnv("AZURE_DEVOPS_PROJECT", ""),
			PAT:          getEnv("AZURE_DEVOPS_PAT", ""),
			APIVersion:   getEnv("AZURE_DEVOPS_API_VERSION", "7.0"),
		},
		Trello: TrelloConfig{
			Enabled: getEnvBool("TRELLO_ENABLED", false),
			APIKey:  getEnv("TRELLO_API_KEY", ""),
			Token:   getEnv("TRELLO_TOKEN", ""),
		},
		Telegram: TelegramConfig{
			Enabled:   getEnvBool("TELEGRAM_ENABLED", false),
			BotToken:  getEnv("TELEGRAM_BOT_TOKEN", ""),
			AllowFrom: getEnvInt64Slice("TELEGRAM_ALLOWED_USERS", nil),
		},
		Tools: ToolsConfig{
			FileRead: FileReadConfig{
				Enabled:          getEnvBool("TOOLS_FILE_READ", true),
				AllowedPaths:     getEnvSlice("TOOLS_FILE_ALLOWED_PATHS", []string{"/workspace"}),
				MaxFileSizeBytes: getEnvInt64("TOOLS_FILE_MAX_SIZE", 10*1024*1024), // 10MB
			},
			CommandExecute: CommandExecuteConfig{
				Enabled:         getEnvBool("TOOLS_COMMAND_EXEC", false),
				AllowedCommands: getEnvSlice("TOOLS_ALLOWED_COMMANDS", []string{"ls", "cat", "grep", "find"}),
				TimeoutSec:      getEnvInt("TOOLS_COMMAND_TIMEOUT", 30),
			},
			WebSearch: WebSearchConfig{
				Enabled: getEnvBool("TOOLS_WEB_SEARCH", false),
				Engine:  getEnv("TOOLS_SEARCH_ENGINE", "duckduckgo"),
				BaseURL: getEnv("TOOLS_SEARCH_URL", ""),
			},
		},
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	// Security: require JWT secret in jwt mode
	if c.Security.AuthMode == "jwt" && c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required when auth mode is 'jwt'")
	}

	// Azure DevOps validation
	if c.AzureDevOps.Enabled {
		if c.AzureDevOps.Organization == "" {
			return fmt.Errorf("AZURE_DEVOPS_ORGANIZATION is required when Azure DevOps is enabled")
		}
		if c.AzureDevOps.Project == "" {
			return fmt.Errorf("AZURE_DEVOPS_PROJECT is required when Azure DevOps is enabled")
		}
		if c.AzureDevOps.PAT == "" {
			return fmt.Errorf("AZURE_DEVOPS_PAT is required when Azure DevOps is enabled")
		}
	}

	// Trello validation
	if c.Trello.Enabled {
		if c.Trello.APIKey == "" {
			return fmt.Errorf("TRELLO_API_KEY is required when Trello is enabled")
		}
		if c.Trello.Token == "" {
			return fmt.Errorf("TRELLO_TOKEN is required when Trello is enabled")
		}
	}

	// Telegram validation
	if c.Telegram.Enabled && c.Telegram.BotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required when Telegram is enabled")
	}

	return nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvInt64Slice(key string, defaultValue []int64) []int64 {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]int64, 0, len(parts))
		for _, p := range parts {
			if i, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
				result = append(result, i)
			}
		}
		return result
	}
	return defaultValue
}
