package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/abelclopes/nomad-iabot/internal/agent"
	"github.com/abelclopes/nomad-iabot/internal/channels"
	"github.com/abelclopes/nomad-iabot/internal/config"
)

// Gateway is the main HTTP/WS server for Nomad Agent
type Gateway struct {
	cfg        *config.Config
	logger     *slog.Logger
	httpServer *http.Server
	router     *chi.Mux
	agent      *agent.Agent
	webchat    *channels.WebChatChannel
}

// New creates a new Gateway instance
func New(cfg *config.Config, logger *slog.Logger, ag *agent.Agent) (*Gateway, error) {
	g := &Gateway{
		cfg:    cfg,
		logger: logger,
		router: chi.NewRouter(),
		agent:  ag,
	}

	g.setupMiddleware()
	g.setupRoutes()

	return g, nil
}

// RegisterWebChat registers the WebChat channel
func (g *Gateway) RegisterWebChat(wc *channels.WebChatChannel) {
	g.webchat = wc
	wc.RegisterRoutes(g.router)
}

func (g *Gateway) setupMiddleware() {
	// Request ID
	g.router.Use(middleware.RequestID)

	// Real IP
	g.router.Use(middleware.RealIP)

	// Structured logging
	g.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			g.logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	})

	// Recovery
	g.router.Use(middleware.Recoverer)

	// CORS
	g.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   g.cfg.Gateway.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	g.router.Use(httprate.LimitByIP(
		g.cfg.Security.RateLimitRPS,
		time.Second,
	))

	// Timeout
	g.router.Use(middleware.Timeout(60 * time.Second))
}

func (g *Gateway) setupRoutes() {
	// Health check (no auth required)
	g.router.Get("/health", g.handleHealth)
	g.router.Get("/ready", g.handleReady)

	// API routes (with auth)
	g.router.Route("/api/v1", func(r chi.Router) {
		// Auth middleware for API routes
		if g.cfg.Security.AuthMode == "token" {
			r.Use(g.authMiddleware)
		}

		// Chat/Agent endpoints
		r.Post("/chat", g.handleChat)
		r.Post("/chat/stream", g.handleChatStream)

		// Sessions
		r.Get("/sessions", g.handleListSessions)
		r.Get("/sessions/{id}", g.handleGetSession)
		r.Delete("/sessions/{id}", g.handleDeleteSession)

		// Tools
		r.Get("/tools", g.handleListTools)
		r.Post("/tools/{name}/execute", g.handleExecuteTool)

		// Azure DevOps (if enabled)
		r.Route("/devops", func(r chi.Router) {
			r.Get("/workitems", g.handleListWorkItems)
			r.Post("/workitems", g.handleCreateWorkItem)
			r.Get("/workitems/{id}", g.handleGetWorkItem)
			r.Patch("/workitems/{id}", g.handleUpdateWorkItem)
			r.Get("/pipelines", g.handleListPipelines)
			r.Post("/pipelines/{id}/run", g.handleRunPipeline)
			r.Get("/repos", g.handleListRepos)
			r.Get("/boards", g.handleListBoards)
		})

		// Config
		r.Get("/config", g.handleGetConfig)
	})

	// WebChat static files
	g.router.Handle("/webchat/*", http.StripPrefix("/webchat/", http.FileServer(http.Dir("./web/dist"))))

	// WebSocket for real-time chat
	g.router.Get("/ws", g.handleWebSocket)
}

// Start starts the HTTP server
func (g *Gateway) Start(ctx context.Context) error {
	bindAddr := "127.0.0.1"
	if g.cfg.Gateway.Bind == "all" {
		bindAddr = "0.0.0.0"
	}

	addr := fmt.Sprintf("%s:%d", bindAddr, g.cfg.Gateway.HTTPPort)
	g.httpServer = &http.Server{
		Addr:         addr,
		Handler:      g.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	g.logger.Info("HTTP server starting", "addr", addr)
	return g.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (g *Gateway) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return g.httpServer.Shutdown(shutdownCtx)
}
