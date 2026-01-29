package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/abelclopes/nomad-iabot/internal/agent"
	"github.com/abelclopes/nomad-iabot/internal/channels"
	"github.com/abelclopes/nomad-iabot/internal/config"
	"github.com/abelclopes/nomad-iabot/internal/gateway"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("🚀 Starting Nomad Agent", "version", "0.1.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create the AI agent
	aiAgent, err := agent.New(cfg, logger)
	if err != nil {
		slog.Error("Failed to create agent", "error", err)
		os.Exit(1)
	}

	// Message handler using the agent
	messageHandler := func(ctx context.Context, msg channels.IncomingMessage) (string, error) {
		return aiAgent.ProcessMessage(ctx, msg.UserID, msg.Channel, msg.Text)
	}

	// Create and start gateway
	gw, err := gateway.New(cfg, logger, aiAgent)
	if err != nil {
		slog.Error("Failed to create gateway", "error", err)
		os.Exit(1)
	}

	// Setup WebChat channel
	webchat := channels.NewWebChatChannel(logger, messageHandler)
	gw.RegisterWebChat(webchat)

	// Start webchat session cleanup routine
	go webchat.StartCleanupRoutine(ctx, 5*time.Minute, 1*time.Hour)

	// Start Telegram bot if configured
	if cfg.Telegram.BotToken != "" {
		telegramBot, err := channels.NewTelegramChannel(&cfg.Telegram, logger, messageHandler)
		if err != nil {
			slog.Error("Failed to create Telegram bot", "error", err)
		} else {
			go telegramBot.Start(ctx)
			slog.Info("Telegram bot started")
		}
	}

	// Start gateway in goroutine
	go func() {
		if err := gw.Start(ctx); err != nil {
			slog.Error("Gateway error", "error", err)
			cancel()
		}
	}()

	slog.Info("Nomad Agent is running",
		"http_port", cfg.Gateway.HTTPPort,
	)

	// Wait for shutdown signal
	<-sigChan
	slog.Info("Shutting down gracefully...")
	cancel()

	if err := gw.Shutdown(ctx); err != nil {
		slog.Error("Error during shutdown", "error", err)
	}

	slog.Info("Nomad Agent stopped")
}
