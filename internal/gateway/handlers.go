package gateway

import (
	"encoding/json"
	"net/http"
)

// Health check handlers
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"version": "0.1.0",
	})
}

func (g *Gateway) handleReady(w http.ResponseWriter, r *http.Request) {
	// TODO: Check LLM connectivity
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// Chat handlers
func (g *Gateway) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "message is required")
		return
	}

	// Get user ID from context (set by auth middleware) or use default
	userID := "anonymous"
	if id, ok := r.Context().Value("user_id").(string); ok {
		userID = id
	}

	// Process message with agent
	response, err := g.agent.ProcessMessage(r.Context(), userID, "api", req.Message)
	if err != nil {
		g.logger.Error("failed to process chat message", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to process message")
		return
	}

	respondJSON(w, http.StatusOK, ChatResponse{
		ID:      req.SessionID,
		Message: response,
	})
}

func (g *Gateway) handleChatStream(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement SSE streaming
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// Placeholder streaming response
	w.Write([]byte("data: {\"content\": \"Streaming placeholder\"}\n\n"))
	flusher.Flush()
	w.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}

// Session handlers
func (g *Gateway) handleListSessions(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session listing
	respondJSON(w, http.StatusOK, []Session{})
}

func (g *Gateway) handleGetSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session retrieval
	respondJSON(w, http.StatusOK, Session{})
}

func (g *Gateway) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement session deletion
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Tools handlers
func (g *Gateway) handleListTools(w http.ResponseWriter, r *http.Request) {
	var tools []Tool

	// Add DevOps tools if available
	if devopsTool := g.agent.GetDevOpsTool(); devopsTool != nil {
		for _, def := range devopsTool.GetToolDefinitions() {
			tools = append(tools, Tool{
				Name:        def.Function.Name,
				Description: def.Function.Description,
				Enabled:     true,
			})
		}
	}

	respondJSON(w, http.StatusOK, tools)
}

func (g *Gateway) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement tool execution
	respondJSON(w, http.StatusOK, map[string]string{"status": "executed"})
}

// Config handler
func (g *Gateway) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// Return safe config (no secrets)
	safeConfig := map[string]interface{}{
		"llm": map[string]interface{}{
			"provider": g.cfg.LLM.Provider,
			"model":    g.cfg.LLM.Model,
		},
		"tools": map[string]interface{}{
			"file_read":    g.cfg.Tools.FileRead.Enabled,
			"command_exec": g.cfg.Tools.CommandExecute.Enabled,
			"web_search":   g.cfg.Tools.WebSearch.Enabled,
		},
		"azure_devops": map[string]interface{}{
			"enabled":      g.cfg.AzureDevOps.Enabled,
			"organization": g.cfg.AzureDevOps.Organization,
			"project":      g.cfg.AzureDevOps.Project,
		},
		"telegram": map[string]interface{}{
			"enabled": g.cfg.Telegram.Enabled,
		},
	}
	respondJSON(w, http.StatusOK, safeConfig)
}

// WebSocket handler
func (g *Gateway) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket handling
	http.Error(w, "WebSocket not implemented yet", http.StatusNotImplemented)
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Types
type ChatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

type ChatResponse struct {
	ID         string   `json:"id"`
	Message    string   `json:"message"`
	ToolCalls  []string `json:"tool_calls,omitempty"`
	TokensUsed int      `json:"tokens_used,omitempty"`
}

type Session struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Messages  int    `json:"messages"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}
