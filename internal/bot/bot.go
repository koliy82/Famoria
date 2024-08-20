package bot

import (
	"famoria/internal/config"
	"github.com/koliy82/telego"
	"go.uber.org/zap"
)

func New(cfg config.Config) *telego.Bot {
	bot, err := telego.NewBot(cfg.TelegramToken, telego.WithDefaultLogger(false, true))
	if err != nil {
		panic(err)
	}
	return bot
}

func PrintMe(log *zap.Logger, bot *telego.Bot) {
	me, err := bot.GetMe()
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
