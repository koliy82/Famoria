package family

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go_tg_bot/internal/bot/callback"
	"go_tg_bot/internal/database/mongo/repositories/brak"
	"go_tg_bot/internal/database/mongo/repositories/user"
)

type leaveKid struct {
	cm    *callback.CallbacksManager
	braks brak.Repository
	users user.Repository
}

func (e leaveKid) Handle(bot *telego.Bot, update telego.Update) {
	_, _ = bot.SendMessage(tu.Messagef(
		tu.ID(update.Message.Chat.ID),
		"Hello %s!\n Данная команда пока не реализована..", update.Message.From.FirstName,
	))
}
