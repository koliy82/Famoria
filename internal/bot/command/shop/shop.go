package shop

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
)

func Register(bh *th.BotHandler, log *zap.Logger) {
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Данная команда пока не реализована :(",
		))
	}, th.Or(th.CommandEqual("shop"), th.TextEqual("💳 Магазин")))
}
