package admin

import (
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"strings"
)

type sendTextCmd struct {
	log *zap.Logger
}

func (c sendTextCmd) Handle(bot *telego.Bot, update telego.Update) {
	chatID := tu.ID(update.Message.Chat.ID)
	args := strings.Split(update.Message.Text, " ")
	err := bot.DeleteMessage(
		&telego.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: update.Message.MessageID,
		},
	)
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	_, err = bot.SendMessage(
		tu.Messagef(
			chatID,
			strings.Join(args[1:], " "),
		),
	)
	if err != nil {
		c.log.Sugar().Error(err)
		return
	}
}
