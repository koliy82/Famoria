package admin

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"go_tg_bot/internal/bot/predicate"
	"go_tg_bot/internal/repositories/user"
	"strings"
)

func Register(bh *th.BotHandler, log *zap.Logger, ch user.Repository) {
	if ch == nil {
		panic("user repository is nil")
	}
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		chatID := tu.ID(update.Message.Chat.ID)
		args := strings.Split(update.Message.Text, " ")
		err := bot.DeleteMessage(
			&telego.DeleteMessageParams{
				ChatID:    chatID,
				MessageID: update.Message.MessageID,
			},
		)
		if err != nil {
			log.Error(err.Error())
			return
		}
		_, _ = bot.SendMessage(
			tu.Messagef(
				chatID,
				strings.Join(args[1:], " "),
			),
		)
		if err != nil {
			return
		}
	}, th.And(
		th.CommandEqual("text"),
		predicate.AdminCommand(),
	))

}
