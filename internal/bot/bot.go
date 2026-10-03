package bot

import (
	"context"
	"fmt"

	"famoria/internal/config"
	"famoria/internal/pkg/proxy"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

func New(cfg config.Config) *telego.Bot {
	token := cfg.TelegramToken

	opts := []telego.BotOption{
		telego.WithDefaultLogger(false, true),
	}
	if cfg.TelegramTestToken != nil {
		token = *cfg.TelegramTestToken
		opts = append(opts, telego.WithTestServerPath())
	}

	// Routing Bot API traffic through a proxy is how the bot reaches Telegram at
	// all where api.telegram.org is blocked. This also covers long polling and the
	// zap-to-Telegram log sink, because both go through the same caller.
	if cfg.BotProxy {
		ep := proxy.MustParse(deref(cfg.BotProxyURL))
		opts = append(opts, telego.WithHTTPClient(ep.HTTPClient()))
		// The zap logger is not usable here: SetupLogger depends on this bot, so
		// the global logger is still the no-op default. stdout is what docker logs
		// captures, so the proxy state is still visible at startup.
		fmt.Printf("bot: Telegram API traffic routed through proxy %s\n", ep.Masked())
	}

	bot, err := telego.NewBot(token, opts...)
	if err != nil {
		panic(err)
	}
	return bot
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func PrintMe(log *zap.Logger, bot *telego.Bot) {
	me, err := bot.GetMe(context.Background())
	if err != nil {
		panic(err)
	}
	m := Me{
		ID:        me.ID,
		Username:  me.Username,
		FirstName: me.FirstName,
		LastName:  me.LastName,
		IsBot:     me.IsBot,
	}
	m.Print(log)
}
