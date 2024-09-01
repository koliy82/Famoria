package info

import (
	"famoria/internal/database/mongo/repositories/brak"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"go.uber.org/zap"
	"strings"
)

type helpCmd struct {
	brakRepo brak.Repository
	log      *zap.Logger
}

func (c helpCmd) Handle(bot *telego.Bot, update telego.Update) {
	commands, err := bot.GetMyCommands(&telego.GetMyCommandsParams{})
	if err != nil {
		c.log.Sugar().Error(err)
		return
	}
	text := ""
	for _, command := range commands {
		text += "/" + command.Command + " - " + command.Description + "\n"
	}
	_, err = bot.SendMessage(&telego.SendMessageParams{
		ChatID: tu.ID(update.Message.Chat.ID),
		Text:   strings.TrimSpace(text),
		ReplyParameters: &telego.ReplyParameters{
			MessageID:                update.Message.MessageID,
			ChatID:                   tu.ID(update.Message.Chat.ID),
			AllowSendingWithoutReply: true,
		},
		ReplyMarkup: GenerateButtons(c.brakRepo, update.Message.From.ID),
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
}
