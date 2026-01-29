package channels

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/abelclopes/nomad-iabot/internal/config"
)

// TelegramChannel handles Telegram bot integration
type TelegramChannel struct {
	cfg     *config.TelegramConfig
	bot     *tele.Bot
	logger  *slog.Logger
	handler MessageHandler
}

// MessageHandler processes incoming messages
type MessageHandler func(ctx context.Context, msg IncomingMessage) (string, error)

// IncomingMessage represents an incoming message from any channel
type IncomingMessage struct {
	Channel   string // "telegram", "webchat", etc.
	UserID    string
	Username  string
	Text      string
	ChatID    string
	IsGroup   bool
	ReplyToID string
	Metadata  map[string]string
}

// NewTelegramChannel creates a new Telegram channel
func NewTelegramChannel(cfg *config.TelegramConfig, logger *slog.Logger, handler MessageHandler) (*TelegramChannel, error) {
	pref := tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	tc := &TelegramChannel{
		cfg:     cfg,
		bot:     bot,
		logger:  logger,
		handler: handler,
	}

	tc.setupHandlers()

	return tc, nil
}

func (tc *TelegramChannel) setupHandlers() {
	// Handle text messages
	tc.bot.Handle(tele.OnText, func(c tele.Context) error {
		return tc.handleMessage(c)
	})

	// Handle /start command
	tc.bot.Handle("/start", func(c tele.Context) error {
		return c.Send("👋 Olá! Eu sou o Nomad Agent. Como posso ajudar?")
	})

	// Handle /help command
	tc.bot.Handle("/help", func(c tele.Context) error {
		help := `🤖 *Nomad Agent*

Comandos disponíveis:
/start - Iniciar conversa
/help - Mostrar esta ajuda
/status - Ver status do sistema
/workitems - Listar work items (Azure DevOps)

Envie qualquer mensagem para conversar com o agente.`
		return c.Send(help, tele.ModeMarkdown)
	})

	// Handle /status command
	tc.bot.Handle("/status", func(c tele.Context) error {
		return c.Send("✅ Sistema operacional")
	})

	// Handle /workitems command (Azure DevOps integration)
	tc.bot.Handle("/workitems", func(c tele.Context) error {
		// This will be handled by the agent with the DevOps tool
		return tc.handleMessage(c)
	})
}

func (tc *TelegramChannel) handleMessage(c tele.Context) error {
	// Check if user is allowed
	if !tc.isUserAllowed(c.Sender().ID) {
		tc.logger.Warn("unauthorized user attempted access",
			"user_id", c.Sender().ID,
			"username", c.Sender().Username,
		)
		return c.Send("❌ Você não tem permissão para usar este bot.")
	}

	// Build incoming message
	msg := IncomingMessage{
		Channel:  "telegram",
		UserID:   strconv.FormatInt(c.Sender().ID, 10),
		Username: c.Sender().Username,
		Text:     c.Text(),
		ChatID:   strconv.FormatInt(c.Chat().ID, 10),
		IsGroup:  c.Chat().Type == tele.ChatGroup || c.Chat().Type == tele.ChatSuperGroup,
		Metadata: map[string]string{
			"first_name": c.Sender().FirstName,
			"last_name":  c.Sender().LastName,
		},
	}

	if c.Message().ReplyTo != nil {
		msg.ReplyToID = strconv.Itoa(c.Message().ReplyTo.ID)
	}

	tc.logger.Info("received telegram message",
		"user_id", msg.UserID,
		"username", msg.Username,
		"is_group", msg.IsGroup,
	)

	// Show typing indicator
	_ = c.Notify(tele.Typing)

	// Process message
	ctx := context.Background()
	response, err := tc.handler(ctx, msg)
	if err != nil {
		tc.logger.Error("failed to process message", "error", err)
		return c.Send("❌ Desculpe, ocorreu um erro ao processar sua mensagem.")
	}

	// Send response (split if too long)
	return tc.sendLongMessage(c, response)
}

func (tc *TelegramChannel) isUserAllowed(userID int64) bool {
	// If no allowlist configured, allow all
	if len(tc.cfg.AllowFrom) == 0 {
		return true
	}

	for _, allowed := range tc.cfg.AllowFrom {
		if allowed == userID {
			return true
		}
	}
	return false
}

func (tc *TelegramChannel) sendLongMessage(c tele.Context, text string) error {
	const maxLength = 4000

	if len(text) <= maxLength {
		return c.Send(text)
	}

	// Split into chunks
	chunks := splitText(text, maxLength)
	for _, chunk := range chunks {
		if err := c.Send(chunk); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the Telegram bot
func (tc *TelegramChannel) Start(ctx context.Context) error {
	tc.logger.Info("starting Telegram bot")
	
	go func() {
		tc.bot.Start()
	}()

	<-ctx.Done()
	tc.bot.Stop()
	return nil
}

// Stop stops the Telegram bot
func (tc *TelegramChannel) Stop() {
	tc.bot.Stop()
}

// SendMessage sends a message to a specific chat
func (tc *TelegramChannel) SendMessage(chatID string, text string) error {
	id, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return err
	}

	chat := &tele.Chat{ID: id}
	
	if len(text) <= 4000 {
		_, err = tc.bot.Send(chat, text)
		return err
	}

	// Split long messages
	chunks := splitText(text, 4000)
	for _, chunk := range chunks {
		if _, err := tc.bot.Send(chat, chunk); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to split text into chunks
func splitText(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	lines := strings.Split(text, "\n")
	current := ""

	for _, line := range lines {
		if len(current)+len(line)+1 > maxLen {
			if current != "" {
				chunks = append(chunks, current)
			}
			// If single line is too long, split by words
			if len(line) > maxLen {
				words := strings.Fields(line)
				current = ""
				for _, word := range words {
					if len(current)+len(word)+1 > maxLen {
						chunks = append(chunks, current)
						current = word
					} else {
						if current != "" {
							current += " "
						}
						current += word
					}
				}
			} else {
				current = line
			}
		} else {
			if current != "" {
				current += "\n"
			}
			current += line
		}
	}

	if current != "" {
		chunks = append(chunks, current)
	}

	return chunks
}
