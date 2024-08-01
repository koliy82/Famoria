package bot

import (
	"github.com/mymmrac/telego"
	"go.uber.org/zap"
	"go_tg_bot/internal/config"
)

func New(log *zap.Logger, cfg config.Config) *telego.Bot {
	bot, err := telego.NewBot(cfg.TelegramToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Sugar().Error(err)
		panic(err)
	}
	me, err := bot.GetMe()
	if err != nil {
		log.Sugar().Error(err)
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
	return bot
}
