package channels

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WebChatChannel handles the web-based chat interface
type WebChatChannel struct {
	logger   *slog.Logger
	handler  MessageHandler
	sessions sync.Map // map[sessionID]*WebChatSession
}

// WebChatSession represents a webchat session
type WebChatSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []WebChatMessage `json:"messages"`
	mu        sync.Mutex
}

// WebChatMessage represents a message in the web chat
type WebChatMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// NewWebChatChannel creates a new WebChat channel
func NewWebChatChannel(logger *slog.Logger, handler MessageHandler) *WebChatChannel {
	return &WebChatChannel{
		logger:  logger,
		handler: handler,
	}
}

// RegisterRoutes registers the WebChat routes
func (wc *WebChatChannel) RegisterRoutes(r chi.Router) {
	r.Route("/webchat/api", func(r chi.Router) {
		r.Post("/sessions", wc.handleCreateSession)
		r.Get("/sessions/{id}", wc.handleGetSession)
		r.Delete("/sessions/{id}", wc.handleDeleteSession)
		r.Post("/sessions/{id}/messages", wc.handleSendMessage)
		r.Get("/sessions/{id}/messages", wc.handleGetMessages)
	})
}

func (wc *WebChatChannel) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.UserID = "anonymous"
	}

	session := &WebChatSession{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		CreatedAt: time.Now(),
		Messages:  []WebChatMessage{},
	}

	wc.sessions.Store(session.ID, session)

	wc.logger.Info("created webchat session", "session_id", session.ID, "user_id", session.UserID)

	respondJSON(w, http.StatusCreated, session)
}

func (wc *WebChatChannel) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	session, ok := wc.sessions.Load(sessionID)
	if !ok {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}

	respondJSON(w, http.StatusOK, session)
}

func (wc *WebChatChannel) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	wc.sessions.Delete(sessionID)

	wc.logger.Info("deleted webchat session", "session_id", sessionID)

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (wc *WebChatChannel) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	sessionVal, ok := wc.sessions.Load(sessionID)
	if !ok {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}

	session := sessionVal.(*WebChatSession)

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "content is required")
		return
	}

	// Add user message
	userMsg := WebChatMessage{
		ID:        uuid.New().String(),
		Role:      "user",
		Content:   req.Content,
		Timestamp: time.Now(),
	}

	session.mu.Lock()
	session.Messages = append(session.Messages, userMsg)
	session.mu.Unlock()

	// Process with handler
	incomingMsg := IncomingMessage{
		Channel:  "webchat",
		UserID:   session.UserID,
		Username: session.UserID,
		Text:     req.Content,
		ChatID:   session.ID,
		IsGroup:  false,
		Metadata: map[string]string{
			"session_id": session.ID,
		},
	}

	ctx := r.Context()
	response, err := wc.handler(ctx, incomingMsg)
	if err != nil {
		wc.logger.Error("failed to process message", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to process message")
		return
	}

	// Add assistant message
	assistantMsg := WebChatMessage{
		ID:        uuid.New().String(),
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
	}

	session.mu.Lock()
	session.Messages = append(session.Messages, assistantMsg)
	session.mu.Unlock()

	wc.logger.Info("processed webchat message",
		"session_id", session.ID,
		"user_id", session.UserID,
	)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_message":      userMsg,
		"assistant_message": assistantMsg,
	})
}

func (wc *WebChatChannel) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")

	sessionVal, ok := wc.sessions.Load(sessionID)
	if !ok {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}

	session := sessionVal.(*WebChatSession)

	session.mu.Lock()
	messages := make([]WebChatMessage, len(session.Messages))
	copy(messages, session.Messages)
	session.mu.Unlock()

	respondJSON(w, http.StatusOK, messages)
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

// CleanupOldSessions removes sessions older than the specified duration
func (wc *WebChatChannel) CleanupOldSessions(maxAge time.Duration) {
	now := time.Now()
	wc.sessions.Range(func(key, value interface{}) bool {
		session := value.(*WebChatSession)
		if now.Sub(session.CreatedAt) > maxAge {
			wc.sessions.Delete(key)
			wc.logger.Info("cleaned up old session", "session_id", session.ID)
		}
		return true
	})
}

// StartCleanupRoutine starts a goroutine that periodically cleans up old sessions
func (wc *WebChatChannel) StartCleanupRoutine(ctx context.Context, interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			wc.CleanupOldSessions(maxAge)
		}
	}
}
