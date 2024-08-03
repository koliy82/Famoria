package bot

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"go_tg_bot/internal/config"
)

func New(cfg config.Config) *telego.Bot {
	bot, err := telego.NewBot(cfg.TelegramToken, telego.WithDefaultLogger(true, true))
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
