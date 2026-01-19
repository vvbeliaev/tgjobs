package in

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"svpb-tmpl/config"
	"svpb-tmpl/pkg/collector/core"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// TelegramAdapter handles Telegram message collection.
type TGAdapter struct {
	cfg        config.TelegramConfig
	client     *telegram.Client
	dispatcher tg.UpdateDispatcher
	service    core.CollectorService
	logger     *zap.Logger
}

// NewTelegram creates a new Telegram adapter.
func NewTG(cfg config.TelegramConfig, service core.CollectorService, logger *zap.Logger) *TGAdapter {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	dispatcher := tg.NewUpdateDispatcher()

	client := telegram.NewClient(cfg.APIID, cfg.APIHash, telegram.Options{
		Logger:         logger,
		SessionStorage: &telegram.FileSessionStorage{Path: cfg.SessionPath},
		UpdateHandler:  dispatcher,
		Device: telegram.DeviceConfig{
			DeviceModel:    "Desktop",
			SystemVersion:  "Windows 10",
			AppVersion:     "1.0.0",
			SystemLangCode: "en",
			LangCode:       "en",
		},
	})

	adapter := &TGAdapter{
		cfg:        cfg,
		client:     client,
		dispatcher: dispatcher,
		service:    service,
		logger:     logger,
	}

	// Set up message handlers
	adapter.setupHandlers()

	return adapter
}

// setupHandlers registers Telegram update handlers.
func (t *TGAdapter) setupHandlers() {
	// Channels and Supergroups
	t.dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}

		peer, ok := msg.PeerID.(*tg.PeerChannel)
		if !ok {
			return nil
		}

		return t.service.Handle(ctx, core.Message{
			Text:      msg.Message,
			ChannelID: peer.ChannelID,
			MessageID: msg.ID,
			RawData:   msg,
		})
	})

	// Legacy Groups and Private Chats
	t.dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok || msg.Out {
			return nil
		}

		var peerID int64
		switch p := msg.PeerID.(type) {
		case *tg.PeerChat:
			peerID = p.ChatID
		case *tg.PeerUser:
			peerID = p.UserID
			if user, ok := e.Users[p.UserID]; ok {
				if !user.Bot {
					t.logger.Debug("Ignoring private message from human user", zap.Int64("user_id", p.UserID))
					return nil
				}
			}
		default:
			return nil
		}

		return t.service.Handle(ctx, core.Message{
			Text:      msg.Message,
			ChannelID: peerID,
			MessageID: msg.ID,
			RawData:   msg,
		})
	})
}

// RegisterCommand adds tg-login command to PocketBase.
func (t *TGAdapter) RegisterCommand(app *pocketbase.PocketBase) {
	app.RootCmd.AddCommand(&cobra.Command{
		Use:   "tg-login",
		Short: "Login to Telegram and save session",
		Long:  "Performs interactive Telegram authentication. Use this to generate session.json before deploying to server.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			if err := t.Login(ctx); err != nil {
				t.logger.Fatal("Login failed", zap.Error(err))
			}
		},
	})
}

// Login performs interactive Telegram authentication.
func (t *TGAdapter) Login(ctx context.Context) error {
	return t.client.Run(ctx, func(ctx context.Context) error {
		flow := auth.NewFlow(terminalAuth{phone: t.cfg.Phone}, auth.SendCodeOptions{})

		if err := t.client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}

		self, err := t.client.Self(ctx)
		if err != nil {
			return fmt.Errorf("failed to get self: %w", err)
		}

		t.logger.Info("Successfully logged in",
			zap.String("username", self.Username),
			zap.Int64("user_id", self.ID),
			zap.String("first_name", self.FirstName),
		)

		fmt.Printf("\nLogged in as: %s (@%s)\n", self.FirstName, self.Username)
		fmt.Printf("Session saved to: %s\n", t.cfg.SessionPath)

		return nil
	})
}

// Start begins listening for Telegram messages.
func (t *TGAdapter) Start(ctx context.Context) error {
	return t.client.Run(ctx, func(ctx context.Context) error {
		status, err := t.client.Auth().Status(ctx)
		if err != nil {
			return fmt.Errorf("failed to get auth status: %w", err)
		}

		if !status.Authorized {
			return fmt.Errorf("not authorized - run 'tg-login' first")
		}

		self, err := t.client.Self(ctx)
		if err != nil {
			return fmt.Errorf("failed to get self: %w", err)
		}

		t.logger.Info("Telegram client started",
			zap.String("username", self.Username),
			zap.Int64("user_id", self.ID),
		)

		<-ctx.Done()
		return ctx.Err()
	})
}

// IsConfigured returns true if Telegram credentials are set.
func (t *TGAdapter) IsConfigured() bool {
	return t.cfg.APIID != 0 && t.cfg.APIHash != ""
}

// SessionExists returns true if session file exists.
func (t *TGAdapter) SessionExists() bool {
	_, err := os.Stat(t.cfg.SessionPath)
	return !os.IsNotExist(err)
}

// --- Terminal Auth Helper ---

type terminalAuth struct {
	phone string
}

func (a terminalAuth) Phone(_ context.Context) (string, error) {
	if a.phone != "" {
		return a.phone, nil
	}
	fmt.Print("Enter phone number: ")
	reader := bufio.NewReader(os.Stdin)
	phone, _ := reader.ReadString('\n')
	return strings.TrimSpace(phone), nil
}

func (a terminalAuth) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")
	reader := bufio.NewReader(os.Stdin)
	password, _ := reader.ReadString('\n')
	return strings.TrimSpace(password), nil
}

func (a terminalAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter auth code: ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	return strings.TrimSpace(code), nil
}

func (a terminalAuth) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return nil
}

func (a terminalAuth) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, fmt.Errorf("sign up not supported")
}
