package family

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go_tg_bot/internal/bot/callback"
)

type endFamily struct {
	cm *callback.CallbacksManager
}

func (e endFamily) Handle(bot *telego.Bot, update telego.Update) {
	_, _ = bot.SendMessage(tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
	))
}
