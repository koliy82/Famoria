package info

import (
	"famoria/internal/database/mongo/repositories/brak"
	"github.com/koliy82/telego"
	tu "github.com/koliy82/telego/telegoutil"
	"go.uber.org/zap"
	"strings"
)

type help struct {
	brakRepo brak.Repository
	log      *zap.Logger
}

func (h help) Handle(bot *telego.Bot, update telego.Update) {
	commands, err := bot.GetMyCommands(&telego.GetMyCommandsParams{})
	if err != nil {
		h.log.Sugar().Error(err)
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
		ReplyMarkup: GenerateButtons(h.brakRepo, update.Message.From.ID),
	})
	if err != nil {
		h.log.Sugar().Error(err)
	}
}
